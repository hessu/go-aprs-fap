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

	// Object name is 9 printable characters
	for i := 0; i < 9; i++ {
		if body[i] < 0x20 || body[i] > 0x7e {
			return p.fail(ErrObjInvalid, fmt.Sprintf("invalid character in object name: 0x%02x", body[i]))
		}
	}

	// Alive/killed indicator
	aliveChar := body[9]
	if aliveChar != '*' && aliveChar != '_' {
		return p.fail(ErrObjInvalid, fmt.Sprintf("invalid object alive/killed indicator: %c", aliveChar))
	}

	// Timestamp (7 characters): the object is invalid unless it has the
	// format of a timestamp, but an invalid time is only a warning, as
	// in the Perl module.
	if !isTimestampFormat(body[10:17]) {
		return p.fail(ErrObjInvalid, fmt.Sprintf("invalid object timestamp: %q", body[10:17]))
	}

	p.ObjectName = body[:9]
	p.Alive = new(aliveChar == '*')

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
