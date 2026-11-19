package fap

import (
	"fmt"
	"testing"
	"time"
)

// Tests ported from perl-aprs-fap/t/56decode-status.t

// TestParseCapabilities tests capabilities parsing.
func TestParseCapabilities(t *testing.T) {
	packet := "OH7LZB>APRS:<IGATE,MSG_CNT=5,LOC_CNT=10"
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse capabilities: %v", err)
	}

	if p.Type != PacketTypeCapabilities {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeCapabilities)
	}
	if v, ok := p.Capabilities["MSG_CNT"]; !ok || v != "5" {
		t.Errorf("MSG_CNT: got %q, want %q", v, "5")
	}
	if _, ok := p.Capabilities["IGATE"]; !ok {
		t.Error("IGATE capability not found")
	}
}

func TestStatusWithTimestamp(t *testing.T) {
	// Build a timestamp from the current time, matching the Perl test.
	// The status message's timestamp is not affected by the raw_timestamp flag.
	now := time.Now().UTC()
	tstamp := fmt.Sprintf("%02d%02d%02dz", now.Day(), now.Hour(), now.Minute())
	// Expected: rounded down to the minute
	outcome := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)

	msg := ">>Nashville,TN>>Toronto,ON"

	packet := "KB3HVP-14>APU25N,WIDE2-2,qAR,LANSNG:>" + tstamp + msg

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse a status message packet: %v", err)
	}

	if p.Type != PacketTypeStatus {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeStatus)
	}

	if p.Timestamp == nil {
		t.Fatalf("timestamp is nil, want %v", outcome)
	}
	if !p.Timestamp.Equal(outcome) {
		t.Errorf("timestamp = %v, want %v", *p.Timestamp, outcome)
	}

	if p.Status != msg {
		t.Errorf("status = %q, want %q", p.Status, msg)
	}
}

// TestStatusTimestampFormat tests that only a DDHHMMz timestamp is taken
// from the start of a status report; other text is status text, even if
// its 7th character is 'z' or '/'. The first two packets are from
// APRS-IS. The expected values are from the Perl module.
func TestStatusTimestampFormat(t *testing.T) {
	tests := []struct {
		name    string
		packet  string
		status  string
		hasTime bool
	}{
		{
			name:   "locator and slash",
			packet: "TEST-15>APWW11,TCPIP*,qAC,T2CSNGRAD:>CN85uj/# CN85uj Winlink Gateway 144.93",
			status: "CN85uj/# CN85uj Winlink Gateway 144.93",
		},
		{
			name:   "local time format",
			packet: "TEST-2>APRS,WIDE1-1,qAS,TEST-3:>191350/M/V -- NPY PH43",
			status: "191350/M/V -- NPY PH43",
		},
		{
			name:    "zulu time",
			packet:  "TEST>APRS:>191350zHello",
			status:  "Hello",
			hasTime: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if p.Status != tc.status {
				t.Errorf("status = %q, want %q", p.Status, tc.status)
			}
			if (p.Timestamp != nil) != tc.hasTime {
				t.Errorf("timestamp = %v, want set: %v", p.Timestamp, tc.hasTime)
			}
			if len(p.Warnings) > 0 {
				t.Errorf("warnings = %v, want none", p.Warnings)
			}
		})
	}
}
