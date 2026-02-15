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
		return p.fail(ErrTimestampInvalid, fmt.Sprintf("invalid object timestamp: %v", err))
	}
	p.Timestamp = ts

	// Position data follows the timestamp
	posBody := body[17:]

	if len(posBody) == 0 {
		return p.fail(ErrPosShort, "no position data in object")
	}

	if posBody[0] >= '0' && posBody[0] <= '9' || posBody[0] == ' ' {
		return p.parseUncompressedPosition(posBody, opt)
	}
	return p.parseCompressedPosition(posBody, opt)
}

// parseItem parses an APRS item packet.
// Format: )ITEMNAME!LATITUDE/LONGITUDEsymbol...
func (p *Packet) parseItem(opt *options) error {
	p.Type = PacketTypeItem

	body := p.Body[1:] // skip ')'

	if len(body) < 18 {
		return p.fail(ErrItemShort, "item packet too short")
	}

	// Item name is 3-9 characters, terminated by ! (alive) or _ (killed)
	nameEnd := -1
	for i := 0; i < len(body) && i < 10; i++ {
		if body[i] == '!' || body[i] == '_' {
			nameEnd = i
			break
		}
	}

	if nameEnd < 0 {
		return p.fail(ErrItemInvalid, "item name terminator not found")
	}

	p.ItemName = body[:nameEnd]

	if body[nameEnd] == '!' {
		p.Alive = new(true)
	} else {
		p.Alive = new(false)
	}

	posBody := body[nameEnd+1:]
	if len(posBody) == 0 {
		return p.fail(ErrPosShort, "no position data in item")
	}

	if posBody[0] >= '0' && posBody[0] <= '9' || posBody[0] == ' ' {
		return p.parseUncompressedPosition(posBody, opt)
	}
	return p.parseCompressedPosition(posBody, opt)
}
