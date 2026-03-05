package fap

import (
	"testing"
)

// Tests ported from perl-aprs-fap/t/52decode-beacon.t

func TestBeacon(t *testing.T) {
	t.Run("non-APRS", func(t *testing.T) {
		// Non-APRS beacon packet — body starts with space, not a recognized type.
		// Should parse successfully with PacketTypeBeacon.
		p, err := Parse("OH2RDU>UIDIGI: UIDIGI 1.9")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type != PacketTypeBeacon {
			t.Errorf("type = %q, want %q", p.Type, PacketTypeBeacon)
		}
		if p.SrcCallsign != "OH2RDU" {
			t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2RDU")
		}
		if p.DstCallsign != "UIDIGI" {
			t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "UIDIGI")
		}
		if p.Body != " UIDIGI 1.9" {
			t.Errorf("body = %q, want %q", p.Body, " UIDIGI 1.9")
		}
	})

	t.Run("all beacon destinations", func(t *testing.T) {
		for _, dst := range []string{"ID", "BEACON", "UIDIGI", "CQ"} {
			t.Run(dst, func(t *testing.T) {
				p, err := Parse("OH2RDU>" + dst + ":hello world")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if p.Type != PacketTypeBeacon {
					t.Errorf("type = %q, want %q", p.Type, PacketTypeBeacon)
				}
			})
		}
	})

	t.Run("beacon with position fallback", func(t *testing.T) {
		// Beacon destination but body contains a position after '!'.
		// Position fallback should succeed, producing a location packet.
		p, err := Parse("OH2RDU>BEACON:testing!6028.51N/02505.68E#PHG2360")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type != PacketTypeLocation {
			t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
		}
		if p.Latitude == nil {
			t.Fatal("latitude is nil, want ~60.475")
		}
		if *p.Latitude < 60.47 || *p.Latitude > 60.48 {
			t.Errorf("latitude = %f, want ~60.475", *p.Latitude)
		}
	})
}
