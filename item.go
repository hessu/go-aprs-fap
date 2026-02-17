package fap

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

	if posBody[0] >= '0' && posBody[0] <= '9' || posBody[0] == ' ' {
		return p.parseUncompressedPosition(posBody, opt)
	}
	return p.parseCompressedPosition(posBody, opt)
}
