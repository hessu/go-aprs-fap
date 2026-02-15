package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/22decode-compressed.t

func TestCompressedNonMoving(t *testing.T) {
	header := "OH2KKU-15>APRS,TCPIP*,qAC,FOURTH"
	body := "!I0-X;T_Wv&{-Aigate testing"
	packet := header + ":" + body

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}

	if p.SrcCallsign != "OH2KKU-15" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2KKU-15")
	}
	if p.DstCallsign != "APRS" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "APRS")
	}
	if p.Header != header {
		t.Errorf("header = %q, want %q", p.Header, header)
	}
	if p.Body != body {
		t.Errorf("body = %q, want %q", p.Body, body)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Format != FormatCompressed {
		t.Errorf("format = %q, want %q", p.Format, FormatCompressed)
	}
	if p.Comment != "igate testing" {
		t.Errorf("comment = %q, want %q", p.Comment, "igate testing")
	}

	// Digipeaters
	if len(p.Digipeaters) != 3 {
		t.Fatalf("digipeaters count = %d, want 3", len(p.Digipeaters))
	}
	wantDigis := []struct {
		call      string
		wasDigied bool
	}{
		{"TCPIP", true},
		{"qAC", false},
		{"FOURTH", false},
	}
	for i, want := range wantDigis {
		if p.Digipeaters[i].Call != want.call {
			t.Errorf("digi[%d].call = %q, want %q", i, p.Digipeaters[i].Call, want.call)
		}
		if p.Digipeaters[i].WasDigied != want.wasDigied {
			t.Errorf("digi[%d].wasdigied = %v, want %v", i, p.Digipeaters[i].WasDigied, want.wasDigied)
		}
	}

	if p.SymbolTable != 'I' {
		t.Errorf("symboltable = %c, want I", p.SymbolTable)
	}
	if p.SymbolCode != '&' {
		t.Errorf("symbolcode = %c, want &", p.SymbolCode)
	}

	// Compressed positions don't set posambiguity
	if p.PosAmbiguity != nil {
		t.Errorf("posambiguity = %v, want nil", *p.PosAmbiguity)
	}
	if p.Messaging == nil || *p.Messaging != false {
		t.Errorf("messaging = %v, want false", p.Messaging)
	}

	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.0520" {
		t.Errorf("latitude = %s, want 60.0520", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "24.5045" {
		t.Errorf("longitude = %s, want 24.5045", got)
	}
	if got := fmt.Sprintf("%.3f", *p.PosResolution); got != "0.291" {
		t.Errorf("posresolution = %s, want 0.291", got)
	}

	// No speed, course, altitude for this packet
	if p.Speed != nil {
		t.Errorf("speed = %v, want nil", *p.Speed)
	}
	if p.Course != nil {
		t.Errorf("course = %v, want nil", *p.Course)
	}
	if p.Altitude != nil {
		t.Errorf("altitude = %v, want nil", *p.Altitude)
	}
}

func TestCompressedMoving(t *testing.T) {
	header := "OH2LCQ-10>APZMDR,WIDE3-2,qAo,OH2MQK-1"
	comment := "Tero, Green Volvo 960, GGL-880"
	// Inline telemetry in the comment
	body := "!//zPHTfVv>!V_ " + comment + "|!!!!!!!!!!!!!!|"
	packet := header + ":" + body

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}

	if p.SrcCallsign != "OH2LCQ-10" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2LCQ-10")
	}
	if p.DstCallsign != "APZMDR" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "APZMDR")
	}
	if p.Header != header {
		t.Errorf("header = %q, want %q", p.Header, header)
	}
	if p.Body != body {
		t.Errorf("body = %q, want %q", p.Body, body)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeLocation)
	}
	if p.Comment != comment {
		t.Errorf("comment = %q, want %q", p.Comment, comment)
	}

	// Digipeaters
	if len(p.Digipeaters) != 3 {
		t.Fatalf("digipeaters count = %d, want 3", len(p.Digipeaters))
	}
	wantDigis := []struct {
		call      string
		wasDigied bool
	}{
		{"WIDE3-2", false},
		{"qAo", false},
		{"OH2MQK-1", false},
	}
	for i, want := range wantDigis {
		if p.Digipeaters[i].Call != want.call {
			t.Errorf("digi[%d].call = %q, want %q", i, p.Digipeaters[i].Call, want.call)
		}
		if p.Digipeaters[i].WasDigied != want.wasDigied {
			t.Errorf("digi[%d].wasdigied = %v, want %v", i, p.Digipeaters[i].WasDigied, want.wasDigied)
		}
	}

	if p.SymbolTable != '/' {
		t.Errorf("symboltable = %c, want /", p.SymbolTable)
	}
	if p.SymbolCode != '>' {
		t.Errorf("symbolcode = %c, want >", p.SymbolCode)
	}

	if p.PosAmbiguity != nil {
		t.Errorf("posambiguity = %v, want nil", *p.PosAmbiguity)
	}
	if p.Messaging == nil || *p.Messaging != false {
		t.Errorf("messaging = %v, want false", p.Messaging)
	}

	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.3582" {
		t.Errorf("latitude = %s, want 60.3582", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "24.8084" {
		t.Errorf("longitude = %s, want 24.8084", got)
	}
	if got := fmt.Sprintf("%.3f", *p.PosResolution); got != "0.291" {
		t.Errorf("posresolution = %s, want 0.291", got)
	}

	if p.Speed == nil {
		t.Fatal("speed is nil")
	}
	if got := fmt.Sprintf("%.2f", *p.Speed); got != "107.57" {
		t.Errorf("speed = %s, want 107.57", got)
	}
	if p.Course == nil || *p.Course != 360 {
		t.Errorf("course = %v, want 360", p.Course)
	}
	if p.Altitude != nil {
		t.Errorf("altitude = %v, want nil", *p.Altitude)
	}
}

func TestCompressedTooShort(t *testing.T) {
	// Short compressed packet without speed, altitude or course.
	// The APRS 1.01 spec says compressed packet is always 13 bytes long.
	// Must not decode, even though this packet is otherwise valid.
	packet := "KJ4ERJ-AL>APWW05,TCPIP*,qAC,FOURTH:@075111h/@@.Y:*lol "

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for too-short compressed packet")
	}
}

func TestCompressedWeather(t *testing.T) {
	// Compressed packet with weather data
	packet := "SV4IKL-2>APU25N,WIDE2-2,qAR,SV6EXB-1:@011444z/:JF!T/W-_e!bg001t054r000p010P010h65b10073WS 2300 {UIV32N}"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.ResultCode != "" {
		t.Fatalf("unexpected result code: %s", p.ResultCode)
	}

	if p.SymbolTable != '/' {
		t.Errorf("symboltable = %c, want /", p.SymbolTable)
	}
	if p.SymbolCode != '_' {
		t.Errorf("symbolcode = %c, want _", p.SymbolCode)
	}
	if p.Comment != "WS 2300 {UIV32N}" {
		t.Errorf("comment = %q, want %q", p.Comment, "WS 2300 {UIV32N}")
	}

	if p.Wx == nil {
		t.Fatal("wx is nil")
	}
	// wind_gust: 1 mph * 0.44704 = 0.44704, sprintf('%.1f') = "0.4"
	if p.Wx.WindGust == nil {
		t.Fatal("wind_gust is nil")
	}
	if got := fmt.Sprintf("%.1f", *p.Wx.WindGust); got != "0.4" {
		t.Errorf("wind_gust = %s, want 0.4", got)
	}
	// temp: (54-32)*5/9 = 12.2
	if p.Wx.Temp == nil {
		t.Fatal("temp is nil")
	}
	if got := fmt.Sprintf("%.1f", *p.Wx.Temp); got != "12.2" {
		t.Errorf("temp = %s, want 12.2", got)
	}
	// humidity: 65
	if p.Wx.Humidity == nil || *p.Wx.Humidity != 65 {
		t.Errorf("humidity = %v, want 65", p.Wx.Humidity)
	}
	// pressure: 10073/10 = 1007.3
	if p.Wx.Pressure == nil {
		t.Fatal("pressure is nil")
	}
	if got := fmt.Sprintf("%.1f", *p.Wx.Pressure); got != "1007.3" {
		t.Errorf("pressure = %s, want 1007.3", got)
	}
}

func TestCompressedWeatherSpaceGust(t *testing.T) {
	// Compressed packet with weather, space in wind gust field
	packet := "SV4IKL-2>APU25N,WIDE2-2,qAR,SV6EXB-1:@011444z/:JF!T/W-_e!bg   t054r000p010P010h65b10073WS 2300 {UIV32N}"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if p.Wx == nil {
		t.Fatal("wx is nil")
	}
	// temp should still parse correctly
	if p.Wx.Temp == nil {
		t.Fatal("temp is nil")
	}
	if got := fmt.Sprintf("%.1f", *p.Wx.Temp); got != "12.2" {
		t.Errorf("temp = %s, want 12.2", got)
	}
}
