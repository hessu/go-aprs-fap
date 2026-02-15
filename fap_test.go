package fap

import (
	"fmt"
	"testing"
)

// TestParseStatus tests status report parsing.
func TestParseStatus(t *testing.T) {
	packet := "OH7LZB>APRS:>Testing status"
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Type != PacketTypeStatus {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeStatus)
	}
	if p.Status != "Testing status" {
		t.Errorf("status: got %q, want %q", p.Status, "Testing status")
	}
}

// TestParseNoBody tests error handling for packet without body.
func TestParseNoBody(t *testing.T) {
	_, err := Parse("OH7LZB>APRS")
	if err == nil {
		t.Error("expected error for packet without body")
	}
}

// TestParseNoGT tests error handling for packet without >.
func TestParseNoGT(t *testing.T) {
	_, err := Parse("OH7LZB:body")
	if err == nil {
		t.Error("expected error for packet without >")
	}
}

func ExampleParse() {
	packet := "OH7LZB-2>APRS,WIDE1-1,WIDE2-1,qAo,OH7LZB:!6128.23N/02353.52E-PHG2360/Testing"
	p, err := Parse(packet)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Source: %s\n", p.SrcCallsign)
	fmt.Printf("Type: %s\n", p.Type)
	fmt.Printf("Lat: %.4f\n", *p.Latitude)
	fmt.Printf("Lon: %.4f\n", *p.Longitude)
	// Output:
	// Source: OH7LZB-2
	// Type: location
	// Lat: 61.4705
	// Lon: 23.8920
}
