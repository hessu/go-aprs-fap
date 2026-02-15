package fap

import (
	"strings"
)

// parseStatus parses an APRS status report.
// Format: >DDHHMMzstatus text  or  >status text
func (p *Packet) parseStatus(opt *options) error {
	p.Type = PacketTypeStatus

	body := p.Body[1:] // skip '>'

	// Check if body starts with a timestamp (7 chars ending in z, h, or /)
	if len(body) >= 7 {
		indicator := body[6]
		if indicator == 'z' || indicator == '/' {
			// Timestamp: DDHHMMz or DDHHMM/
			ts, err := parseTimestamp(body[:7])
			if err == nil {
				p.Timestamp = ts
				body = body[7:]
			}
		}
	}

	p.Status = body
	return nil
}

// parseCapabilities parses an APRS station capabilities packet.
// Format: <cap1=val1,cap2=val2,...
func (p *Packet) parseCapabilities(opt *options) error {
	p.Type = PacketTypeCapabilities
	p.Capabilities = make(map[string]string)

	body := p.Body[1:] // skip '<'
	parts := strings.Split(body, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		if idx := strings.IndexByte(part, '='); idx >= 0 {
			p.Capabilities[part[:idx]] = part[idx+1:]
		} else {
			p.Capabilities[part] = ""
		}
	}

	return nil
}
