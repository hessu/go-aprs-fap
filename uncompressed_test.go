package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/20decode-uncompressed.t

func TestUncompressedNortheast(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!6028.51N/02505.68E#PHG7220/RELAY,WIDE, OH2AP Jarvenpaa"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.Format != FormatUncompressed {
		t.Errorf("format = %q, want %q", p.Format, FormatUncompressed)
	}
	if p.SrcCallsign != "OH2RDP-1" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2RDP-1")
	}
	if p.DstCallsign != "BEACON-15" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "BEACON-15")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.4752" {
		t.Errorf("latitude = %s, want 60.4752", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "25.0947" {
		t.Errorf("longitude = %s, want 25.0947", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}
	if p.PHG != "7220" {
		t.Errorf("phg = %q, want %q", p.PHG, "7220")
	}
	if p.Comment != "RELAY,WIDE, OH2AP Jarvenpaa" {
		t.Errorf("comment = %q, want %q", p.Comment, "RELAY,WIDE, OH2AP Jarvenpaa")
	}

	// Digipeaters
	if len(p.Digipeaters) != 2 {
		t.Fatalf("digipeaters count = %d, want 2", len(p.Digipeaters))
	}
	if p.Digipeaters[0].Call != "OH2RDG" {
		t.Errorf("digi[0].call = %q, want %q", p.Digipeaters[0].Call, "OH2RDG")
	}
	if !p.Digipeaters[0].WasDigied {
		t.Error("digi[0].wasdigied = false, want true")
	}
	if p.Digipeaters[1].Call != "WIDE" {
		t.Errorf("digi[1].call = %q, want %q", p.Digipeaters[1].Call, "WIDE")
	}
	if p.Digipeaters[1].WasDigied {
		t.Error("digi[1].wasdigied = true, want false")
	}
}

func TestUncompressedSouthwest(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!6028.51S/02505.68W#PHG7220RELAY,WIDE, OH2AP Jarvenpaa"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "-60.4752" {
		t.Errorf("latitude = %s, want -60.4752", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-25.0947" {
		t.Errorf("longitude = %s, want -25.0947", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}
}

func TestUncompressedAmbiguity3(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!602 .  S/0250 .  W#PHG7220RELAY,WIDE, OH2AP Jarvenpaa"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "-60.4167" {
		t.Errorf("latitude = %s, want -60.4167", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-25.0833" {
		t.Errorf("longitude = %s, want -25.0833", got)
	}
	if p.PosAmbiguity == nil || *p.PosAmbiguity != 3 {
		t.Errorf("posambiguity = %v, want 3", p.PosAmbiguity)
	}
	if got := fmt.Sprintf("%.0f", *p.PosResolution); got != "18520" {
		t.Errorf("posresolution = %s, want 18520", got)
	}
}

func TestUncompressedAmbiguity4(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!60  .  S/025  .  W#PHG7220RELAY,WIDE, OH2AP Jarvenpaa"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "-60.5000" {
		t.Errorf("latitude = %s, want -60.5000", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-25.5000" {
		t.Errorf("longitude = %s, want -25.5000", got)
	}
	if p.PosAmbiguity == nil || *p.PosAmbiguity != 4 {
		t.Errorf("posambiguity = %v, want 4", p.PosAmbiguity)
	}
	if got := fmt.Sprintf("%.0f", *p.PosResolution); got != "111120" {
		t.Errorf("posresolution = %s, want 111120", got)
	}
}

func TestUncompressedLastResort(t *testing.T) {
	// Last-resort !-location parsing: body starts with non-APRS text, position found at '!'
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:hoponassualku!6028.51S/02505.68W#PHG7220RELAY,WIDE, OH2AP Jarvenpaa"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "-60.4752" {
		t.Errorf("latitude = %s, want -60.4752", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-25.0947" {
		t.Errorf("longitude = %s, want -25.0947", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}
	if p.Comment != "RELAY,WIDE, OH2AP Jarvenpaa" {
		t.Errorf("comment = %q, want %q", p.Comment, "RELAY,WIDE, OH2AP Jarvenpaa")
	}
}

func TestUncompressedWxSymbolComment(t *testing.T) {
	// Station with WX symbol (_). Comment is ignored because it gets confused with weather data.
	packet := "A0RID-1>KC0PID-7,WIDE1,qAR,NX0R-6:=3851.38N/09908.75W_Home of KA0RID"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "38.8563" {
		t.Errorf("latitude = %s, want 38.8563", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-99.1458" {
		t.Errorf("longitude = %s, want -99.1458", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}
	if p.Comment != "" {
		t.Errorf("comment = %q, want empty", p.Comment)
	}
}

func TestUncompressedWhitespaceTrimming(t *testing.T) {
	// Whitespace should be trimmed from comment
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!6028.51N/02505.68E#PHG7220   RELAY,WIDE, OH2AP Jarvenpaa  \t "

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.PHG != "7220" {
		t.Errorf("phg = %q, want %q", p.PHG, "7220")
	}
	if p.Comment != "RELAY,WIDE, OH2AP Jarvenpaa" {
		t.Errorf("comment = %q, want %q", p.Comment, "RELAY,WIDE, OH2AP Jarvenpaa")
	}
}

func TestUncompressedTimestampAltitude(t *testing.T) {
	// Position with timestamp and altitude
	packet := "YB1RUS-9>APOTC1,WIDE2-2,qAS,YC0GIN-1:/180000z0609.31S/10642.85E>058/010/A=000079 13.8V 15CYB1RUS-9 Mobile Tracker"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.5f", *p.Latitude); got != "-6.15517" {
		t.Errorf("latitude = %s, want -6.15517", got)
	}
	if got := fmt.Sprintf("%.5f", *p.Longitude); got != "106.71417" {
		t.Errorf("longitude = %s, want 106.71417", got)
	}
	if p.Altitude == nil {
		t.Fatal("altitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Altitude); got != "24.07920" {
		t.Errorf("altitude = %s, want 24.07920", got)
	}
}

func TestUncompressedNegativeAltitude(t *testing.T) {
	// Position with negative altitude
	packet := "YB1RUS-9>APOTC1,WIDE2-2,qAS,YC0GIN-1:/180000z0609.31S/10642.85E>058/010/A=-00079 13.8V 15CYB1RUS-9 Mobile Tracker"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.Altitude == nil {
		t.Fatal("altitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Altitude); got != "-24.07920" {
		t.Errorf("altitude = %s, want -24.07920", got)
	}
}

func TestUncompressedBasicYC0SHR(t *testing.T) {
	// Rather basic position packet
	packet := "YC0SHR>APU25N,TCPIP*,qAC,ALDIMORI:=0606.23S/10644.61E-GW SAHARA PENJARINGAN JAKARTA 147.880 MHz"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if got := fmt.Sprintf("%.5f", *p.Latitude); got != "-6.10383" {
		t.Errorf("latitude = %s, want -6.10383", got)
	}
	if got := fmt.Sprintf("%.5f", *p.Longitude); got != "106.74350" {
		t.Errorf("longitude = %s, want 106.74350", got)
	}
}
