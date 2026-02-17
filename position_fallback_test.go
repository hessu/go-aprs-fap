package fap

import (
	"errors"
	"fmt"
	"testing"
)

// Tests for parsePositionFallback last-resort !-position parsing.

func TestFallbackUncompressed(t *testing.T) {
	// From perl-aprs-fap/t/20decode-uncompressed.t: last resort !-location parsing
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:hoponassualku!6028.51S/02505.68W#PHG7220/RELAY,WIDE, OH2AP Jarvenpaa"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatUncompressed {
		t.Errorf("format = %q, want %q", p.Format, FormatUncompressed)
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

func TestFallbackCompressed(t *testing.T) {
	// Compressed position found via fallback: '!' at offset 5, followed by
	// 13 chars of compressed data.
	packet := "OH2KKU>APRS,TCPIP*:hello!/I0-X;T_Wv&{-Aigate testing"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatCompressed {
		t.Errorf("format = %q, want %q", p.Format, FormatCompressed)
	}
	if p.Latitude == nil || p.Longitude == nil {
		t.Fatal("expected latitude and longitude to be set")
	}
}

func TestFallbackExclAtOffset39(t *testing.T) {
	// '!' at exactly offset 39 (last allowed position) with enough data for uncompressed
	padding := "012345678901234567890123456789012345678" // 39 chars
	packet := "OH2RDP-1>BEACON-15:" + padding + "!6028.51N/02505.68E#"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.4752" {
		t.Errorf("latitude = %s, want 60.4752", got)
	}
}

func TestFallbackExclAtOffset40(t *testing.T) {
	// '!' at offset 40 is too far — should fail as unsupported
	padding := "0123456789012345678901234567890123456789" // 40 chars
	packet := "OH2RDP-1>BEACON-15:" + padding + "!6028.51N/02505.68E#"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for '!' too far into body")
	}
	if !errors.Is(err, ErrTypeNotSupported) {
		t.Errorf("error = %v, want %v", err, ErrTypeNotSupported)
	}
}

func TestFallbackNoExcl(t *testing.T) {
	// No '!' in body at all
	packet := "OH2RDP-1>BEACON-15:Tno position here at all"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for missing '!'")
	}
	if !errors.Is(err, ErrTypeNotSupported) {
		t.Errorf("error = %v, want %v", err, ErrTypeNotSupported)
	}
}

func TestFallbackUncompressedTooShort(t *testing.T) {
	// '!' followed by a digit (looks like uncompressed) but fewer than 19 chars
	packet := "OH2RDP-1>BEACON-15:X!6028.51N/02505.6"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for too-short uncompressed fallback")
	}
	if !errors.Is(err, ErrTypeNotSupported) {
		t.Errorf("error = %v, want %v", err, ErrTypeNotSupported)
	}
}

func TestFallbackCompressedTooShort(t *testing.T) {
	// '!' followed by a valid compressed table char but fewer than 13 chars
	packet := "OH2RDP-1>BEACON-15:X!/I0-X;T_Wv&"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for too-short compressed fallback")
	}
	if !errors.Is(err, ErrTypeNotSupported) {
		t.Errorf("error = %v, want %v", err, ErrTypeNotSupported)
	}
}

func TestFallbackBadCompressedTableChar(t *testing.T) {
	// '!' followed by a character that is not a valid compressed table char
	// and not a digit — should fail as unsupported
	packet := "OH2RDP-1>BEACON-15:X!zI0-X;T_Wv&{-A"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for invalid compressed table char")
	}
	if !errors.Is(err, ErrTypeNotSupported) {
		t.Errorf("error = %v, want %v", err, ErrTypeNotSupported)
	}
}
