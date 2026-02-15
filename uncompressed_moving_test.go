package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/21decode-uncomp-moving.t

func TestUncompressedMoving(t *testing.T) {
	header := "OH7FDN>APZMDR,OH7AA-1*,WIDE2-1,qAR,OH7AA"
	// Comment contains inline telemetry that shouldn't break position parsing
	body := "!6253.52N/02739.47E>036/010/A=000465 |!!!!!!!!!!!!!!|"
	packet := header + ":" + body

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}

	if p.SrcCallsign != "OH7FDN" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH7FDN")
	}
	if p.DstCallsign != "APZMDR" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "APZMDR")
	}
	if p.Header != header {
		t.Errorf("header = %q, want %q", p.Header, header)
	}
	if p.Body != body {
		t.Errorf("body = %q, want %q", p.Body, body)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}

	// Digipeaters
	if len(p.Digipeaters) != 4 {
		t.Fatalf("digipeaters count = %d, want 4", len(p.Digipeaters))
	}
	wantDigis := []struct {
		call      string
		wasDigied bool
	}{
		{"OH7AA-1", true},
		{"WIDE2-1", false},
		{"qAR", false},
		{"OH7AA", false},
	}
	for i, want := range wantDigis {
		if p.Digipeaters[i].Call != want.call {
			t.Errorf("digi[%d].call = %q, want %q", i, p.Digipeaters[i].Call, want.call)
		}
		if p.Digipeaters[i].WasDigied != want.wasDigied {
			t.Errorf("digi[%d].wasdigied = %v, want %v", i, p.Digipeaters[i].WasDigied, want.wasDigied)
		}
	}

	if p.SymbolTable != '/' {
		t.Errorf("symboltable = %c, want /", p.SymbolTable)
	}
	if p.SymbolCode != '>' {
		t.Errorf("symbolcode = %c, want >", p.SymbolCode)
	}
	if p.PosAmbiguity == nil || *p.PosAmbiguity != 0 {
		t.Errorf("posambiguity = %v, want 0", p.PosAmbiguity)
	}
	if p.Messaging == nil || *p.Messaging != false {
		t.Errorf("messaging = %v, want false", p.Messaging)
	}

	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "62.8920" {
		t.Errorf("latitude = %s, want 62.8920", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "27.6578" {
		t.Errorf("longitude = %s, want 27.6578", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}

	// Speed: 10 knots * 1.852 = 18.52 km/h
	if p.Speed == nil {
		t.Fatal("speed is nil")
	}
	if got := fmt.Sprintf("%.2f", *p.Speed); got != "18.52" {
		t.Errorf("speed = %s, want 18.52", got)
	}
	if p.Course == nil || *p.Course != 36 {
		t.Errorf("course = %v, want 36", p.Course)
	}
	// Altitude: 465 feet * 0.3048 = 141.732 meters
	if p.Altitude == nil {
		t.Fatal("altitude is nil")
	}
	if got := fmt.Sprintf("%.3f", *p.Altitude); got != "141.732" {
		t.Errorf("altitude = %s, want 141.732", got)
	}
}
