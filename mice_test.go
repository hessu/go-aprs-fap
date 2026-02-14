package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/23decode-mice.t

func TestMicENonMoving(t *testing.T) {
	packet := "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH:'I',l \x1C>/]"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse a non-moving target's mic-e packet: %v", err)
	}

	if p.SrcCallsign != "OH7LZB-13" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH7LZB-13")
	}
	if p.DstCallsign != "SX15S6" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "SX15S6")
	}
	if p.Header != "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH" {
		t.Errorf("header = %q, want %q", p.Header, "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH")
	}
	if p.Body != "'I',l \x1C>/]" {
		t.Errorf("body = %q, want %q", p.Body, "'I',l \x1C>/]")
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatMicE {
		t.Errorf("format = %q, want %q", p.Format, FormatMicE)
	}
	if p.Comment != "]" {
		t.Errorf("comment = %q, want %q", p.Comment, "]")
	}

	// Digipeaters
	if len(p.Digipeaters) != 3 {
		t.Fatalf("digipeaters count = %d, want 3", len(p.Digipeaters))
	}
	if p.Digipeaters[0].Call != "TCPIP" {
		t.Errorf("digi[0].call = %q, want %q", p.Digipeaters[0].Call, "TCPIP")
	}
	if !p.Digipeaters[0].WasDigied {
		t.Errorf("digi[0].wasdigied = false, want true")
	}
	if p.Digipeaters[1].Call != "qAC" {
		t.Errorf("digi[1].call = %q, want %q", p.Digipeaters[1].Call, "qAC")
	}
	if p.Digipeaters[1].WasDigied {
		t.Errorf("digi[1].wasdigied = true, want false")
	}
	if p.Digipeaters[2].Call != "FOURTH" {
		t.Errorf("digi[2].call = %q, want %q", p.Digipeaters[2].Call, "FOURTH")
	}
	if p.Digipeaters[2].WasDigied {
		t.Errorf("digi[2].wasdigied = true, want false")
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
	if p.Messaging != nil {
		t.Errorf("messaging = %v, want nil", p.Messaging)
	}

	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "-38.2560" {
		t.Errorf("latitude = %s, want -38.2560", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "145.1860" {
		t.Errorf("longitude = %s, want 145.1860", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}

	if p.Speed == nil || *p.Speed != 0 {
		t.Errorf("speed = %v, want 0", p.Speed)
	}
	if p.Course == nil || *p.Course != 0 {
		t.Errorf("course = %v, want 0", p.Course)
	}
	if p.Altitude != nil {
		t.Errorf("altitude = %v, want nil", *p.Altitude)
	}
}

func TestMicEMoving(t *testing.T) {
	packet := "OH7LZB-2>TQ4W2V,WIDE2-1,qAo,OH7LZB:`c51!f?>/]\"3x}="
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse a moving target's mic-e: %v", err)
	}

	if p.SrcCallsign != "OH7LZB-2" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH7LZB-2")
	}
	if p.DstCallsign != "TQ4W2V" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "TQ4W2V")
	}
	if p.Header != "OH7LZB-2>TQ4W2V,WIDE2-1,qAo,OH7LZB" {
		t.Errorf("header = %q, want %q", p.Header, "OH7LZB-2>TQ4W2V,WIDE2-1,qAo,OH7LZB")
	}
	if p.Body != "`c51!f?>/]\"3x}=" {
		t.Errorf("body = %q, want %q", p.Body, "`c51!f?>/]\"3x}=")
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}

	if p.Comment != "]=" {
		t.Errorf("comment = %q, want %q", p.Comment, "]=")
	}

	// Digipeaters
	if len(p.Digipeaters) != 3 {
		t.Fatalf("digipeaters count = %d, want 3", len(p.Digipeaters))
	}
	if p.Digipeaters[0].Call != "WIDE2-1" {
		t.Errorf("digi[0].call = %q, want %q", p.Digipeaters[0].Call, "WIDE2-1")
	}
	if p.Digipeaters[0].WasDigied {
		t.Errorf("digi[0].wasdigied = true, want false")
	}
	if p.Digipeaters[1].Call != "qAo" {
		t.Errorf("digi[1].call = %q, want %q", p.Digipeaters[1].Call, "qAo")
	}
	if p.Digipeaters[1].WasDigied {
		t.Errorf("digi[1].wasdigied = true, want false")
	}
	if p.Digipeaters[2].Call != "OH7LZB" {
		t.Errorf("digi[2].call = %q, want %q", p.Digipeaters[2].Call, "OH7LZB")
	}
	if p.Digipeaters[2].WasDigied {
		t.Errorf("digi[2].wasdigied = true, want false")
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
	if p.Messaging != nil {
		t.Errorf("messaging = %v, want nil", p.Messaging)
	}
	if p.MBits != "110" {
		t.Errorf("mbits = %q, want %q", p.MBits, "110")
	}

	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "41.7877" {
		t.Errorf("latitude = %s, want 41.7877", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-71.4202" {
		t.Errorf("longitude = %s, want -71.4202", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}

	if p.Speed == nil {
		t.Fatal("speed is nil")
	}
	if got := fmt.Sprintf("%.2f", *p.Speed); got != "105.56" {
		t.Errorf("speed = %s, want 105.56", got)
	}
	if p.Course == nil || *p.Course != 35 {
		t.Errorf("course = %v, want 35", p.Course)
	}
	if p.Altitude == nil || *p.Altitude != 6 {
		t.Errorf("altitude = %v, want 6", p.Altitude)
	}
}

func TestMicEInvalidSymbolTable(t *testing.T) {
	packet := "OZ2BRN-4>5U2V08,OZ3RIN-3,OZ4DIA-2*,WIDE2-1,qAR,DB0KUE:`'O<l!{,,\"4R}"
	p, err := ParseAPRS(packet)
	if err == nil {
		t.Fatal("expected error for invalid symbol table, got nil")
	}

	if p.ResultCode != ErrSymInvTable {
		t.Errorf("resultcode = %q, want %q", p.ResultCode, ErrSymInvTable)
	}
	if p.SrcCallsign != "OZ2BRN-4" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OZ2BRN-4")
	}
	if p.DstCallsign != "5U2V08" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "5U2V08")
	}
	if p.Header != "OZ2BRN-4>5U2V08,OZ3RIN-3,OZ4DIA-2*,WIDE2-1,qAR,DB0KUE" {
		t.Errorf("header = %q, want %q", p.Header, "OZ2BRN-4>5U2V08,OZ3RIN-3,OZ4DIA-2*,WIDE2-1,qAR,DB0KUE")
	}
	if p.Body != "`'O<l!{,,\"4R}" {
		t.Errorf("body = %q, want %q", p.Body, "`'O<l!{,,\"4R}")
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Comment != "" {
		t.Errorf("comment = %q, want empty", p.Comment)
	}
}

func TestMicEHexTelemetry5Ch(t *testing.T) {
	// 5-channel Mic-E hex telemetry
	packet := "OZ2BRN-4>5U2V08,WIDE2-1,qAo,OH7LZB:`c51!f?>/'102030FFff commeeeent"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse mic-e packet with 5-ch telemetry: %v", err)
	}

	if p.Comment != "commeeeent" {
		t.Errorf("comment = %q, want %q", p.Comment, "commeeeent")
	}

	if p.TelemetryData == nil {
		t.Fatal("no telemetry data")
	}
	vals := p.TelemetryData.Vals
	if len(vals) != 5 {
		t.Fatalf("telemetry vals count = %d, want 5", len(vals))
	}
	expected := []float64{16, 32, 48, 255, 255}
	for i, want := range expected {
		if vals[i] == nil {
			t.Errorf("vals[%d] = nil, want %.0f", i, want)
		} else if *vals[i] != want {
			t.Errorf("vals[%d] = %.0f, want %.0f", i, *vals[i], want)
		}
	}
}

func TestMicEHexTelemetry2Ch(t *testing.T) {
	// 2-channel Mic-E hex telemetry (channels 1 and 3, channel 2 is zero)
	packet := "OZ2BRN-4>5U2V08,WIDE2-1,qAo,OH7LZB:`c51!f?>/'1020 commeeeent"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse mic-e packet with 2-ch telemetry: %v", err)
	}

	if p.Comment != "commeeeent" {
		t.Errorf("comment = %q, want %q", p.Comment, "commeeeent")
	}

	if p.TelemetryData == nil {
		t.Fatal("no telemetry data")
	}
	vals := p.TelemetryData.Vals
	if len(vals) != 3 {
		t.Fatalf("telemetry vals count = %d, want 3", len(vals))
	}
	expected := []float64{16, 0, 32}
	for i, want := range expected {
		if vals[i] == nil {
			t.Errorf("vals[%d] = nil, want %.0f", i, want)
		} else if *vals[i] != want {
			t.Errorf("vals[%d] = %.0f, want %.0f", i, *vals[i], want)
		}
	}
}

func TestMicEMangled(t *testing.T) {
	// Packet with a binary byte removed, parsed with AcceptBrokenMicE
	comment := "]Greetings via ISS="
	packet := "KD0KZE>TUPX9R,RS0ISS*,qAR,K0GDI-6:'yaIl -/" + comment
	p, err := ParseAPRS(packet, Options{AcceptBrokenMicE: true})
	if err != nil {
		t.Fatalf("failed to parse mangled mic-e packet: %v", err)
	}

	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "45.1487" {
		t.Errorf("latitude = %s, want 45.1487", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-93.1575" {
		t.Errorf("longitude = %s, want -93.1575", got)
	}
	if p.SymbolTable != '/' {
		t.Errorf("symboltable = %c, want /", p.SymbolTable)
	}
	if p.SymbolCode != '-' {
		t.Errorf("symbolcode = %c, want -", p.SymbolCode)
	}
	if p.Comment != comment {
		t.Errorf("comment = %q, want %q", p.Comment, comment)
	}
	if p.Course != nil {
		t.Errorf("course = %v, want nil", *p.Course)
	}
	if p.Speed != nil {
		t.Errorf("speed = %v, want nil", *p.Speed)
	}
	if !p.MiceMangled {
		t.Errorf("mice_mangled = false, want true")
	}
}
