package fap

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// Tests ported from perl-aprs-fap/t/55decode-timestamp.t

func TestTimestampRawDHMzUTC(t *testing.T) {
	// Raw timestamp from @..z (DDHHMMz) position packet
	now := time.Now().UTC()
	tstamp := fmt.Sprintf("%02d%02d%02d", now.Day(), now.Hour(), now.Minute())

	packet := "KB3HVP-14>APU25N,N8TJG-10*,WIDE2-1,qAR,LANSNG:@" + tstamp + "z4231.16N/08449.88Wu227/052/A=000941 {UIV32N}"
	p, err := Parse(packet, WithRawTimestamp())
	if err != nil {
		t.Fatalf("failed to parse a position packet with @..z timestamp: %v", err)
	}
	if p.RawTimestamp != tstamp {
		t.Errorf("raw timestamp = %q, want %q", p.RawTimestamp, tstamp)
	}
}

func TestTimestampDecodedDHMzUTC(t *testing.T) {
	// Decoded UNIX timestamp from @..z (DDHHMMz) position packet
	now := time.Now().UTC()
	tstamp := fmt.Sprintf("%02d%02d%02d", now.Day(), now.Hour(), now.Minute())
	outcome := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)

	packet := "KB3HVP-14>APU25N,N8TJG-10*,WIDE2-1,qAR,LANSNG:@" + tstamp + "z4231.16N/08449.88Wu227/052/A=000941 {UIV32N}"
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a position packet with @..z timestamp: %v", err)
	}
	if p.Timestamp == nil {
		t.Fatalf("timestamp is nil, want %v", outcome)
	}
	if !p.Timestamp.Equal(outcome) {
		t.Errorf("timestamp = %v, want %v", *p.Timestamp, outcome)
	}
}

func TestTimestampRawHMSh(t *testing.T) {
	// Raw timestamp from /..h (HHMMSSh) position packet
	packet := "G4EUM-9>APOTC1,G4EUM*,WIDE2-2,qAS,M3SXA-10:/055816h5134.38N/00019.47W>155/023!W26!/A=000188 14.3V 27C HDOP01.0 SATS09"
	p, err := Parse(packet, WithRawTimestamp())
	if err != nil {
		t.Fatalf("failed to parse a position packet with /..h timestamp: %v", err)
	}
	if p.RawTimestamp != "055816" {
		t.Errorf("raw timestamp = %q, want %q", p.RawTimestamp, "055816")
	}
}

func TestTimestampDecodedHMSh(t *testing.T) {
	// Decoded UNIX timestamp from /..h (HHMMSSh) position packet
	now := time.Now().UTC()
	tstamp := fmt.Sprintf("%02d%02d%02d", now.Hour(), now.Minute(), now.Second())
	outcome := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)

	packet := "G4EUM-9>APOTC1,G4EUM*,WIDE2-2,qAS,M3SXA-10:/" + tstamp + "h5134.38N/00019.47W>155/023!W26!/A=000188 14.3V 27C HDOP01.0 SATS09"
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a position packet with /..h timestamp: %v", err)
	}
	if p.Timestamp == nil {
		t.Fatalf("timestamp is nil, want %v", outcome)
	}
	if !p.Timestamp.Equal(outcome) {
		t.Errorf("timestamp = %v, want %v", *p.Timestamp, outcome)
	}
}

func TestTimestampRawDHMLocal(t *testing.T) {
	// Raw timestamp from /../ (DDHHMM/) local time position packet
	packet := "G4EUM-9>APOTC1,G4EUM*,WIDE2-2,qAS,M3SXA-10:/060642/5134.38N/00019.47W>155/023!W26!/A=000188 14.3V 27C HDOP01.0 SATS09"
	p, err := Parse(packet, WithRawTimestamp())
	if err != nil {
		t.Fatalf("failed to parse a position packet with /../ local timestamp: %v", err)
	}
	if p.RawTimestamp != "060642" {
		t.Errorf("raw timestamp = %q, want %q", p.RawTimestamp, "060642")
	}
}

func TestTimestampDecodedDHMLocal(t *testing.T) {
	// Decoded UNIX timestamp from /../ (DDHHMM/) local time position packet
	now := time.Now()
	loc := now.Location()
	tstamp := fmt.Sprintf("%02d%02d%02d", now.Day(), now.Hour(), now.Minute())
	outcome := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, loc)

	packet := "G4EUM-9>APOTC1,G4EUM*,WIDE2-2,qAS,M3SXA-10:/" + tstamp + "/5134.38N/00019.47W>155/023!W26!/A=000188 14.3V 27C HDOP01.0 SATS09"
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a position packet with /../ local timestamp: %v", err)
	}
	if p.Timestamp == nil {
		t.Fatalf("timestamp is nil, want %v", outcome)
	}
	if !p.Timestamp.Equal(outcome) {
		t.Errorf("timestamp = %v, want %v", *p.Timestamp, outcome)
	}
}

func TestTimestampInvalidLocation(t *testing.T) {
	// Invalid timestamp (day 00) in a position packet should produce a warning, not an error.
	packet := "KB3HVP-14>APU25N,N8TJG-10*,WIDE2-1,qAR,LANSNG:@000000z4231.16N/08449.88Wu227/052/A=000941 {UIV32N}"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("parsing failed: %v (should succeed with a warning)", err)
	}
	if p.Timestamp != nil {
		t.Errorf("timestamp = %v, want nil for invalid timestamp", *p.Timestamp)
	}
	if len(p.Warnings) != 1 {
		t.Fatalf("warnings count = %d, want 1", len(p.Warnings))
	}
	if !errors.Is(&p.Warnings[0], ErrTimestampInvalid) {
		t.Errorf("warning code = %q, want %q", p.Warnings[0].Code, ErrTimestampInvalid.Code)
	}
	// Position should still be parsed
	if p.Latitude == nil {
		t.Error("latitude is nil, want parsed position")
	}
}

func TestTimestampInvalidObject(t *testing.T) {
	// Invalid timestamp (day 00) in an object packet should produce a warning, not an error.
	packet := "SRC>APRS,TCPIP*:;TestObj  *000000z4903.50N/07201.75W-Test"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("parsing failed: %v (should succeed with a warning)", err)
	}
	if p.Timestamp != nil {
		t.Errorf("timestamp = %v, want nil for invalid timestamp", *p.Timestamp)
	}
	if len(p.Warnings) != 1 {
		t.Fatalf("warnings count = %d, want 1", len(p.Warnings))
	}
	if !errors.Is(&p.Warnings[0], ErrTimestampInvalid) {
		t.Errorf("warning code = %q, want %q", p.Warnings[0].Code, ErrTimestampInvalid.Code)
	}
	// Position should still be parsed
	if p.Latitude == nil {
		t.Error("latitude is nil, want parsed position")
	}
}

func TestTimestampInvalidStatus(t *testing.T) {
	// Invalid timestamp (day 00) in a status packet should produce a warning, not an error.
	packet := "SRC>APRS,TCPIP*:>000000zStatus text here"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("parsing failed: %v (should succeed with a warning)", err)
	}
	if p.Timestamp != nil {
		t.Errorf("timestamp = %v, want nil for invalid timestamp", *p.Timestamp)
	}
	if len(p.Warnings) != 1 {
		t.Fatalf("warnings count = %d, want 1", len(p.Warnings))
	}
	if !errors.Is(&p.Warnings[0], ErrTimestampInvalid) {
		t.Errorf("warning code = %q, want %q", p.Warnings[0].Code, ErrTimestampInvalid.Code)
	}
	if p.Status != "Status text here" {
		t.Errorf("status = %q, want %q", p.Status, "Status text here")
	}
}
