package fap

import (
	"fmt"
)

// parseObject parses an APRS object packet.
// Format: ;OBJNAME  *DDMMSS/LATITUDE/LONGITUDEsymbol...
func (p *Packet) parseObject(opt *options) error {
	p.Type = PacketTypeObject

	body := p.Body[1:] // skip ';'

	if len(body) < 31 {
		return p.fail(ErrObjShort, "object packet too short")
	}

	// Object name is 9 characters
	p.ObjectName = body[:9]

	// Alive/killed indicator
	aliveChar := body[9]
	if aliveChar == '*' {
		p.Alive = new(true)
	} else if aliveChar == '_' {
		p.Alive = new(false)
	} else {
		return p.fail(ErrObjInvalid, fmt.Sprintf("invalid object alive/killed indicator: %c", aliveChar))
	}

	// Timestamp (7 characters)
	ts, err := parseTimestamp(body[10:17])
	if err != nil {
		p.warn(ErrTimestampInvalid, fmt.Sprintf("invalid object timestamp: %v", err))
	} else {
		p.Timestamp = ts
	}

	// Position data follows the timestamp
	posBody := body[17:]

	if posBody[0] >= '0' && posBody[0] <= '9' || posBody[0] == ' ' {
		return p.parseUncompressedPosition(posBody, opt)
	}
	return p.parseCompressedPosition(posBody, opt)
}
