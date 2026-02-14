package fap

import (
	"strings"
)

// parseStatus parses an APRS status report.
// Format: >status text
func (p *Packet) parseStatus(opt Options) error {
	p.Type = PacketTypeStatus
	p.Status = strings.TrimSpace(p.Body[1:])
	return nil
}

// parseCapabilities parses an APRS station capabilities packet.
// Format: <cap1=val1,cap2=val2,...
func (p *Packet) parseCapabilities(opt Options) error {
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
