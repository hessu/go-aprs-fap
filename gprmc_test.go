package fap

// Tests ported from perl-aprs-fap/t/24decode-gprmc.t

import (
	"fmt"
	"testing"
	"time"
)

func TestGPRMC(t *testing.T) {
	packet := "OH7LZB-11>APRS,W4GR*,WIDE2-1,qAR,WA4DSY:$GPRMC,145526,A,3349.0378,N,08406.2617,W,23.726,27.9,121207,4.9,W*7A"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a GPRMC NMEA packet: %v", err)
	}

	if p.Header != "OH7LZB-11>APRS,W4GR*,WIDE2-1,qAR,WA4DSY" {
		t.Errorf("header = %q, want %q", p.Header, "OH7LZB-11>APRS,W4GR*,WIDE2-1,qAR,WA4DSY")
	}
	if p.Body != "$GPRMC,145526,A,3349.0378,N,08406.2617,W,23.726,27.9,121207,4.9,W*7A" {
		t.Errorf("body = %q, want %q", p.Body, "$GPRMC,145526,A,3349.0378,N,08406.2617,W,23.726,27.9,121207,4.9,W*7A")
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatNMEA {
		t.Errorf("format = %q, want %q", p.Format, FormatNMEA)
	}

	// check for undefined value, when there is no such data in the packet
	if p.PosAmbiguity != nil {
		t.Errorf("posambiguity = %v, want nil", *p.PosAmbiguity)
	}
	if p.Messaging != nil {
		t.Errorf("messaging = %v, want nil", p.Messaging)
	}

	if p.ChecksumOK == nil || !*p.ChecksumOK {
		t.Errorf("checksumok = %v, want true", p.ChecksumOK)
	}

	// timestamp = 1197471326 (2007-12-12 14:55:26 UTC)
	expectedTS := time.Unix(1197471326, 0).UTC()
	if p.Timestamp == nil {
		t.Fatalf("timestamp is nil, want %v", expectedTS)
	}
	if !p.Timestamp.Equal(expectedTS) {
		t.Errorf("timestamp = %v (%d), want %v (%d)", *p.Timestamp, p.Timestamp.Unix(), expectedTS, expectedTS.Unix())
	}

	if p.Latitude == nil {
		t.Fatalf("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "33.8173" {
		t.Errorf("latitude = %s, want 33.8173", got)
	}

	if p.Longitude == nil {
		t.Fatalf("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-84.1044" {
		t.Errorf("longitude = %s, want -84.1044", got)
	}

	if p.PosResolution == nil {
		t.Fatalf("posresolution is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.PosResolution); got != "0.1852" {
		t.Errorf("posresolution = %s, want 0.1852", got)
	}

	if p.Speed == nil {
		t.Fatalf("speed is nil")
	}
	if got := fmt.Sprintf("%.2f", *p.Speed); got != "43.94" {
		t.Errorf("speed = %s, want 43.94", got)
	}

	if p.Course == nil {
		t.Fatalf("course is nil")
	}
	if *p.Course != 28 {
		t.Errorf("course = %d, want 28", *p.Course)
	}

	if p.Altitude != nil {
		t.Errorf("altitude = %v, want nil", *p.Altitude)
	}
}
