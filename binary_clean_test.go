package fap

import (
	"fmt"
	"strings"
	"testing"
)

// buildBinaryMessagePacket constructs an APRS message packet with the given content.
// Characters {, ~, | are stripped from the message content as they have
// special meaning in APRS messages.
func buildBinaryMessagePacket(content string) (string, string) {
	srccall := "OH7AA-1"
	dstcall := "APRS"
	destination := "OH7LZB   "
	messageid := "42"

	// Strip characters not allowed in a message
	message := strings.NewReplacer("{", "", "~", "", "|", "").Replace(content)

	packet := fmt.Sprintf("%s>%s,WIDE1-1,WIDE2-2,qAo,OH7AA::%s:%s{%s",
		srccall, dstcall, destination, message, messageid)
	return packet, message
}

func testBinaryMessage(t *testing.T, setName string, content string) {
	t.Helper()

	packet, message := buildBinaryMessagePacket(content)

	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("%s: failed to parse a message packet: %v", setName, err)
	}
	if p.Type != PacketTypeMessage {
		t.Errorf("%s: wrong packet type: got %q, want %q", setName, p.Type, PacketTypeMessage)
	}
	if p.Destination != "OH7LZB" {
		t.Errorf("%s: wrong message dst callsign: got %q, want %q", setName, p.Destination, "OH7LZB")
	}
	if p.MessageID != "42" {
		t.Errorf("%s: wrong message id: got %q, want %q", setName, p.MessageID, "42")
	}
	if p.Message != message {
		t.Errorf("%s: wrong message: got %q, want %q", setName, p.Message, message)
	}
}

func TestBinaryCleanASCII(t *testing.T) {
	var b strings.Builder
	for i := 32; i <= 126; i++ {
		b.WriteByte(byte(i))
	}
	testBinaryMessage(t, "msg set ascii", b.String())
}

func TestBinaryCleanSet1(t *testing.T) {
	var b strings.Builder
	for i := 32; i < 32+67; i++ {
		b.WriteByte(byte(i))
	}
	testBinaryMessage(t, "msg set 1", b.String())
}

func TestBinaryCleanSet2(t *testing.T) {
	var b strings.Builder
	for i := 32 + 67; i < 32+67+67; i++ {
		if i != 127 {
			b.WriteByte(byte(i))
		}
	}
	testBinaryMessage(t, "msg set 2", b.String())
}

func TestBinaryCleanSet3(t *testing.T) {
	var b strings.Builder
	for i := 32 + 67 + 67; i < 32+67+67+67; i++ {
		b.WriteByte(byte(i))
	}
	testBinaryMessage(t, "msg set 3", b.String())
}

func TestBinaryCleanSet4(t *testing.T) {
	var b strings.Builder
	for i := 32 + 67 + 67 + 67; i < 255; i++ {
		b.WriteByte(byte(i))
	}
	testBinaryMessage(t, "msg set 4", b.String())
}

func TestBinaryCleanIndividualChars(t *testing.T) {
	for i := 32; i < 255; i++ {
		if i == 127 {
			continue
		}
		ch := string([]byte{byte(i)})
		content := "char: " + ch
		setName := fmt.Sprintf("binary char %d", i)
		testBinaryMessage(t, setName, content)
	}
}
