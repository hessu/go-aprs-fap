package fap

import (
	"strings"
)

// parseMessage parses an APRS message packet.
// Format: :ADDRESSEE:message text{XXXXX
func (p *Packet) parseMessage(opt Options) error {
	p.Type = PacketTypeMessage

	body := p.Body[1:] // skip ':'

	if len(body) < 11 {
		return p.fail(ErrMsgShort, "message packet too short")
	}

	// Addressee is 9 characters, followed by ':'
	if body[9] != ':' {
		return p.fail(ErrMsgInvalid, "message addressee field malformed")
	}

	p.Destination = strings.TrimSpace(body[:9])
	msgBody := body[10:]

	// Check for ack
	if strings.HasPrefix(msgBody, "ack") {
		p.MessageAck = msgBody[3:]
		return nil
	}

	// Check for rej
	if strings.HasPrefix(msgBody, "rej") {
		p.MessageRej = msgBody[3:]
		return nil
	}

	// Look for message ID: {XXXXX
	if idx := strings.LastIndexByte(msgBody, '{'); idx >= 0 {
		p.Message = msgBody[:idx]
		idPart := msgBody[idx+1:]

		// Check for reply-ack: {XXXXX}YY
		if ridx := strings.IndexByte(idPart, '}'); ridx >= 0 {
			p.MessageID = idPart[:ridx]
			p.MessageAck = idPart[ridx+1:]
		} else {
			p.MessageID = idPart
		}
	} else {
		p.Message = msgBody
	}

	return nil
}
