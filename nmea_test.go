package fap

// Tests for NMEA sentence parsing ($GPRMC, $GPGGA, $GPGLL).
// GPRMC tests ported from perl-aprs-fap/t/24decode-gprmc.t.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGPRMC(t *testing.T) {
	packet := "N0CALL-11>APRS,DIGI*,WIDE2-1,qAR,IGATE:$GPRMC,145526,A,3349.0378,N,08406.2617,W,23.726,27.9,121207,4.9,W*7A"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a GPRMC NMEA packet: %v", err)
	}

	if p.Header != "N0CALL-11>APRS,DIGI*,WIDE2-1,qAR,IGATE" {
		t.Errorf("header = %q, want %q", p.Header, "N0CALL-11>APRS,DIGI*,WIDE2-1,qAR,IGATE")
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

func TestGPRMCNoFix(t *testing.T) {
	packet := "SRC>APRS,WIDE1-1,GATE,qAR,GATE-1:$GPRMC,041518,V,,,,,,,230309,,*33"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for GPRMC with no valid fix, got nil")
	}
	if !errors.Is(err, ErrGPRMCNoFix) {
		t.Errorf("error = %v, want ErrGPRMCNoFix", err)
	}
}

func TestGPGGA(t *testing.T) {
	packet := "N0CALL-11>APRS,WIDE2-1,qAR,IGATE:$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,47.0,M,,*4F"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a GPGGA NMEA packet: %v", err)
	}

	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatNMEA {
		t.Errorf("format = %q, want %q", p.Format, FormatNMEA)
	}

	if p.ChecksumOK == nil || !*p.ChecksumOK {
		t.Errorf("checksumok = %v, want true", p.ChecksumOK)
	}

	if p.Latitude == nil {
		t.Fatalf("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "48.1173" {
		t.Errorf("latitude = %s, want 48.1173", got)
	}

	if p.Longitude == nil {
		t.Fatalf("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "11.5167" {
		t.Errorf("longitude = %s, want 11.5167", got)
	}

	if p.PosResolution == nil {
		t.Fatalf("posresolution is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.PosResolution); got != "1.8520" {
		t.Errorf("posresolution = %s, want 1.8520", got)
	}

	if p.Altitude == nil {
		t.Fatalf("altitude is nil")
	}
	if got := fmt.Sprintf("%.1f", *p.Altitude); got != "545.4" {
		t.Errorf("altitude = %s, want 545.4", got)
	}

	// Fields not present in GPGGA
	if p.Speed != nil {
		t.Errorf("speed = %v, want nil", *p.Speed)
	}
	if p.Course != nil {
		t.Errorf("course = %v, want nil", *p.Course)
	}
	if p.Timestamp != nil {
		t.Errorf("timestamp = %v, want nil", *p.Timestamp)
	}
}

func TestGPGGANoFix(t *testing.T) {
	// Fix quality 0 means no fix
	packet := "N0CALL-11>APRS,WIDE2-1,qAR,IGATE:$GPGGA,123519,4807.038,N,01131.000,E,0,08,0.9,545.4,M,47.0,M,,*48"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for GPGGA with no fix")
	}
	if !errors.Is(err, ErrNMEAInvalid) {
		t.Errorf("error = %v, want ErrNMEAInvalid", err)
	}
}

func TestGPGLL(t *testing.T) {
	packet := "N0CALL-11>APRS,WIDE2-1,qAR,IGATE:$GPGLL,4916.45,N,12311.12,W,225444,A*31"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a GPGLL NMEA packet: %v", err)
	}

	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatNMEA {
		t.Errorf("format = %q, want %q", p.Format, FormatNMEA)
	}

	if p.ChecksumOK == nil || !*p.ChecksumOK {
		t.Errorf("checksumok = %v, want true", p.ChecksumOK)
	}

	if p.Latitude == nil {
		t.Fatalf("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "49.2742" {
		t.Errorf("latitude = %s, want 49.2742", got)
	}

	if p.Longitude == nil {
		t.Fatalf("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-123.1853" {
		t.Errorf("longitude = %s, want -123.1853", got)
	}

	if p.PosResolution == nil {
		t.Fatalf("posresolution is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.PosResolution); got != "18.5200" {
		t.Errorf("posresolution = %s, want 18.5200", got)
	}

	// Fields not present in GPGLL
	if p.Altitude != nil {
		t.Errorf("altitude = %v, want nil", *p.Altitude)
	}
	if p.Speed != nil {
		t.Errorf("speed = %v, want nil", *p.Speed)
	}
	if p.Course != nil {
		t.Errorf("course = %v, want nil", *p.Course)
	}
	if p.Timestamp != nil {
		t.Errorf("timestamp = %v, want nil", *p.Timestamp)
	}
}

func TestGPGLLMinimal(t *testing.T) {
	// GPGLL with only 5 fields (no time, no status) — minimum valid
	packet := "N0CALL-11>APRS,WIDE2-1,qAR,IGATE:$GPGLL,4916.45,N,12311.12,W"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a minimal GPGLL NMEA packet: %v", err)
	}

	if p.Latitude == nil {
		t.Fatalf("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "49.2742" {
		t.Errorf("latitude = %s, want 49.2742", got)
	}

	if p.Longitude == nil {
		t.Fatalf("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-123.1853" {
		t.Errorf("longitude = %s, want -123.1853", got)
	}
}

func TestParseNMEACoordWithResErrors(t *testing.T) {
	tests := []struct {
		name       string
		coord      string
		hemisphere string
		isLon      bool
		errIs      string
	}{
		// Empty inputs
		{"empty coord", "", "N", false, "empty coordinate or hemisphere"},
		{"empty hemisphere", "4807.038", "", false, "empty coordinate or hemisphere"},
		{"both empty", "", "", false, "empty coordinate or hemisphere"},

		// Too short
		{"lat too short", "48", "N", false, "coordinate too short"},
		{"lon too short", "011", "E", true, "coordinate too short"},

		// Invalid degrees
		{"lat invalid degrees", "XX07.038", "N", false, "invalid degrees"},
		{"lon invalid degrees", "XXX31.000", "E", true, "invalid degrees"},

		// Invalid minutes
		{"lat invalid minutes", "48XX.XXX", "N", false, "invalid minutes"},
		{"lon invalid minutes", "011XX.XXX", "E", true, "invalid minutes"},

		// Out of range: latitude > 90
		{"lat out of range", "9100.000", "N", false, "coordinate out of range"},
		// Out of range: longitude > 180
		{"lon out of range", "18100.000", "E", true, "coordinate out of range"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseNMEACoordWithRes(tc.coord, tc.hemisphere, tc.isLon)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errIs)
			}
			if !strings.Contains(err.Error(), tc.errIs) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.errIs)
			}
		})
	}
}

func TestParseNMEAErrors(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		errIs error
		msgIs string
	}{
		{
			name:  "not starting with $GP",
			body:  "$XXXXX,A,B,C",
			errIs: ErrNMEAInvalid,
			msgIs: "must start with $GP",
		},
		{
			name:  "checksum mismatch",
			body:  "$GPRMC,145526,A,3349.0378,N,08406.2617,W,23.726,27.9,121207,4.9,W*FF",
			errIs: ErrNMEAInvalid,
			msgIs: "checksum mismatch",
		},
		{
			name:  "too few fields",
			body:  "$GPXXX",
			errIs: ErrNMEAShort,
			msgIs: "too short",
		},
		{
			name:  "unsupported sentence",
			body:  "$GPVTG,0,T,0,M,0,N,0,K",
			errIs: ErrNMEAInvalid,
			msgIs: "unsupported NMEA sentence",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet := "N0CALL>APRS,WIDE2-1:" + tc.body
			_, err := Parse(packet)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
			if !strings.Contains(err.Error(), tc.msgIs) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.msgIs)
			}
		})
	}
}

func TestParseGPRMCErrors(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		errIs error
	}{
		{
			name:  "too few fields",
			body:  "$GPRMC,145526,A,3349.0378,N,08406.2617,W,23.726,27.9",
			errIs: ErrNMEAShort,
		},
		{
			name:  "no fix",
			body:  "$GPRMC,145526,V,3349.0378,N,08406.2617,W,23.726,27.9,121207,4.9,W",
			errIs: ErrGPRMCNoFix,
		},
		{
			name:  "invalid latitude",
			body:  "$GPRMC,145526,A,XXXX.XXXX,N,08406.2617,W,23.726,27.9,121207,4.9,W",
			errIs: ErrPosLatInvalid,
		},
		{
			name:  "empty latitude",
			body:  "$GPRMC,145526,A,,N,08406.2617,W,23.726,27.9,121207,4.9,W",
			errIs: ErrPosLatInvalid,
		},
		{
			name:  "invalid longitude",
			body:  "$GPRMC,145526,A,3349.0378,N,XXXXX.XXXX,W,23.726,27.9,121207,4.9,W",
			errIs: ErrPosLonInvalid,
		},
		{
			name:  "empty longitude",
			body:  "$GPRMC,145526,A,3349.0378,N,,W,23.726,27.9,121207,4.9,W",
			errIs: ErrPosLonInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet := "N0CALL>APRS,WIDE2-1:" + tc.body
			_, err := Parse(packet)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.errIs)
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
		})
	}
}

func TestParseGPGGAErrors(t *testing.T) {
	// Valid base: $GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,47.0,M,,
	tests := []struct {
		name  string
		body  string
		errIs error
	}{
		{
			name:  "too few fields",
			body:  "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4",
			errIs: ErrNMEAShort,
		},
		{
			name:  "no fix",
			body:  "$GPGGA,123519,4807.038,N,01131.000,E,0,08,0.9,545.4,M,47.0,M,,",
			errIs: ErrNMEAInvalid,
		},
		{
			name:  "invalid latitude",
			body:  "$GPGGA,123519,XXXX.XXX,N,01131.000,E,1,08,0.9,545.4,M,47.0,M,,",
			errIs: ErrPosLatInvalid,
		},
		{
			name:  "empty latitude",
			body:  "$GPGGA,123519,,N,01131.000,E,1,08,0.9,545.4,M,47.0,M,,",
			errIs: ErrPosLatInvalid,
		},
		{
			name:  "invalid longitude",
			body:  "$GPGGA,123519,4807.038,N,XXXXX.XXX,E,1,08,0.9,545.4,M,47.0,M,,",
			errIs: ErrPosLonInvalid,
		},
		{
			name:  "empty longitude",
			body:  "$GPGGA,123519,4807.038,N,,E,1,08,0.9,545.4,M,47.0,M,,",
			errIs: ErrPosLonInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet := "N0CALL>APRS,WIDE2-1:" + tc.body
			_, err := Parse(packet)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.errIs)
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
		})
	}
}

func TestParseGPGLLErrors(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		errIs error
	}{
		{
			name:  "too few fields",
			body:  "$GPGLL,4916.45,N,12311.12",
			errIs: ErrNMEAShort,
		},
		{
			name:  "invalid latitude",
			body:  "$GPGLL,XXXX.XX,N,12311.12,W",
			errIs: ErrPosLatInvalid,
		},
		{
			name:  "empty latitude",
			body:  "$GPGLL,,N,12311.12,W",
			errIs: ErrPosLatInvalid,
		},
		{
			name:  "invalid longitude",
			body:  "$GPGLL,4916.45,N,XXXXX.XX,W",
			errIs: ErrPosLonInvalid,
		},
		{
			name:  "empty longitude",
			body:  "$GPGLL,4916.45,N,,W",
			errIs: ErrPosLonInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet := "N0CALL>APRS,WIDE2-1:" + tc.body
			_, err := Parse(packet)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.errIs)
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
		})
	}
}

func TestParseGPRMCTimestampErrors(t *testing.T) {
	// Build a GPRMC packet with the given time and date fields.
	// A valid base: $GPRMC,HHMMSS,A,3349.0378,N,08406.2617,W,0,0,DDMMYY,0,W
	mkPacket := func(timeStr, dateStr string) string {
		return fmt.Sprintf("N0CALL>APRS,WIDE2-1:$GPRMC,%s,A,3349.0378,N,08406.2617,W,0,0,%s,0,W", timeStr, dateStr)
	}

	tests := []struct {
		name  string
		time  string
		date  string
		errIs string // substring in error message
	}{
		// Time errors
		{"time too short", "1234", "121207", "invalid time"},
		{"time too long", "12345678", "121207", "invalid time"},
		{"time empty", "", "121207", "invalid time"},
		{"hour non-numeric", "XX0000", "121207", "invalid time"},
		{"hour 24", "240000", "121207", "invalid time"},
		{"minute non-numeric", "12XX00", "121207", "invalid time"},
		{"minute 60", "126000", "121207", "invalid time"},
		{"second non-numeric", "1200XX", "121207", "invalid time"},
		{"second 60", "120060", "121207", "invalid time"},

		// Date errors
		{"date too short", "145526", "1212", "invalid date"},
		{"date too long", "145526", "12120712", "invalid date"},
		{"date empty", "145526", "", "invalid date"},
		{"day non-numeric", "145526", "XX1207", "invalid date"},
		{"month non-numeric", "145526", "12XX07", "invalid date"},
		{"year non-numeric", "145526", "1212XX", "invalid date"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(mkPacket(tc.time, tc.date))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errIs)
			}
			if !errors.Is(err, ErrNMEAInvalid) {
				t.Errorf("error = %v, want ErrNMEAInvalid", err)
			}
			if !strings.Contains(err.Error(), tc.errIs) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.errIs)
			}
		})
	}
}
