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

// TestParseObject tests object parsing.
func TestParseObject(t *testing.T) {
	packet := "OH7LZB>APRS::;LEADER   *092345z4903.50N/07201.75W>"
	p, err := Parse(packet)

	// This packet has a message-type body (starts with :), the actual object packet would be:
	packet = "OH7LZB>APRS:;LEADER   *092345z4903.50N/07201.75W>comment"
	p, err = Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse object: %v", err)
	}

	if p.Type != PacketTypeObject {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeObject)
	}
	if p.ObjectName != "LEADER   " {
		t.Errorf("object name: got %q, want %q", p.ObjectName, "LEADER   ")
	}
	if p.Alive == nil || !*p.Alive {
		t.Error("expected object to be alive")
	}
}

// TestParseTelemetry tests telemetry parsing.
func TestParseTelemetry(t *testing.T) {
	packet := "OH7LZB>APRS:T#123,001,002,003,004,005,10101010"
	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse telemetry: %v", err)
	}

	if p.Type != PacketTypeTelemetry {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeTelemetry)
	}
	if p.TelemetryData == nil {
		t.Fatal("telemetry data is nil")
	}
	if p.TelemetryData.Seq != "123" {
		t.Errorf("seq: got %q, want %q", p.TelemetryData.Seq, "123")
	}
	if len(p.TelemetryData.Vals) < 5 {
		t.Fatalf("vals count: got %d, want 5", len(p.TelemetryData.Vals))
	}

	for i, want := range []float64{1, 2, 3, 4, 5} {
		if p.TelemetryData.Vals[i] == nil || *p.TelemetryData.Vals[i] != want {
			t.Errorf("val[%d]: got %v, want %f", i, p.TelemetryData.Vals[i], want)
		}
	}
}

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
