package fap

import (
	"testing"
)

// Tests ported from perl-aprs-fap/t/40decode-object-inv.t

func TestObjectInvalidBroken(t *testing.T) {
	// Broken object: binary characters destroyed in cut'n'paste,
	// only one space between "HQ" and "*" instead of two,
	// so the alive/killed indicator lands in the wrong position.
	packet := "OH2KKU-1>APRS,TCPIP*,qAC,FIRST:;SRAL HQ *110507zS0%E/Th4_a AKaupinmaenpolku9,open M-Th12-17,F12-14 lcl"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for broken object packet")
	}
	if p.ResultCode != "obj_inv" {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, "obj_inv")
	}
	if p.Type != PacketTypeObject {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeObject)
	}
}
