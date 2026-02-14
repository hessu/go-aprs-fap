package fap

import (
	"regexp"
	"strconv"
	"strings"
)

// telemetryValueRe matches valid telemetry values (numbers, optionally negative and/or floating point)
var telemetryValueRe = regexp.MustCompile(`^-?[0-9]*\.?[0-9]+$`)

// parseTelemetry parses an APRS telemetry packet.
// Format: T#seq,a1,a2,a3,a4,a5,bbbbbbbb
func (p *Packet) parseTelemetry(opt Options) error {
	p.Type = PacketTypeTelemetry

	body := p.Body[2:] // skip 'T#'

	parts := strings.SplitN(body, ",", 8)
	if len(parts) < 2 {
		return p.fail(ErrTlmInvalid, "telemetry packet has too few fields")
	}

	tlm := &Telemetry{}
	tlm.Seq = parts[0]

	// Parse analog values (up to 5)
	// Values can be: numeric (int or float), empty (undefined), or invalid (treated as end of values)
	vals := make([]*float64, 5)
	for i := 1; i < len(parts) && i <= 5; i++ {
		field := strings.TrimSpace(parts[i])
		if field == "" {
			// Empty field = undefined value
			continue
		}

		// Validate the field format
		if !telemetryValueRe.MatchString(field) {
			// Check for specific invalid patterns
			if field == "-" || strings.HasSuffix(field, ".") {
				return p.fail(ErrTlmInvalid, "invalid telemetry value: "+field)
			}
			// Non-numeric value: stop parsing values here (may be a comment)
			break
		}

		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			break
		}
		vals[i-1] = &val
	}
	tlm.Vals = vals

	// Parse digital bits (field 7, 8-bit binary string)
	if len(parts) >= 7 {
		bitsField := strings.TrimSpace(parts[6])
		if len(bitsField) >= 8 {
			// Validate that the first 8 chars are binary
			bits := bitsField[:8]
			valid := true
			for _, c := range bits {
				if c != '0' && c != '1' {
					valid = false
					break
				}
			}
			if valid {
				tlm.Bits = bits
			}
		}
	}

	p.TelemetryData = tlm

	return nil
}
