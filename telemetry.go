package fap

import (
	"strconv"
	"strings"
)

// parseTelemetry parses an APRS telemetry packet.
// Format: T#seq,a1,a2,a3,a4,a5,bbbbbbbb
func (p *Packet) parseTelemetry(opt Options) error {
	p.Type = PacketTypeTelemetry

	body := p.Body[2:] // skip 'T#'

	parts := strings.SplitN(body, ",", 7)
	if len(parts) < 2 {
		return p.fail(ErrTlmInvalid, "telemetry packet has too few fields")
	}

	tlm := &Telemetry{}
	tlm.Seq = parts[0]

	// Parse analog values (up to 5)
	for i := 1; i < len(parts) && i <= 5; i++ {
		val, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			continue
		}
		tlm.Vals = append(tlm.Vals, val)
	}

	// Parse digital bits (field 7, 8-bit binary string)
	if len(parts) >= 7 {
		bits := strings.TrimSpace(parts[6])
		if len(bits) >= 8 {
			tlm.Bits = bits[:8]
		}
	}

	p.TelemetryData = tlm

	return nil
}
