package fap

import (
	"errors"
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/51decode-msg.t

var testMessageIDs = []string{"1", "42", "10512", "a", "1Ff84", "F00b4"}

func TestMessageNormal(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :Testing, 1 2 3{%s", msgid)
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.ID != msgid {
				t.Errorf("messageid = %q, want %q", p.Message.ID, msgid)
			}
			if p.Message.Text != "Testing, 1 2 3" {
				t.Errorf("message = %q, want %q", p.Message.Text, "Testing, 1 2 3")
			}
			if p.Message.AckID != "" {
				t.Errorf("messageack = %q, want empty", p.Message.AckID)
			}
		})
	}
}

func TestMessageReplyAckEmpty(t *testing.T) {
	// Reply-ack format with no ack ID: {messageid}
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :Testing, 1 2 3{%s}", msgid)
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.ID != msgid {
				t.Errorf("messageid = %q, want %q", p.Message.ID, msgid)
			}
			if p.Message.AckID != "" {
				t.Errorf("messageack = %q, want empty", p.Message.AckID)
			}
		})
	}
}

func TestMessageReplyAck(t *testing.T) {
	// Reply-ack with ack ID: {messageid}replyack
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :Testing, 1 2 3{%s}f001", msgid)
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.ID != msgid {
				t.Errorf("messageid = %q, want %q", p.Message.ID, msgid)
			}
			if p.Message.AckID != "f001" {
				t.Errorf("messageack = %q, want %q", p.Message.AckID, "f001")
			}
		})
	}
}

func TestMessageAck(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :ack%s", msgid)
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse ack: %v", err)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.AckID != msgid {
				t.Errorf("messageack = %q, want %q", p.Message.AckID, msgid)
			}
		})
	}
}

func TestMessageReject(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :rej%s", msgid)
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse reject: %v", err)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.RejID != msgid {
				t.Errorf("messagerej = %q, want %q", p.Message.RejID, msgid)
			}
		})
	}
}

func TestMessageNoID(t *testing.T) {
	// Message without a message ID (no '{' in the message text)
	packet := "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :Hello world"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse message without ID: %v", err)
	}
	if p.Type != PacketTypeMessage {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
	}
	if p.Message == nil {
		t.Fatalf("message is nil")
	}
	if p.Message.Destination != "N0CALL" {
		t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
	}
	if p.Message.Text != "Hello world" {
		t.Errorf("message = %q, want %q", p.Message.Text, "Hello world")
	}
	if p.Message.ID != "" {
		t.Errorf("messageid = %q, want empty", p.Message.ID)
	}
	if p.Message.AckID != "" {
		t.Errorf("messageack = %q, want empty", p.Message.AckID)
	}
	if p.Message.RejID != "" {
		t.Errorf("messagerej = %q, want empty", p.Message.RejID)
	}
}

func TestEncodeMessageNormal(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			msg := &Message{
				Destination: "N0CALL",
				Text:        "Testing, 1 2 3",
				ID:          msgid,
			}
			body, err := EncodeMessage(msg)
			if err != nil {
				t.Fatalf("EncodeMessage failed: %v", err)
			}

			// Verify round-trip: parse the encoded body
			packet := fmt.Sprintf("OH7AA-1>APRS::%s", body[1:]) // body starts with ':', header already has ':'
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse encoded message: %v", err)
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.Text != "Testing, 1 2 3" {
				t.Errorf("text = %q, want %q", p.Message.Text, "Testing, 1 2 3")
			}
			if p.Message.ID != msgid {
				t.Errorf("id = %q, want %q", p.Message.ID, msgid)
			}
			if p.Message.AckID != "" {
				t.Errorf("ackid = %q, want empty", p.Message.AckID)
			}
		})
	}
}

func TestEncodeMessageAck(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			msg := &Message{
				Destination: "N0CALL",
				AckID:       msgid,
			}
			body, err := EncodeMessage(msg)
			if err != nil {
				t.Fatalf("EncodeMessage failed: %v", err)
			}

			packet := fmt.Sprintf("OH7AA-1>APRS::%s", body[1:])
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse encoded ack: %v", err)
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.AckID != msgid {
				t.Errorf("ackid = %q, want %q", p.Message.AckID, msgid)
			}
		})
	}
}

func TestEncodeMessageReject(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			msg := &Message{
				Destination: "N0CALL",
				RejID:       msgid,
			}
			body, err := EncodeMessage(msg)
			if err != nil {
				t.Fatalf("EncodeMessage failed: %v", err)
			}

			packet := fmt.Sprintf("OH7AA-1>APRS::%s", body[1:])
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse encoded reject: %v", err)
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.RejID != msgid {
				t.Errorf("rejid = %q, want %q", p.Message.RejID, msgid)
			}
		})
	}
}

func TestEncodeMessageReplyAck(t *testing.T) {
	tests := []struct {
		id    string
		ackID string
	}{
		{"1", "abc"},
		{"42", "ab"},
		{"ab", "cd"},
		{"abc", "d"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("id_%s_ack_%s", tc.id, tc.ackID), func(t *testing.T) {
			msg := &Message{
				Destination: "N0CALL",
				Text:        "Testing, 1 2 3",
				ID:          tc.id,
				AckID:       tc.ackID,
			}
			body, err := EncodeMessage(msg)
			if err != nil {
				t.Fatalf("EncodeMessage failed: %v", err)
			}

			packet := fmt.Sprintf("OH7AA-1>APRS::%s", body[1:])
			p, err := Parse(packet)
			if err != nil {
				t.Fatalf("failed to parse encoded reply-ack: %v", err)
			}
			if p.Message.Destination != "N0CALL" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
			}
			if p.Message.Text != "Testing, 1 2 3" {
				t.Errorf("text = %q, want %q", p.Message.Text, "Testing, 1 2 3")
			}
			if p.Message.ID != tc.id {
				t.Errorf("id = %q, want %q", p.Message.ID, tc.id)
			}
			if p.Message.AckID != tc.ackID {
				t.Errorf("ackid = %q, want %q", p.Message.AckID, tc.ackID)
			}
		})
	}
}

func TestEncodeMessageNoID(t *testing.T) {
	msg := &Message{
		Destination: "N0CALL",
		Text:        "Hello world",
	}
	body, err := EncodeMessage(msg)
	if err != nil {
		t.Fatalf("EncodeMessage failed: %v", err)
	}

	packet := fmt.Sprintf("OH7AA-1>APRS::%s", body[1:])
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse encoded message: %v", err)
	}
	if p.Message.Destination != "N0CALL" {
		t.Errorf("destination = %q, want %q", p.Message.Destination, "N0CALL")
	}
	if p.Message.Text != "Hello world" {
		t.Errorf("text = %q, want %q", p.Message.Text, "Hello world")
	}
	if p.Message.ID != "" {
		t.Errorf("id = %q, want empty", p.Message.ID)
	}
}

func TestEncodeMessageErrors(t *testing.T) {
	tests := []struct {
		name  string
		msg   *Message
		errIs error
	}{
		{
			name:  "empty destination",
			msg:   &Message{Text: "hello"},
			errIs: ErrMsgNoDst,
		},
		{
			name:  "destination too long",
			msg:   &Message{Destination: "0123456789", Text: "hello"},
			errIs: ErrMsgDstTooLong,
		},
		{
			name:  "message ID too long",
			msg:   &Message{Destination: "N0CALL", Text: "hello", ID: "123456"},
			errIs: ErrMsgIDInvalid,
		},
		{
			name:  "message ID with space",
			msg:   &Message{Destination: "N0CALL", Text: "hello", ID: "1 2"},
			errIs: ErrMsgIDInvalid,
		},
		{
			name:  "message ID with special char",
			msg:   &Message{Destination: "N0CALL", Text: "hello", ID: "ab!c"},
			errIs: ErrMsgIDInvalid,
		},
		{
			name:  "reply-ack too long",
			msg:   &Message{Destination: "N0CALL", Text: "hello", ID: "abc", AckID: "de"},
			errIs: ErrMsgReplyAck,
		},
		{
			name:  "reply-ack just over limit",
			msg:   &Message{Destination: "N0CALL", Text: "hello", ID: "ab", AckID: "cde"},
			errIs: ErrMsgReplyAck,
		},
		{
			name:  "both ack and reject",
			msg:   &Message{Destination: "N0CALL", AckID: "1", RejID: "2"},
			errIs: ErrMsgAckRej,
		},
		{
			name:  "content with reject",
			msg:   &Message{Destination: "N0CALL", Text: "hello", RejID: "1"},
			errIs: ErrMsgAckRej,
		},
		{
			name:  "LF in text",
			msg:   &Message{Destination: "N0CALL", Text: "hello\nworld"},
			errIs: ErrMsgCRLF,
		},
		{
			name:  "CR in text",
			msg:   &Message{Destination: "N0CALL", Text: "hello\rworld"},
			errIs: ErrMsgCRLF,
		},
		{
			name:  "LF in destination",
			msg:   &Message{Destination: "N0\nCALL", Text: "hello"},
			errIs: ErrMsgCRLF,
		},
		{
			name:  "CR in ack ID",
			msg:   &Message{Destination: "N0CALL", AckID: "1\r"},
			errIs: ErrMsgCRLF,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := EncodeMessage(tc.msg)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.errIs)
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
		})
	}
}

func TestMessageErrors(t *testing.T) {
	tests := []struct {
		name   string
		packet string
		errIs  error
	}{
		{
			name:   "too short",
			packet: "OH7AA-1>APRS::N0CALL  ",
			errIs:  ErrMsgShort,
		},
		{
			name:   "missing colon after addressee",
			packet: "OH7AA-1>APRS::N0CALL  XHello world",
			errIs:  ErrMsgInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.packet)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.errIs)
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
		})
	}
}

// TestMessageIDFormat tests that the message ID, ack and reject IDs are
// only separated from the message text if they are valid: 1-5 letters
// or digits, after the only '{' of the message. Otherwise it is all
// message text. The first two packets are from APRS-IS. The expected
// values are from the Perl module.
func TestMessageIDFormat(t *testing.T) {
	tests := []struct {
		name   string
		packet string
		dst    string
		text   string
		id     string
		ack    string
		rej    string
	}{
		{
			name:   "braces in text",
			packet: "TEST-13>APRS,TCPIP*,qAC,T2QUEBEC::TEST-13  :IPINFO={'pub'='192.0.2.8','priv'='192.168.16.39/24'}",
			dst:    "TEST-13",
			text:   "IPINFO={'pub'='192.0.2.8','priv'='192.168.16.39/24'}",
		},
		{
			name:   "binary text",
			packet: "TEST-1>APCLUB,WIDE1-1,WIDE2-1,qAR,ON4MAX-2::CLUBCHAT :[MLY4 5/6] EiVjx0{JPD2+}&j>^7~c35",
			dst:    "CLUBCHAT",
			text:   "[MLY4 5/6] EiVjx0{JPD2+}&j>^7~c35",
		},
		{
			name:   "ack without id",
			packet: "TEST>APZ500,WIDE3-1,qAR,KE4PMP-1::N0CALL   :ack",
			dst:    "N0CALL",
			text:   "ack",
		},
		{
			name:   "ack with trailing space",
			packet: "TEST>APRS::N0CALL   :ack12 ",
			dst:    "N0CALL",
			ack:    "12",
		},
		{
			name:   "rej with brace",
			packet: "TEST>APRS::N0CALL   :rej1}",
			dst:    "N0CALL",
			rej:    "1}",
		},
		{
			name:   "reply-ack",
			packet: "TEST>APRS::N0CALL   :Hello{AB}CD",
			dst:    "N0CALL",
			text:   "Hello",
			id:     "AB",
			ack:    "CD",
		},
		{
			name:   "id too long",
			packet: "TEST>APRS::N0CALL   :Hello{ABCDEF",
			dst:    "N0CALL",
			text:   "Hello{ABCDEF",
		},
		{
			name:   "two braces",
			packet: "TEST>APRS::N0CALL   :Hi {there} {1",
			dst:    "N0CALL",
			text:   "Hi {there} {1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			m := p.Message
			if m == nil {
				t.Fatal("message is nil")
			}
			if m.Destination != tc.dst {
				t.Errorf("destination = %q, want %q", m.Destination, tc.dst)
			}
			if m.Text != tc.text {
				t.Errorf("message = %q, want %q", m.Text, tc.text)
			}
			if m.ID != tc.id {
				t.Errorf("messageid = %q, want %q", m.ID, tc.id)
			}
			if m.AckID != tc.ack {
				t.Errorf("messageack = %q, want %q", m.AckID, tc.ack)
			}
			if m.RejID != tc.rej {
				t.Errorf("messagerej = %q, want %q", m.RejID, tc.rej)
			}
		})
	}
}
