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
			p, err := Parse(packet, nil)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "OH7LZB")
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
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{%s}", msgid)
			p, err := Parse(packet, nil)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "OH7LZB")
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
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{%s}f001", msgid)
			p, err := Parse(packet, nil)
			if err != nil {
				t.Fatalf("failed to parse message: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "OH7LZB")
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
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :ack%s", msgid)
			p, err := Parse(packet, nil)
			if err != nil {
				t.Fatalf("failed to parse ack: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "OH7LZB")
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
			packet := fmt.Sprintf("OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :rej%s", msgid)
			p, err := Parse(packet, nil)
			if err != nil {
				t.Fatalf("failed to parse reject: %v", err)
			}
			if p.ResultCode != "" {
				t.Fatalf("unexpected result code: %s", p.ResultCode)
			}
			if p.Type != PacketTypeMessage {
				t.Errorf("type = %q, want %q", p.Type, PacketTypeMessage)
			}
			if p.Message == nil {
				t.Fatalf("message is nil")
			}
			if p.Message.Destination != "OH7LZB" {
				t.Errorf("destination = %q, want %q", p.Message.Destination, "OH7LZB")
			}
			if p.Message.RejID != msgid {
				t.Errorf("messagerej = %q, want %q", p.Message.RejID, msgid)
			}
		})
	}
}
