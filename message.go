package fap

import (
	"fmt"
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
		if !isMsgIDChar(id[i]) {
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

	// Like the Perl module, the addressee may only have letters, digits,
	// '_', '-' and spaces, and the message may not have control
	// characters or 0xff.
	for i := 0; i < 9; i++ {
		c := body[i]
		if !isMsgIDChar(c) && c != '_' && c != '-' && c != ' ' {
			return p.fail(ErrMsgInvalid, fmt.Sprintf("invalid character in message addressee: 0x%02x", c))
		}
	}
	msgBody := body[10:]
	for i := 0; i < len(msgBody); i++ {
		if c := msgBody[i]; c < 0x20 || c == 0x7f || c == 0xff {
			return p.fail(ErrMsgInvalid, fmt.Sprintf("invalid character in message: 0x%02x", c))
		}
	}

	msg := &Message{}
	p.Message = msg

	msg.Destination = strings.TrimRight(body[:9], " ")

	// Check for ack and rej. Trailing spaces are allowed because some
	// broken software inserts them. Otherwise "ack" is message text.
	if id, ok := ackRejID(msgBody, "ack"); ok {
		msg.AckID = id
		return nil
	}
	if id, ok := ackRejID(msgBody, "rej"); ok {
		msg.RejID = id
		return nil
	}

	// Separate the message ID {XXXXX, and the reply-ack {XXXXX}YY, from
	// the text, if they are valid; otherwise it is all message text.
	msg.Text = msgBody
	if text, id, ack, ok := splitMessageID(msgBody); ok {
		msg.Text = text
		msg.ID = id
		msg.AckID = ack
	}

	// Catch telemetry parameter messages (PARM, UNIT, EQNS, BITS)
	if isTelemetryMessage(msg.Text) {
		p.Type = PacketTypeTelemetryMessage
	}

	return nil
}

// trimTrailingSpace removes trailing white space, like \s+$ in Perl.
func trimTrailingSpace(s string) string {
	return strings.TrimRight(s, " \t\r\n\f\v")
}

// isMsgIDChar tells whether c can be in a message ID.
func isMsgIDChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// ackRejID returns the message ID of an ack or a reject, if the message
// text is the prefix ("ack" or "rej") followed by 1-5 ID characters (or
// '}', which some software includes) and optional trailing white space.
func ackRejID(text, prefix string) (string, bool) {
	if !strings.HasPrefix(text, prefix) {
		return "", false
	}
	id := trimTrailingSpace(text[len(prefix):])
	if len(id) < 1 || len(id) > 5 {
		return "", false
	}
	for i := 0; i < len(id); i++ {
		if !isMsgIDChar(id[i]) && id[i] != '}' {
			return "", false
		}
	}
	return id, true
}

// splitMessageID splits a message into the text, the message ID and the
// reply-ack ID, if it ends with a valid message ID: '{' and 1-5 ID
// characters, optionally followed by '}' and a reply-ack ID of 1-5 ID
// characters, and trailing white space. The text must not contain a '{'.
// This is the Perl module's /^([^{]*)\{([A-Za-z0-9]{1,5})(}[A-Za-z0-9]{1,5}|\}|)\s*$/.
func splitMessageID(body string) (text, id, ack string, ok bool) {
	idx := strings.IndexByte(body, '{')
	if idx < 0 {
		return "", "", "", false
	}
	rest := trimTrailingSpace(body[idx+1:])

	n := 0
	for n < len(rest) && n < 6 && isMsgIDChar(rest[n]) {
		n++
	}
	if n < 1 || n > 5 {
		return "", "", "", false
	}
	id, rest = rest[:n], rest[n:]

	if rest != "" {
		if rest[0] != '}' {
			return "", "", "", false
		}
		ack = rest[1:]
		if len(ack) > 5 {
			return "", "", "", false
		}
		for i := 0; i < len(ack); i++ {
			if !isMsgIDChar(ack[i]) {
				return "", "", "", false
			}
		}
	}

	return body[:idx], id, ack, true
}

// isTelemetryMessage checks if a message body starts with a telemetry
// parameter keyword followed by a dot (case-insensitive).
func isTelemetryMessage(text string) bool {
	if len(text) < 5 || text[4] != '.' {
		return false
	}
	prefix := text[:4]
	return strings.EqualFold(prefix, "PARM") ||
		strings.EqualFold(prefix, "UNIT") ||
		strings.EqualFold(prefix, "EQNS") ||
		strings.EqualFold(prefix, "BITS")
}
