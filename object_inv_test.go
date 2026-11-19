package fap

import (
	"errors"
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/40decode-object-inv.t

func TestObjectInvalidBroken(t *testing.T) {
	// Broken object: binary characters destroyed in cut'n'paste,
	// only one space between "HQ" and "*" instead of two,
	// so the alive/killed indicator lands in the wrong position.
	packet := "OH2KKU-1>APRS,TCPIP*,qAC,FIRST:;SRAL HQ *110507zS0%E/Th4_a AKaupinmaenpolku9,open M-Th12-17,F12-14 lcl"

	p, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for broken object packet")
	}
	if !errors.Is(err, ErrObjInvalid) {
		t.Errorf("error = %v, want %v", err, ErrObjInvalid)
	}
	if p.Type != PacketTypeObject {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeObject)
	}
}

// TestObjectInvalidHeader tests that an object is invalid if the name has
// non-printable characters or the timestamp does not have the format of
// a timestamp, but an invalid time in a timestamp of the right format is
// only a warning. The first packet is from APRS-IS. The expected values
// are from the Perl module.
func TestObjectInvalidHeader(t *testing.T) {
	const pos = "4851.65N/12345.41Wr145.470MHz T127"
	tests := []struct {
		name    string
		packet  string
		wantErr bool
		alive   bool
		hasTime bool
	}{
		{name: "uppercase Z", packet: "TEST>BEACON,qAR,VE7VIC-4:;145.470-T*111111Z" + pos, wantErr: true},
		{name: "control character", packet: "TEST>BEACON,qAR,VE7VIC-4:;145.47\x01-T*111111z" + pos, wantErr: true},
		{name: "invalid time", packet: "TEST>BEACON,qAR,VE7VIC-4:;145.470-T*999999z" + pos, alive: true},
		{name: "valid", packet: "TEST>BEACON,qAR,VE7VIC-4:;145.470-T_111111z" + pos, hasTime: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if tc.wantErr {
				if !errors.Is(err, ErrObjInvalid) {
					t.Fatalf("error = %v, want %v", err, ErrObjInvalid)
				}
				if p.ObjectName != "" {
					t.Errorf("objectname = %q on an invalid object", p.ObjectName)
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if p.ObjectName != "145.470-T" {
				t.Errorf("objectname = %q, want %q", p.ObjectName, "145.470-T")
			}
			if p.Alive == nil || *p.Alive != tc.alive {
				t.Errorf("alive = %v, want %v", p.Alive, tc.alive)
			}
			if (p.Timestamp != nil) != tc.hasTime {
				t.Errorf("timestamp = %v, want set: %v", p.Timestamp, tc.hasTime)
			}
			if got := fmt.Sprintf("%.4f", *p.Latitude); got != "48.8608" {
				t.Errorf("latitude = %s, want 48.8608", got)
			}
		})
	}
}
