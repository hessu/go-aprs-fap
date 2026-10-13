package fap

import (
	"fmt"
	"strconv"
	"strings"
)

// Mic-E destination field encoding tables

// micEDestDigit maps destination callsign characters to latitude digits and message bits.
type micEDestInfo struct {
	digit   int
	msgBit  int // 0=standard, 1=custom
	isNorth bool
	isWest  bool // only applies to longitude offset check
}

var micEDestTable = map[byte]micEDestInfo{
	'0': {0, 0, false, false},
	'1': {1, 0, false, false},
	'2': {2, 0, false, false},
	'3': {3, 0, false, false},
	'4': {4, 0, false, false},
	'5': {5, 0, false, false},
	'6': {6, 0, false, false},
	'7': {7, 0, false, false},
	'8': {8, 0, false, false},
	'9': {9, 0, false, false},
	'A': {0, 1, false, false},
	'B': {1, 1, false, false},
	'C': {2, 1, false, false},
	'D': {3, 1, false, false},
	'E': {4, 1, false, false},
	'F': {5, 1, false, false},
	'G': {6, 1, false, false},
	'H': {7, 1, false, false},
	'I': {8, 1, false, false},
	'J': {9, 1, false, false},
	'K': {0, 1, false, false},
	'L': {0, 0, false, false},
	'P': {0, 1, true, false},
	'Q': {1, 1, true, false},
	'R': {2, 1, true, false},
	'S': {3, 1, true, false},
	'T': {4, 1, true, false},
	'U': {5, 1, true, false},
	'V': {6, 1, true, false},
	'W': {7, 1, true, false},
	'X': {8, 1, true, false},
	'Y': {9, 1, true, false},
	'Z': {0, 1, true, false},
}

// parseMicE parses a Mic-E encoded packet.
func (p *Packet) parseMicE(opt *options) error {
	p.Type = PacketTypeLocation
	p.Format = FormatMicE

	body := p.Body[1:] // skip type character (` or ')
	dst := p.DstCallsign

	// Strip SSID from destination for Mic-E decoding
	if idx := strings.IndexByte(dst, '-'); idx >= 0 {
		dst = dst[:idx]
	}

	if len(dst) < 6 {
		return p.fail(ErrMiceInvDstCall, "Mic-E destination callsign too short")
	}

	if len(body) < 8 {
		return p.fail(ErrMiceShort, "Mic-E information field too short")
	}

	// Decode latitude from destination callsign
	latDigits := make([]int, 6)
	var msgBuf [3]byte
	isNorth := false
	lonOffset := 0
	isWest := false

	for i := range 6 {
		info, ok := micEDestTable[dst[i]]
		if !ok {
			return p.fail(ErrMiceInvDstCall, fmt.Sprintf("invalid Mic-E destination character: %c", dst[i]))
		}
		latDigits[i] = info.digit

		if i < 3 {
			msgBuf[i] = '0' + byte(info.msgBit)
		}
		if i == 3 && info.isNorth {
			isNorth = true
		}
		if i == 4 {
			// Longitude offset: if the character is P-Z, add 100 degrees
			if dst[i] >= 'P' && dst[i] <= 'Z' {
				lonOffset = 100
			}
		}
		if i == 5 {
			// West indicator
			if dst[i] >= 'P' && dst[i] <= 'Z' {
				isWest = true
			}
		}
	}

	msgBits := string(msgBuf[:])
	p.MBits = msgBits

	// Build latitude
	latDeg := float64(latDigits[0]*10 + latDigits[1])
	latMin := float64(latDigits[2]*10+latDigits[3]) + float64(latDigits[4]*10+latDigits[5])/100.0
	lat := latDeg + latMin/60.0

	if !isNorth {
		lat = -lat
	}

	p.Latitude = &lat
	amb := 0
	p.PosAmbiguity = &amb
	res := posResolution(0)
	p.PosResolution = &res

	// Decode longitude from information field
	lonDeg := int(body[0]) - 28 + lonOffset
	if lonDeg >= 180 && lonDeg <= 189 {
		lonDeg -= 80
	} else if lonDeg >= 190 && lonDeg <= 199 {
		lonDeg -= 190
	}

	lonMin := int(body[1]) - 28
	if lonMin >= 60 {
		lonMin -= 60
	}

	lonHMin := int(body[2]) - 28

	lon := float64(lonDeg) + (float64(lonMin)+float64(lonHMin)/100.0)/60.0

	if isWest {
		lon = -lon
	}

	p.Longitude = &lon

	// Speed and course from bytes 3-5
	sp := int(body[3]) - 28
	dc := int(body[4]) - 28
	se := int(body[5]) - 28

	speed := float64(sp*10+dc/10) - 0 // in knots (raw)
	// The speed encoding: sp contributes tens, dc/10 contributes ones
	speed = float64(sp*10 + dc/10)
	course := (dc%10)*100 + se

	if speed >= 800 {
		speed -= 800
	}
	if course >= 400 {
		course -= 400
	}

	speedKmh := speed * 1.852
	p.Speed = &speedKmh
	p.Course = &course

	// Symbol table and code
	p.SymbolCode = body[6]
	p.SymbolTable = body[7]

	// Validate symbol table
	if p.SymbolTable != '/' && p.SymbolTable != '\\' &&
		!(p.SymbolTable >= 'A' && p.SymbolTable <= 'Z') &&
		!(p.SymbolTable >= '0' && p.SymbolTable <= '9') {
		return p.fail(ErrSymInvTable, fmt.Sprintf("invalid Mic-E symbol table: %c", p.SymbolTable))
	}

	// Rest is comment, possibly with altitude and telemetry
	comment := ""
	if len(body) > 8 {
		comment = body[8:]
	}

	// Check for base-91 comment telemetry |...| first (before altitude)
	comment = p.parseBase91CommentTelemetry(comment)

	// Check for altitude in Mic-E format: XXX} where XXX are 3 base-91 chars
	// followed by '}' as terminator. Altitude in meters, origin at -10000m.
	if idx := strings.IndexByte(comment, '}'); idx >= 3 {
		a1 := comment[idx-3]
		a2 := comment[idx-2]
		a3 := comment[idx-1]
		if a1 >= '!' && a1 <= '{' && a2 >= '!' && a2 <= '{' && a3 >= '!' && a3 <= '{' {
			alt := float64((int(a1)-33)*91*91+(int(a2)-33)*91+(int(a3)-33)) - 10000.0
			p.Altitude = &alt
			comment = comment[:idx-3] + comment[idx+1:]
		}
	}

	// Check for DAO extension in mic-e comment
	comment = p.parseDAO(comment)

	// Check for Mic-E hex telemetry (old format)
	if len(comment) >= 2 && (comment[0] == '\'' || comment[0] == '`') {
		comment = p.parseMicETelemetry(comment)
	}

	p.Comment = comment

	return nil
}

// isHexString checks if all characters in a string are hexadecimal.
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// parseMicETelemetry extracts telemetry data from a Mic-E comment.
func (p *Packet) parseMicETelemetry(comment string) string {
	if len(comment) < 2 {
		return comment
	}

	if comment[0] != '\'' {
		return comment
	}

	rest := comment[1:]

	// Try 5-channel (10 hex chars) first, then 2-channel (4 hex chars)
	var hexLen int
	if len(rest) >= 10 && isHexString(rest[:10]) {
		hexLen = 10
	} else if len(rest) >= 4 && isHexString(rest[:4]) {
		hexLen = 4
	} else {
		return comment
	}

	hexStr := rest[:hexLen]
	vals := make([]*float64, 0)
	if hexLen == 4 {
		// 2-channel telemetry: channels 1 and 3, with channel 2 as zero
		v1, _ := strconv.ParseInt(hexStr[0:2], 16, 64)
		v2, _ := strconv.ParseInt(hexStr[2:4], 16, 64)
		f1, f0, f2 := float64(v1), float64(0), float64(v2)
		vals = append(vals, &f1, &f0, &f2)
	} else {
		for i := 0; i < hexLen; i += 2 {
			v, _ := strconv.ParseInt(hexStr[i:i+2], 16, 64)
			f := float64(v)
			vals = append(vals, &f)
		}
	}

	p.TelemetryData = &Telemetry{
		Vals: vals,
	}

	remaining := rest[hexLen:]
	return strings.TrimLeft(remaining, " ")
}

// MicEMBitsToMessage converts Mic-E message bits to a human-readable message type.
func MicEMBitsToMessage(mbits string) string {
	switch mbits {
	case "111":
		return "Off Duty"
	case "110":
		return "En Route"
	case "101":
		return "In Service"
	case "100":
		return "Returning"
	case "011":
		return "Committed"
	case "010":
		return "Special"
	case "001":
		return "Priority"
	case "000":
		return "Emergency"
	default:
		return "Unknown"
	}
}

// parseMicEMangled attempts to parse a Mic-E packet with a missing speed/course byte.
// aprsd replaces non-printable mic-e bytes with spaces, and some software
// collapses multiple spaces into one, losing a byte. This function detects
// that pattern, inserts the missing space, then parses position and symbol
// normally (but skips speed/course since they are unreliable).
func (p *Packet) parseMicEMangled(opt *options) error {
	body := p.Body[1:]
	dst := p.DstCallsign

	if idx := strings.IndexByte(dst, '-'); idx >= 0 {
		dst = dst[:idx]
	}

	if len(dst) < 6 || len(body) < 7 {
		return p.fail(ErrMiceShort, "mangled Mic-E packet too short")
	}

	// Check for the known corruption pattern:
	// 4 valid position bytes, a single space (collapsed from two), symbol code, symbol table
	if len(body) >= 7 &&
		body[0] >= 0x26 && body[0] <= 0x7f &&
		body[1] >= 0x26 && body[1] <= 0x61 &&
		body[2] >= 0x1c && body[2] <= 0x7f &&
		body[3] >= 0x1c && body[3] <= 0x7f &&
		body[4] == 0x20 &&
		body[5] >= 0x21 && body[5] <= 0x7d &&
		(body[6] == '/' || body[6] == '\\' || (body[6] >= 'A' && body[6] <= 'Z') || (body[6] >= '0' && body[6] <= '9')) {
		// Insert the missing space to restore the 8-byte layout
		body = body[:5] + " " + body[5:]
	} else {
		return p.fail(ErrMiceInvInfoField, "invalid Mic-E information field")
	}

	p.MiceMangled = true
	p.Type = PacketTypeLocation
	p.Format = FormatMicE

	// Decode latitude from destination (same as normal)
	latDigits := make([]int, 6)
	var msgBuf2 [3]byte
	isNorth := false
	lonOffset := 0
	isWest := false

	for i := range 6 {
		info, ok := micEDestTable[dst[i]]
		if !ok {
			return p.fail(ErrMiceInvDstCall, "invalid Mic-E destination character")
		}
		latDigits[i] = info.digit
		if i < 3 {
			msgBuf2[i] = '0' + byte(info.msgBit)
		}
		if i == 3 && info.isNorth {
			isNorth = true
		}
		if i == 4 && dst[i] >= 'P' && dst[i] <= 'Z' {
			lonOffset = 100
		}
		if i == 5 && dst[i] >= 'P' && dst[i] <= 'Z' {
			isWest = true
		}
	}

	p.MBits = string(msgBuf2[:])

	latDeg := float64(latDigits[0]*10 + latDigits[1])
	latMin := float64(latDigits[2]*10+latDigits[3]) + float64(latDigits[4]*10+latDigits[5])/100.0
	lat := latDeg + latMin/60.0
	if !isNorth {
		lat = -lat
	}
	p.Latitude = &lat
	amb := 0
	p.PosAmbiguity = &amb
	res := posResolution(0)
	p.PosResolution = &res

	lonDeg := int(body[0]) - 28 + lonOffset
	if lonDeg >= 180 && lonDeg <= 189 {
		lonDeg -= 80
	} else if lonDeg >= 190 && lonDeg <= 199 {
		lonDeg -= 190
	}

	lonMin := int(body[1]) - 28
	if lonMin >= 60 {
		lonMin -= 60
	}

	lonHMin := int(body[2]) - 28
	lon := float64(lonDeg) + (float64(lonMin)+float64(lonHMin)/100.0)/60.0
	if isWest {
		lon = -lon
	}
	p.Longitude = &lon

	// Skip speed/course — unreliable in mangled packets
	// Symbol code and table from the restored 8-byte layout
	p.SymbolCode = body[6]
	p.SymbolTable = body[7]

	// Comment
	if len(body) > 8 {
		p.Comment = body[8:]
	}

	return nil
}
