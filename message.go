package fap

import (
	"strings"
)

// parseMessage parses an APRS message packet.
// Format: :ADDRESSEE:message text{XXXXX
func (p *Packet) parseMessage(opt *Options) error {
	p.Type = PacketTypeMessage

	body := p.Body[1:] // skip ':'

	if len(body) < 11 {
		return p.fail(ErrMsgShort, "message packet too short")
	}

	// Addressee is 9 characters, followed by ':'
	if body[9] != ':' {
		return p.fail(ErrMsgInvalid, "message addressee field malformed")
	}

	msg := &Message{}
	p.Message = msg

	msg.Destination = strings.TrimSpace(body[:9])
	msgBody := body[10:]

	// Check for ack
	if strings.HasPrefix(msgBody, "ack") {
		msg.AckID = msgBody[3:]
		return nil
	}

	// Check for rej
	if strings.HasPrefix(msgBody, "rej") {
		msg.RejID = msgBody[3:]
		return nil
	}

	// Look for message ID: {XXXXX
	if idx := strings.LastIndexByte(msgBody, '{'); idx >= 0 {
		msg.Text = msgBody[:idx]
		idPart := msgBody[idx+1:]

		// Check for reply-ack: {XXXXX}YY
		if ridx := strings.IndexByte(idPart, '}'); ridx >= 0 {
			msg.ID = idPart[:ridx]
			msg.AckID = idPart[ridx+1:]
		} else {
			msg.ID = idPart
		}
	} else {
		msg.Text = msgBody
	}

	return nil
}
