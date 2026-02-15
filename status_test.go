package fap

import (
	"fmt"
	"testing"
	"time"
)

// Tests ported from perl-aprs-fap/t/56decode-status.t

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
