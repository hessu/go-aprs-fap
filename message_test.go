package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/51decode-msg.t

var testMessageIDs = []string{"1", "42", "10512", "a", "1Ff84", "F00b4"}

func TestMessageNormal(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{%s", msgid)
			p, err := ParseAPRS(packet)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Destination, "OH7LZB")
			}
			if p.MessageID != msgid {
				t.Errorf("messageid = %q, want %q", p.MessageID, msgid)
			}
			if p.Message != "Testing, 1 2 3" {
				t.Errorf("message = %q, want %q", p.Message, "Testing, 1 2 3")
			}
			if p.MessageAck != "" {
				t.Errorf("messageack = %q, want empty", p.MessageAck)
			}
		})
	}
}

func TestMessageReplyAckEmpty(t *testing.T) {
	// Reply-ack format with no ack ID: {messageid}
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{%s}", msgid)
			p, err := ParseAPRS(packet)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Destination, "OH7LZB")
			}
			if p.MessageID != msgid {
				t.Errorf("messageid = %q, want %q", p.MessageID, msgid)
			}
			if p.MessageAck != "" {
				t.Errorf("messageack = %q, want empty", p.MessageAck)
			}
		})
	}
}

func TestMessageReplyAck(t *testing.T) {
	// Reply-ack with ack ID: {messageid}replyack
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{%s}f001", msgid)
			p, err := ParseAPRS(packet)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Destination, "OH7LZB")
			}
			if p.MessageID != msgid {
				t.Errorf("messageid = %q, want %q", p.MessageID, msgid)
			}
			if p.MessageAck != "f001" {
				t.Errorf("messageack = %q, want %q", p.MessageAck, "f001")
			}
		})
	}
}

func TestMessageAck(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :ack%s", msgid)
			p, err := ParseAPRS(packet)
			if err != nil {
				t.Fatalf("failed to parse ack: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Destination, "OH7LZB")
			}
			if p.MessageAck != msgid {
				t.Errorf("messageack = %q, want %q", p.MessageAck, msgid)
			}
		})
	}
}

func TestMessageReject(t *testing.T) {
	for _, msgid := range testMessageIDs {
		t.Run(fmt.Sprintf("id_%s", msgid), func(t *testing.T) {
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :rej%s", msgid)
			p, err := ParseAPRS(packet)
			if err != nil {
				t.Fatalf("failed to parse reject: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Destination, "OH7LZB")
			}
			if p.MessageRej != msgid {
				t.Errorf("messagerej = %q, want %q", p.MessageRej, msgid)
			}
		})
	}
}
