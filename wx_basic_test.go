package fap

// Tests ported from perl-aprs-fap/t/30decode-wx-basic.t

import (
	"fmt"
	"testing"
)

func TestWxBasic(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,WIDE2-1,qAo,OH2MQK-1:=6030.35N/02443.91E_150/002g004t039r001P002p004h00b10125XRSW"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse a basic wx packet: %v", err)
	}

	if p.SrcCallsign != "OH2RDP-1" {
		t.Errorf("srccallsign = %q, want %q", p.SrcCallsign, "OH2RDP-1")
	}
	if p.DstCallsign != "BEACON-15" {
		t.Errorf("dstcallsign = %q, want %q", p.DstCallsign, "BEACON-15")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.5058" {
		t.Errorf("latitude = %s, want 60.5058", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "24.7318" {
		t.Errorf("longitude = %s, want 24.7318", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 150 {
		t.Errorf("wind_direction = %v, want 150", wx.WindDirection)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != "0.9" {
		t.Errorf("wind_speed = %s, want 0.9", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "1.8" {
		t.Errorf("wind_gust = %s, want 1.8", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "3.9" {
		t.Errorf("temp = %s, want 3.9", got)
	}
	if wx.Humidity == nil || *wx.Humidity != 100 {
		t.Errorf("humidity = %v, want 100", wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "1012.5" {
		t.Errorf("pressure = %s, want 1012.5", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Rain24h); got != "1.0" {
		t.Errorf("rain_24h = %s, want 1.0", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.Rain1h); got != "0.3" {
		t.Errorf("rain_1h = %s, want 0.3", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "0.5" {
		t.Errorf("rain_midnight = %s, want 0.5", got)
	}

	if wx.Soft != "XRSW" {
		t.Errorf("soft = %q, want %q", wx.Soft, "XRSW")
	}
}

func TestWxWithComment(t *testing.T) {
	packet := "OH2GAX>APU25N,TCPIP*,qAC,OH2GAX:@101317z6024.78N/02503.97E_156/001g005t038r000p000P000h91b10093/type ?sade for more wx info"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse second basic wx packet: %v", err)
	}

	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "60.4130" {
		t.Errorf("latitude = %s, want 60.4130", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "25.0662" {
		t.Errorf("longitude = %s, want 25.0662", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}
	if p.Comment != "/type ?sade for more wx info" {
		t.Errorf("comment = %q, want %q", p.Comment, "/type ?sade for more wx info")
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 156 {
		t.Errorf("wind_direction = %v, want 156", wx.WindDirection)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != "0.4" {
		t.Errorf("wind_speed = %s, want 0.4", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "2.2" {
		t.Errorf("wind_gust = %s, want 2.2", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "3.3" {
		t.Errorf("temp = %s, want 3.3", got)
	}
	if wx.Humidity == nil || *wx.Humidity != 91 {
		t.Errorf("humidity = %v, want 91", wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "1009.3" {
		t.Errorf("pressure = %s, want 1009.3", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Rain24h); got != "0.0" {
		t.Errorf("rain_24h = %s, want 0.0", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.Rain1h); got != "0.0" {
		t.Errorf("rain_1h = %s, want 0.0", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "0.0" {
		t.Errorf("rain_midnight = %s, want 0.0", got)
	}
}

func TestWxThirdWithComment(t *testing.T) {
	packet := "JH9YVX>APU25N,TCPIP*,qAC,T2TOKYO3:@011241z3558.58N/13629.67E_068/001g001t033r000p020P020b09860h98Oregon WMR100N Weather Station {UIV32N}"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse third basic wx packet: %v", err)
	}

	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "35.9763" {
		t.Errorf("latitude = %s, want 35.9763", got)
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "136.4945" {
		t.Errorf("longitude = %s, want 136.4945", got)
	}
	if got := fmt.Sprintf("%.2f", *p.PosResolution); got != "18.52" {
		t.Errorf("posresolution = %s, want 18.52", got)
	}
	if p.Comment != "Oregon WMR100N Weather Station {UIV32N}" {
		t.Errorf("comment = %q, want %q", p.Comment, "Oregon WMR100N Weather Station {UIV32N}")
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 68 {
		t.Errorf("wind_direction = %v, want 68", wx.WindDirection)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != "0.4" {
		t.Errorf("wind_speed = %s, want 0.4", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "0.4" {
		t.Errorf("wind_gust = %s, want 0.4", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "0.6" {
		t.Errorf("temp = %s, want 0.6", got)
	}
	if wx.Humidity == nil || *wx.Humidity != 98 {
		t.Errorf("humidity = %v, want 98", wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "986.0" {
		t.Errorf("pressure = %s, want 986.0", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Rain1h); got != "0.0" {
		t.Errorf("rain_1h = %s, want 0.0", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.Rain24h); got != "5.1" {
		t.Errorf("rain_24h = %s, want 5.1", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "5.1" {
		t.Errorf("rain_midnight = %s, want 5.1", got)
	}
}

func TestWxNoWindDirectionCourse(t *testing.T) {
	packet := "N0CALL>APU25N,TCPIP*,qAC,T2TOKYO3:@011241z3558.58N/13629.67E_.../...g001t033r000p020P020b09860h98Oregon WMR100N Weather Station {UIV32N}"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse wx packet without wind direction/course: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "0.4" {
		t.Errorf("wind_gust = %s, want 0.4", got)
	}
}

func TestWxNoWindNoTemp(t *testing.T) {
	packet := "N0CALL>APJLSX,TCPIP*,qAS,KG4EXY:@061750z3849.10N/07725.10W_.../...g...t...r008p011P011b.....h.."

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse wx packet without wind, gust or temperature: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if got := fmt.Sprintf("%.1f", *wx.Rain1h); got != "2.0" {
		t.Errorf("rain_1h = %s, want 2.0", got)
	}
}

func TestWxSpaceInWindGust(t *testing.T) {
	packet := "N0CALL>APU25N,TCPIP*,qAC,T2TOKYO3:@011241z3558.58N/13629.67E_.../...g   t033r000p020P020b09860h98Oregon WMR100N Weather Station {UIV32N}"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse wx packet with spaces in wind gust: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "0.6" {
		t.Errorf("temp = %s, want 0.6", got)
	}
}

func TestWxPositionlessWithSnowfall(t *testing.T) {
	packet := "JH9YVX>APU25N,TCPIP*,qAC,T2TOKYO3:_12032359c180s001g002t033r010p040P080b09860h98Os010L500"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse positionless wx packet: %v", err)
	}

	if p.Latitude != nil {
		t.Errorf("latitude = %v, want nil", *p.Latitude)
	}
	if p.Longitude != nil {
		t.Errorf("longitude = %v, want nil", *p.Longitude)
	}
	if p.PosResolution != nil {
		t.Errorf("posresolution = %v, want nil", *p.PosResolution)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 180 {
		t.Errorf("wind_direction = %v, want 180", wx.WindDirection)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != "0.4" {
		t.Errorf("wind_speed = %s, want 0.4", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "0.9" {
		t.Errorf("wind_gust = %s, want 0.9", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "0.6" {
		t.Errorf("temp = %s, want 0.6", got)
	}
	if wx.Humidity == nil || *wx.Humidity != 98 {
		t.Errorf("humidity = %v, want 98", wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "986.0" {
		t.Errorf("pressure = %s, want 986.0", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Rain1h); got != "2.5" {
		t.Errorf("rain_1h = %s, want 2.5", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.Rain24h); got != "10.2" {
		t.Errorf("rain_24h = %s, want 10.2", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "20.3" {
		t.Errorf("rain_midnight = %s, want 20.3", got)
	}

	if wx.Snow24h == nil {
		t.Fatalf("snow_24h is nil")
	}
	if got := fmt.Sprintf("%.1f", *wx.Snow24h); got != "2.5" {
		t.Errorf("snow_24h = %s, want 2.5", got)
	}

	if wx.Luminosity == nil || *wx.Luminosity != 500 {
		t.Errorf("luminosity = %v, want 500", wx.Luminosity)
	}
}
