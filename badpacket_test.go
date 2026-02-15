package fap

import (
	"testing"
)

// Tests ported from perl-aprs-fap/t/10badpacket.t

func TestBadPacketCorruptedPosition(t *testing.T) {
	// Corrupted uncompressed position with invalid characters in lat/lon
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!60ff.51N/0250akh3r99hfae"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for corrupted position packet")
	}
	if p.ResultCode != "loc_inv" {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, "loc_inv")
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.SrcCallsign != "OH2RDP-1" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2RDP-1")
	}
	if p.DstCallsign != "BEACON-15" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "BEACON-15")
	}
	if p.Latitude != nil {
		t.Errorf("latitude = %v, want nil", *p.Latitude)
	}
	if p.Longitude != nil {
		t.Errorf("longitude = %v, want nil", *p.Longitude)
	}
}

func TestBadPacketBadSrcCall(t *testing.T) {
	// Bad source callsign (contains underscore)
	packet := "K6IFR_S>APJS10,TCPIP*,qAC,K6IFR-BS:;K6IFR B *250300z3351.79ND11626.40WaRNG0040 440 Voice 447.140 -5.00 Mhz"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for bad source callsign")
	}
	if p.ResultCode != "srccall_badchars" {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, "srccall_badchars")
	}
	if p.Type != "" {
		t.Errorf("type = %q, want empty", p.Type)
	}
}

func TestBadPacketBadDigiCall(t *testing.T) {
	// Bad digipeater callsign (contains underscore)
	packet := "SV2BRF-6>APU25N,TCPXX*,qAX,SZ8L_GREE:=/:$U#T<:G- BVagelis, qrv:434.350, tsq:77 {UIV32N}"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for bad digipeater callsign")
	}
	if p.ResultCode != "digicall_badchars" {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, "digicall_badchars")
	}
	if p.Type != "" {
		t.Errorf("type = %q, want empty", p.Type)
	}
}

func TestBadPacketBadSymbolTable(t *testing.T) {
	// Bad symbol table character (comma instead of /, \, or overlay)
	packet := "ASDF>DSALK,OH2RDG*,WIDE:!6028.51N,02505.68E#"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for bad symbol table")
	}
	if p.ResultCode != "sym_inv_table" {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, "sym_inv_table")
	}
}

func TestBadPacketExperimental(t *testing.T) {
	// Unsupported experimental packet format
	packet := "ASDF>DSALK,OH2RDG*,WIDE:{{ unsupported experimental format"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for experimental packet")
	}
	if p.ResultCode != "exp_unsupp" {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, "exp_unsupp")
	}
}
