package fap

import "strings"

// Base-91 comment telemetry, which may appear in the comment of any of the
// three position formats (uncompressed, compressed and Mic-E), delimited by
// '|' characters: |ssaabbccddeeff| where ss is the sequence number, aa to
// ee up to five analog channels and ff the binary bits, each a two-character
// base-91 integer. The sequence number and the first channel are mandatory,
// so the content between the pipes is 4 to 14 characters.
// Specification: aprs-base91-comment-telemetry.txt in hessu/aprs-specs.
// Port of _comment_telemetry() in the Perl Ham::APRS::FAP module.

const (
	base91TelemetryMinLen = 4  // sequence + 1 channel
	base91TelemetryMaxLen = 14 // sequence + 5 channels + bits
)

// isBase91TelemetryChar checks if a character is valid in base-91 telemetry (0x21-0x7B).
func isBase91TelemetryChar(c byte) bool {
	return c >= '!' && c <= '{'
}

// isBase91TelemetryContent checks whether s is the content of a base-91
// telemetry extension: an even length of 4 to 14 valid characters.
func isBase91TelemetryContent(s string) bool {
	if len(s) < base91TelemetryMinLen || len(s) > base91TelemetryMaxLen || len(s)%2 != 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isBase91TelemetryChar(s[i]) {
			return false
		}
	}
	return true
}

// base91Pair decodes a two-character base-91 integer.
func base91Pair(s string) int {
	return (int(s[0])-33)*91 + (int(s[1]) - 33)
}

// parseBase91CommentTelemetry extracts base-91 comment telemetry from a
// position comment into p.TelemetryData, and returns the comment with the
// telemetry removed. The last valid |...| in the comment is used, matching
// the greedy (.*) before the first \| in the Perl regex, so that a '|' in
// the free-form comment text does not swallow the telemetry.
func (p *Packet) parseBase91CommentTelemetry(comment string) string {
	// Search from the end for the closing |, then find the matching opening |
	// with valid base-91 content between them.
	bestStart := -1
	bestEnd := -1

	for end := len(comment) - 1; end >= 0; end-- {
		if comment[end] != '|' {
			continue
		}
		// Try to find an opening | before this one with valid content
		for start := end - 1; start >= 0; start-- {
			if comment[start] != '|' {
				continue
			}
			if !isBase91TelemetryContent(comment[start+1 : end]) {
				continue
			}
			// Found a valid match - use the one with the latest start (greedy)
			if start > bestStart {
				bestStart = start
				bestEnd = end
			}
			break // only need the first valid opening | for this closing |
		}
		if bestStart >= 0 {
			break // use the last (rightmost) closing | that has a valid match
		}
	}

	if bestStart < 0 {
		return comment
	}

	tlmData := comment[bestStart+1 : bestEnd]
	pairs := len(tlmData) / 2

	// First pair is sequence number
	tlm := &Telemetry{
		Seq: base91Pair(tlmData[0:2]),
	}

	// Remaining pairs are values (up to 5)
	vals := make([]*float64, 5)
	for i := 1; i < pairs && i <= 5; i++ {
		val := float64(base91Pair(tlmData[i*2 : i*2+2]))
		vals[i-1] = &val
	}

	// If we have 7 pairs, the last one is the binary bits
	// Perl uses unpack('b8', ...) which is LSB-first bit order
	if pairs >= 7 {
		bitsVal := base91Pair(tlmData[12:14])
		var bits [8]byte
		for b := range 8 {
			if bitsVal&(1<<uint(b)) != 0 {
				bits[b] = '1'
			} else {
				bits[b] = '0'
			}
		}
		tlm.Bits = string(bits[:])
	}

	tlm.Vals = vals
	p.TelemetryData = tlm

	// Remove the telemetry from the comment
	return strings.TrimSpace(comment[:bestStart] + comment[bestEnd+1:])
}
