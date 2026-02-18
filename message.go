package fap

import (
	"strings"
)

// EncodeMessage encodes a Message struct into an APRS message body string.
// The returned string is suitable for use as Packet.Body and can be decoded
// by parseMessage.
func EncodeMessage(msg *Message) (string, error) {
	if msg.Destination == "" {
		return "", &ParseError{Code: ErrMsgNoDst.Code, Msg: "message destination is required"}
	}
	if len(msg.Destination) > 9 {
		return "", &ParseError{Code: ErrMsgDstTooLong.Code, Msg: "message destination too long (max 9 characters)"}
	}

	if msg.AckID != "" && msg.RejID != "" {
		return "", &ParseError{Code: ErrMsgAckRej.Code, Msg: "message cannot have both ack and reject"}
	}
	if msg.Text != "" && msg.RejID != "" {
		return "", &ParseError{Code: ErrMsgAckRej.Code, Msg: "message cannot have both content and reject"}
	}

	if containsCRLF(msg.Destination) || containsCRLF(msg.Text) || containsCRLF(msg.ID) || containsCRLF(msg.AckID) || containsCRLF(msg.RejID) {
		return "", &ParseError{Code: ErrMsgCRLF.Code, Msg: "message fields must not contain CR or LF"}
	}

	if msg.ID != "" && !isValidMsgID(msg.ID) {
		return "", &ParseError{Code: ErrMsgIDInvalid.Code, Msg: "message ID must be 1-5 alphanumeric characters"}
	}

	// Pad addressee to 9 characters
	addressee := msg.Destination + strings.Repeat(" ", 9-len(msg.Destination))

	// Build message content after the addressee field
	var content string
	switch {
	case msg.AckID != "" && msg.ID == "":
		// Pure ack
		content = "ack" + msg.AckID
	case msg.RejID != "":
		// Reject
		content = "rej" + msg.RejID
	default:
		content = msg.Text
		if msg.ID != "" {
			if msg.AckID != "" {
				// Reply-ack: {ID}ackID — total of ID + '}' + AckID must fit in 5 chars
				if len(msg.ID)+1+len(msg.AckID) > 5 {
					return "", &ParseError{Code: ErrMsgReplyAck.Code, Msg: "reply-ack too long to embed, send ack separately"}
				}
				content += "{" + msg.ID + "}" + msg.AckID
			} else {
				content += "{" + msg.ID
			}
		}
	}

	return ":" + addressee + ":" + content, nil
}

// containsCRLF returns true if s contains a CR or LF character.
func containsCRLF(s string) bool {
	return strings.ContainsAny(s, "\r\n")
}

// isValidMsgID checks that a message ID is 1-5 alphanumeric characters.
func isValidMsgID(id string) bool {
	if len(id) < 1 || len(id) > 5 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// parseMessage parses an APRS message packet.
// Format: :ADDRESSEE:message text{XXXXX
func (p *Packet) parseMessage(opt *options) error {
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
