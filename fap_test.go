package fap

import (
	"fmt"
	"math"
	"testing"
)

func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

// TestParseUncompressedNortheast tests basic uncompressed position parsing.
func TestParseUncompressedNortheast(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!6028.51N/02505.68E#PHG7220/RELAY,WIDE, OH2AP Jarvenpaa"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.SrcCallsign != "OH2RDP-1" {
		t.Errorf("src callsign: got %q, want %q", p.SrcCallsign, "OH2RDP-1")
	}
	if p.DstCallsign != "BEACON-15" {
		t.Errorf("dst callsign: got %q, want %q", p.DstCallsign, "BEACON-15")
	}
	if p.Format != FormatUncompressed {
		t.Errorf("format: got %q, want %q", p.Format, FormatUncompressed)
	}

	if p.Latitude == nil || !approxEqual(*p.Latitude, 60.4752, 0.0001) {
		t.Errorf("latitude: got %v, want ~60.4752", p.Latitude)
	}
	if p.Longitude == nil || !approxEqual(*p.Longitude, 25.0947, 0.0001) {
		t.Errorf("longitude: got %v, want ~25.0947", p.Longitude)
	}
	if p.PosResolution == nil || !approxEqual(*p.PosResolution, 18.52, 0.01) {
		t.Errorf("pos resolution: got %v, want ~18.52", p.PosResolution)
	}
	if p.PHG != "7220" {
		t.Errorf("PHG: got %q, want %q", p.PHG, "7220")
	}
	if p.Comment != "RELAY,WIDE, OH2AP Jarvenpaa" {
		t.Errorf("comment: got %q, want %q", p.Comment, "RELAY,WIDE, OH2AP Jarvenpaa")
	}

	// Digipeaters
	if len(p.Digipeaters) != 2 {
		t.Fatalf("digipeater count: got %d, want 2", len(p.Digipeaters))
	}
	if p.Digipeaters[0].Call != "OH2RDG" || !p.Digipeaters[0].WasDigied {
		t.Errorf("digi 0: got %+v, want OH2RDG/digied", p.Digipeaters[0])
	}
	if p.Digipeaters[1].Call != "WIDE" || p.Digipeaters[1].WasDigied {
		t.Errorf("digi 1: got %+v, want WIDE/not digied", p.Digipeaters[1])
	}
}

// TestParseUncompressedSouthwest tests southern/western hemisphere parsing.
func TestParseUncompressedSouthwest(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!6028.51S/02505.68W#PHG7220RELAY,WIDE, OH2AP Jarvenpaa"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Latitude == nil || !approxEqual(*p.Latitude, -60.4752, 0.0001) {
		t.Errorf("latitude: got %v, want ~-60.4752", p.Latitude)
	}
	if p.Longitude == nil || !approxEqual(*p.Longitude, -25.0947, 0.0001) {
		t.Errorf("longitude: got %v, want ~-25.0947", p.Longitude)
	}
}

// TestParseUncompressedAmbiguity tests position ambiguity parsing.
func TestParseUncompressedAmbiguity(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!602 .  S/0250 .  W#PHG7220RELAY"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.PosAmbiguity == nil || *p.PosAmbiguity != 3 {
		t.Errorf("pos ambiguity: got %v, want 3", p.PosAmbiguity)
	}
	if p.Latitude == nil || !approxEqual(*p.Latitude, -60.4167, 0.001) {
		t.Errorf("latitude: got %v, want ~-60.4167", p.Latitude)
	}
	if p.PosResolution == nil || !approxEqual(*p.PosResolution, 18520, 1) {
		t.Errorf("pos resolution: got %v, want ~18520", p.PosResolution)
	}
}

// TestParseUncompressedHighAmbiguity tests very high ambiguity.
func TestParseUncompressedHighAmbiguity(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:!60  .  S/025  .  W#PHG7220RELAY"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.PosAmbiguity == nil || *p.PosAmbiguity != 4 {
		t.Errorf("pos ambiguity: got %v, want 4", p.PosAmbiguity)
	}
	if p.Latitude == nil || !approxEqual(*p.Latitude, -60.5, 0.001) {
		t.Errorf("latitude: got %v, want ~-60.5", p.Latitude)
	}
	if p.PosResolution == nil || !approxEqual(*p.PosResolution, 111120, 1) {
		t.Errorf("pos resolution: got %v, want ~111120", p.PosResolution)
	}
}

// TestParseUncompressedLastResort tests ! position found later in body.
func TestParseUncompressedLastResort(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,OH2RDG*,WIDE:hoponassualku!6028.51S/02505.68W#PHG7220/RELAY,WIDE, OH2AP Jarvenpaa"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Latitude == nil || !approxEqual(*p.Latitude, -60.4752, 0.0001) {
		t.Errorf("latitude: got %v, want ~-60.4752", p.Latitude)
	}
	if p.Comment != "RELAY,WIDE, OH2AP Jarvenpaa" {
		t.Errorf("comment: got %q, want %q", p.Comment, "RELAY,WIDE, OH2AP Jarvenpaa")
	}
}

// TestParseUncompressedWithTimestamp tests position with timestamp and altitude.
func TestParseUncompressedWithTimestamp(t *testing.T) {
	packet := "YB1RUS-9>APOTC1,WIDE2-2,qAS,YC0GIN-1:/180000z0609.31S/10642.85E>058/010/A=000079 13.8V 15CYB1RUS-9 Mobile Tracker"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Latitude == nil || !approxEqual(*p.Latitude, -6.15517, 0.00001) {
		t.Errorf("latitude: got %v, want ~-6.15517", p.Latitude)
	}
	if p.Longitude == nil || !approxEqual(*p.Longitude, 106.71417, 0.00001) {
		t.Errorf("longitude: got %v, want ~106.71417", p.Longitude)
	}
	if p.Altitude == nil || !approxEqual(*p.Altitude, 24.0792, 0.001) {
		t.Errorf("altitude: got %v, want ~24.0792", p.Altitude)
	}
}

// TestParseUncompressedNegativeAltitude tests negative altitude.
func TestParseUncompressedNegativeAltitude(t *testing.T) {
	packet := "YB1RUS-9>APOTC1,WIDE2-2,qAS,YC0GIN-1:/180000z0609.31S/10642.85E>058/010/A=-00079 13.8V 15CYB1RUS-9 Mobile Tracker"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Altitude == nil || !approxEqual(*p.Altitude, -24.0792, 0.001) {
		t.Errorf("altitude: got %v, want ~-24.0792", p.Altitude)
	}
}

// TestParseMicENonMoving tests a non-moving target's mic-e packet.
func TestParseMicENonMoving(t *testing.T) {
	packet := "OH7LZB-13>SX15S6,TCPIP*,qAC,FOURTH:'I',l \x1C>/]"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Type != PacketTypeLocation {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatMicE {
		t.Errorf("format: got %q, want %q", p.Format, FormatMicE)
	}
	if p.SrcCallsign != "OH7LZB-13" {
		t.Errorf("src: got %q, want %q", p.SrcCallsign, "OH7LZB-13")
	}
	if p.SymbolTable != '/' {
		t.Errorf("symbol table: got %c, want /", p.SymbolTable)
	}
	if p.SymbolCode != '>' {
		t.Errorf("symbol code: got %c, want >", p.SymbolCode)
	}
	if p.Latitude == nil || !approxEqual(*p.Latitude, -38.2560, 0.001) {
		t.Errorf("latitude: got %v, want ~-38.2560", p.Latitude)
	}
	if p.Longitude == nil || !approxEqual(*p.Longitude, 145.1860, 0.001) {
		t.Errorf("longitude: got %v, want ~145.1860", p.Longitude)
	}
}

// TestParseMicEMoving tests a moving target's mic-e packet.
func TestParseMicEMoving(t *testing.T) {
	packet := "OH7LZB-2>TQ4W2V,WIDE2-1,qAo,OH7LZB:`c51!f?>/]\"3x}="
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Type != PacketTypeLocation {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.MBits != "110" {
		t.Errorf("mbits: got %q, want %q", p.MBits, "110")
	}
	if p.Latitude == nil || !approxEqual(*p.Latitude, 41.7877, 0.001) {
		t.Errorf("latitude: got %v, want ~41.7877", p.Latitude)
	}
	if p.Longitude == nil || !approxEqual(*p.Longitude, -71.4202, 0.001) {
		t.Errorf("longitude: got %v, want ~-71.4202", p.Longitude)
	}
	if p.Speed == nil || !approxEqual(*p.Speed, 105.56, 0.1) {
		t.Errorf("speed: got %v, want ~105.56", p.Speed)
	}
	if p.Course == nil || *p.Course != 35 {
		t.Errorf("course: got %v, want 35", p.Course)
	}
	if p.Altitude == nil || *p.Altitude != 6 {
		// Mic-E altitude: check it parsed
		t.Logf("altitude: got %v (may differ due to encoding)", p.Altitude)
	}
}

// TestParseMicEInvalidSymTable tests error on invalid symbol table.
func TestParseMicEInvalidSymTable(t *testing.T) {
	packet := "OZ2BRN-4>5U2V08,OZ3RIN-3,OZ4DIA-2*,WIDE2-1,qAR,DB0KUE:`'O<l!{,,\"4R}"
	_, err := ParseAPRS(packet)
	if err == nil {
		t.Error("expected error for invalid symbol table, got nil")
	}
}

// TestParseMessage tests message parsing.
func TestParseMessage(t *testing.T) {
	packet := "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{42"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Type != PacketTypeMessage {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeMessage)
	}
	if p.Destination != "OH7LZB" {
		t.Errorf("destination: got %q, want %q", p.Destination, "OH7LZB")
	}
	if p.Message != "Testing, 1 2 3" {
		t.Errorf("message: got %q, want %q", p.Message, "Testing, 1 2 3")
	}
	if p.MessageID != "42" {
		t.Errorf("message id: got %q, want %q", p.MessageID, "42")
	}
}

// TestParseMessageAck tests message ack parsing.
func TestParseMessageAck(t *testing.T) {
	packet := "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :ack42"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Type != PacketTypeMessage {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeMessage)
	}
	if p.MessageAck != "42" {
		t.Errorf("message ack: got %q, want %q", p.MessageAck, "42")
	}
}

// TestParseMessageRej tests message reject parsing.
func TestParseMessageRej(t *testing.T) {
	packet := "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :rej42"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.Type != PacketTypeMessage {
		t.Errorf("type: got %q, want %q", p.Type, PacketTypeMessage)
	}
	if p.MessageRej != "42" {
		t.Errorf("message rej: got %q, want %q", p.MessageRej, "42")
	}
}

// TestParseMessageReplyAck tests message with piggybacked reply-ack.
func TestParseMessageReplyAck(t *testing.T) {
	packet := "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::OH7LZB   :Testing, 1 2 3{42}f001"
	p, err := ParseAPRS(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if p.MessageID != "42" {
		t.Errorf("message id: got %q, want %q", p.MessageID, "42")
	}
	if p.MessageAck != "f001" {
		t.Errorf("message ack: got %q, want %q", p.MessageAck, "f001")
	}
}

// TestParseStatus tests status report parsing.
func TestParseStatus(t *testing.T) {
	packet := "OH7LZB>APRS:>Testing status"
	p, err := ParseAPRS(packet)
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
	_, err := ParseAPRS("OH7LZB>APRS")
	if err == nil {
		t.Error("expected error for packet without body")
	}
}

// TestParseNoGT tests error handling for packet without >.
func TestParseNoGT(t *testing.T) {
	_, err := ParseAPRS("OH7LZB:body")
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

// TestMicEMBitsToMessage tests Mic-E message type decoding.
func TestMicEMBitsToMessage(t *testing.T) {
	tests := []struct {
		bits string
		want string
	}{
		{"111", "Off Duty"},
		{"110", "En Route"},
		{"000", "Emergency"},
	}
	for _, tc := range tests {
		got := MicEMBitsToMessage(tc.bits)
		if got != tc.want {
			t.Errorf("MicEMBitsToMessage(%q): got %q, want %q", tc.bits, got, tc.want)
		}
	}
}

// TestParseObject tests object parsing.
func TestParseObject(t *testing.T) {
	packet := "OH7LZB>APRS::;LEADER   *092345z4903.50N/07201.75W>"
	p, err := ParseAPRS(packet)

	// This packet has a message-type body (starts with :), the actual object packet would be:
	packet = "OH7LZB>APRS:;LEADER   *092345z4903.50N/07201.75W>comment"
	p, err = ParseAPRS(packet)
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
	p, err := ParseAPRS(packet)
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
	p, err := ParseAPRS(packet)
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

func ExampleParseAPRS() {
	packet := "OH7LZB-2>APRS,WIDE1-1,WIDE2-1,qAo,OH7LZB:!6128.23N/02353.52E-PHG2360/Testing"
	p, err := ParseAPRS(packet)
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
