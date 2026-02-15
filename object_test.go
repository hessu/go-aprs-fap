package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/41decode-object.t

func TestObjectCompressed(t *testing.T) {
	// Compressed format object (packet built with hex encoding in Perl test)
	// OH2KKU-1>APRS,TCPIP*,qAC,FIRST:;SRAL HQ  *100927zS0%E/Th4_a  AKaupinmaenpolku9,open M-Th12-17,F12-14 lcl
	packet := "OH2KKU-1>APRS,TCPIP*,qAC,FIRST:;SRAL HQ  *100927zS0%E/Th4_a  AKaupinmaenpolku9,open M-Th12-17,F12-14 lcl"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse object packet: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}
	if p.Type != PacketTypeObject {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeObject)
	}

	if p.ObjectName != "SRAL HQ  " {
		t.Errorf("objectname = %q, want %q", p.ObjectName, "SRAL HQ  ")
	}
	if p.Alive == nil || !*p.Alive {
		t.Error("expected alive = true")
	}

	if p.SymbolTable != 'S' {
		t.Errorf("symboltable = %c, want S", p.SymbolTable)
	}
	if p.SymbolCode != 'a' {
		t.Errorf("symbolcode = %c, want a", p.SymbolCode)
	}

	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.2305" {
		t.Errorf("latitude = %s, want 60.2305", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "24.8790" {
		t.Errorf("longitude = %s, want 24.8790", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.3f", *p.PosResolution); got != "0.291" {
		t.Errorf("posresolution = %s, want 0.291", got)
	}

	if p.PHG != "" {
		t.Errorf("phg = %q, want empty", p.PHG)
	}

	wantComment := "Kaupinmaenpolku9,open M-Th12-17,F12-14 lcl"
	if p.Comment != wantComment {
		t.Errorf("comment = %q, want %q", p.Comment, wantComment)
	}
}

func TestObjectUncompressed(t *testing.T) {
	// Regular APRS uncompressed position object
	packet := "OH2KKU-1>APRS:;LEADER   *092345z4903.50N/07201.75W>088/036"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse object packet: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}
	if p.Type != PacketTypeObject {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeObject)
	}

	if p.ObjectName != "LEADER   " {
		t.Errorf("objectname = %q, want %q", p.ObjectName, "LEADER   ")
	}
	if p.Alive == nil || !*p.Alive {
		t.Error("expected alive = true")
	}

	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "49.0583" {
		t.Errorf("latitude = %s, want 49.0583", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-72.0292" {
		t.Errorf("longitude = %s, want -72.0292", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.3f", *p.PosResolution); got != "18.520" {
		t.Errorf("posresolution = %s, want 18.520", got)
	}

	if p.PHG != "" {
		t.Errorf("phg = %q, want empty", p.PHG)
	}
	if p.Comment != "" {
		t.Errorf("comment = %q, want empty", p.Comment)
	}
}

func TestObjectKilled(t *testing.T) {
	// Killed object (underscore instead of asterisk)
	packet := "OH2KKU-1>APRS:;LEADER   _092345z4903.50N/07201.75W>088/036"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse killed object packet: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}
	if p.Type != PacketTypeObject {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeObject)
	}

	if p.ObjectName != "LEADER   " {
		t.Errorf("objectname = %q, want %q", p.ObjectName, "LEADER   ")
	}
	if p.Alive == nil || *p.Alive {
		t.Error("expected alive = false")
	}
}
