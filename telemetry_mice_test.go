package fap

import (
	"testing"
)

// Tests ported from perl-aprs-fap/t/54decode-tlm-mice.t

func TestMicEBase91Telemetry5Ch(t *testing.T) {
	// Sequence 00, 5 channels of telemetry and one channel of binary bits
	// Body: 'I',l \x1C>/ (mic-e encoded position)
	packet := "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH:'I',l \x1c>/ comment |!!!!!!!!!!!!!!|"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse mic-e with 5-ch telemetry: %v", err)
	}

	if p.Comment != "comment" {
		t.Errorf("comment = %q, want %q", p.Comment, "comment")
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}

	if tlm.Seq != "0" {
		t.Errorf("seq = %q, want %q", tlm.Seq, "0")
	}

	if len(tlm.Vals) < 5 {
		t.Fatalf("vals length = %d, want >= 5", len(tlm.Vals))
	}
	for i := 0; i < 5; i++ {
		if tlm.Vals[i] == nil {
			t.Errorf("vals[%d] = nil, want 0", i)
		} else if *tlm.Vals[i] != 0 {
			t.Errorf("vals[%d] = %v, want 0", i, *tlm.Vals[i])
		}
	}

	if tlm.Bits != "00000000" {
		t.Errorf("bits = %q, want %q", tlm.Bits, "00000000")
	}
}

func TestMicEBase91Telemetry1Ch(t *testing.T) {
	// Sequence 00, 1 channel of telemetry
	packet := "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH:'I',l \x1c>/ comment |!!!!|"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse mic-e with 1-ch telemetry: %v", err)
	}

	if p.Comment != "comment" {
		t.Errorf("comment = %q, want %q", p.Comment, "comment")
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}

	if tlm.Seq != "0" {
		t.Errorf("seq = %q, want %q", tlm.Seq, "0")
	}

	if len(tlm.Vals) < 5 {
		t.Fatalf("vals length = %d, want >= 5", len(tlm.Vals))
	}
	if tlm.Vals[0] == nil || *tlm.Vals[0] != 0 {
		t.Errorf("vals[0] = %v, want 0", tlm.Vals[0])
	}
	for i := 1; i <= 4; i++ {
		if tlm.Vals[i] != nil {
			t.Errorf("vals[%d] = %v, want nil", i, *tlm.Vals[i])
		}
	}
}

func TestMicEBase91TelemetryHarder(t *testing.T) {
	// Harder packet with base-91 telemetry
	packet := "N6BG-1>S6QTUX:`+,^l!cR/'\";z}||ss11223344bb!\"|!w>f!|3"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse harder mic-e with telemetry: %v", err)
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}

	if tlm.Bits != "10000000" {
		t.Errorf("bits = %q, want %q", tlm.Bits, "10000000")
	}
}

func TestMicEBase91TelemetryDAOConfusing(t *testing.T) {
	// Telemetry that looks like it could be a DAO extension: |!wEU!![S|
	packet := "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH:'I',l \x1c>/ comment |!wEU!![S|"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse mic-e with DAO-confusing telemetry: %v", err)
	}

	if p.Comment != "comment" {
		t.Errorf("comment = %q, want %q", p.Comment, "comment")
	}

	if p.TelemetryData == nil {
		t.Fatal("no telemetry data - telemetry was probably confused with DAO")
	}
}
