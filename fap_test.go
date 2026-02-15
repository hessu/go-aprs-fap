package fap

import (
	"fmt"
	"math"
	"testing"
)

func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

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

// TestDistance tests great-circle distance calculation.
func TestDistance(t *testing.T) {
	// Helsinki to Tampere, approximately 161 km
	d := Distance(60.17, 24.94, 61.50, 23.78)
	if d < 155 || d > 165 {
		t.Errorf("distance Helsinki-Tampere: got %.1f, want ~161", d)
	}
}

// TestDirection tests bearing calculation.
func TestDirection(t *testing.T) {
	// Due north
	d := Direction(60.0, 25.0, 61.0, 25.0)
	if d < 355 && d > 5 {
		t.Errorf("direction north: got %.1f, want ~0", d)
	}

	// Due east
	d = Direction(60.0, 25.0, 60.0, 26.0)
	if d < 85 || d > 95 {
		t.Errorf("direction east: got %.1f, want ~90", d)
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
