package fap

import (
	"errors"
	"fmt"
	"strings"
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

func TestParseTimestampErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		errIs string // substring expected in error message
	}{
		// Wrong length
		{"too short", "12345z", "7 characters"},
		{"too long", "12345678", "7 characters"},
		{"empty", "", "7 characters"},

		// Unknown indicator
		{"unknown indicator", "010000x", "unknown timestamp indicator"},

		// DDHHMMz - day/hours/minutes UTC
		{"z: day zero", "000000z", "invalid day"},
		{"z: day 32", "320000z", "invalid day"},
		{"z: day non-numeric", "ab0000z", "invalid day"},
		{"z: hours 24", "012400z", "invalid hours"},
		{"z: hours non-numeric", "01xx00z", "invalid hours"},
		{"z: minutes 60", "010060z", "invalid minutes"},
		{"z: minutes non-numeric", "0100xxz", "invalid minutes"},

		// DDHHMM/ - day/hours/minutes local
		{"/: day zero", "000000/", "invalid day"},
		{"/: day 32", "320000/", "invalid day"},
		{"/: day non-numeric", "ab0000/", "invalid day"},
		{"/: hours 24", "012400/", "invalid hours"},
		{"/: hours non-numeric", "01xx00/", "invalid hours"},
		{"/: minutes 60", "010060/", "invalid minutes"},
		{"/: minutes non-numeric", "0100xx/", "invalid minutes"},

		// HHMMSSh - hours/minutes/seconds UTC
		{"h: hours 24", "240000h", "invalid hours"},
		{"h: hours non-numeric", "xx0000h", "invalid hours"},
		{"h: minutes 60", "006000h", "invalid minutes"},
		{"h: minutes non-numeric", "00xx00h", "invalid minutes"},
		{"h: seconds 60", "000060h", "invalid seconds"},
		{"h: seconds non-numeric", "0000xxh", "invalid seconds"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts, err := parseTimestamp(tc.input)
			if err == nil {
				t.Fatalf("parseTimestamp(%q) = %v, want error containing %q", tc.input, ts, tc.errIs)
			}
			if got := err.Error(); !strings.Contains(got, tc.errIs) {
				t.Errorf("parseTimestamp(%q) error = %q, want it to contain %q", tc.input, got, tc.errIs)
			}
		})
	}
}

func TestTimestampFutureRollback(t *testing.T) {
	// A timestamp with a day in the future should roll back to the previous month.
	// Pick a day guaranteed to be in the future: tomorrow, or wrap to 1 if > 28.
	now := time.Now()
	futureDay := now.Day() + 1
	if futureDay > 28 {
		futureDay = 1
	}
	// Use 23:59 to maximize the chance the constructed time is after now.
	tstamp := fmt.Sprintf("%02d2359", futureDay)

	tests := []struct {
		name   string
		packet string
		loc    *time.Location
	}{
		{
			name:   "DDHHMMz UTC",
			packet: "KB3HVP-14>APU25N,N8TJG-10*,WIDE2-1,qAR,LANSNG:@" + tstamp + "z4231.16N/08449.88Wu227/052/A=000941 {UIV32N}",
			loc:    time.UTC,
		},
		{
			name:   "DDHHMM/ local",
			packet: "G4EUM-9>APOTC1,G4EUM*,WIDE2-2,qAS,M3SXA-10:/" + tstamp + "/5134.38N/00019.47W>155/023!W26!/A=000188 14.3V 27C HDOP01.0 SATS09",
			loc:    now.Location(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			futureTime := time.Date(now.Year(), now.Month(), futureDay, 23, 59, 0, 0, tc.loc)
			if !futureTime.After(time.Now().In(tc.loc)) {
				t.Skip("cannot construct a future timestamp at this time of day")
			}

			outcome := futureTime.AddDate(0, -1, 0)

			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if p.Timestamp == nil {
				t.Fatalf("timestamp is nil, want %v", outcome)
			}
			if !p.Timestamp.Equal(outcome) {
				t.Errorf("timestamp = %v, want %v", *p.Timestamp, outcome)
			}
			// Verify it was indeed rolled back to the previous month
			tsInLoc := p.Timestamp.In(tc.loc)
			if tsInLoc.Month() == now.Month() && tsInLoc.Year() == now.Year() {
				t.Errorf("timestamp month = %v, expected previous month rollback", tsInLoc.Month())
			}
		})
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
