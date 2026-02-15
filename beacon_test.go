package fap

import (
	"testing"
)

// Tests ported from perl-aprs-fap/t/52decode-beacon.t

func TestBeaconNonAPRS(t *testing.T) {
	// Non-APRS beacon packet — body starts with space, not a recognized type.
	// Should fail to parse but still populate header fields.
	packet := "OH2RDU>UIDIGI: UIDIGI 1.9"

	p, err := Parse(packet, nil)
	if err == nil {
		t.Fatal("expected error for non-APRS beacon packet")
	}
	if p.ResultCode != ErrTypeNotSupported {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, ErrTypeNotSupported)
	}
	if p.SrcCallsign != "OH2RDU" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2RDU")
	}
	if p.DstCallsign != "UIDIGI" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "UIDIGI")
	}
	if p.Body != " UIDIGI 1.9" {
		t.Errorf("body = %q, want %q", p.Body, " UIDIGI 1.9")
	}
}
