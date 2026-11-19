package fap

// Tests ported from perl-aprs-fap/t/30decode-wx-basic.t

import (
	"fmt"
	"testing"
)

func TestWxBasic(t *testing.T) {
	packet := "OH2RDP-1>BEACON-15,WIDE2-1,qAo,OH2MQK-1:=6030.35N/02443.91E_150/002g004t039r001P002p004h00b10125XRSW"

	p, err := Parse(packet)
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

	if wx.Software != "XRSW" {
		t.Errorf("soft = %q, want %q", wx.Software, "XRSW")
	}
}

func TestWxWithComment(t *testing.T) {
	packet := "OH2GAX>APU25N,TCPIP*,qAC,OH2GAX:@101317z6024.78N/02503.97E_156/001g005t038r000p000P000h91b10093/type ?sade for more wx info"

	p, err := Parse(packet)
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

	p, err := Parse(packet)
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

	p, err := Parse(packet)
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

	p, err := Parse(packet)
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

	p, err := Parse(packet)
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

func TestWxWaterLevel(t *testing.T) {
	tests := []struct {
		name    string
		packet  string
		want    string
		wantNil bool
	}{
		{
			name:   "positive water level",
			packet: "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033F0123",
			want:   fmt.Sprintf("%.4f", 123.0/10.0*0.3048),
		},
		{
			name:   "negative water level",
			packet: "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033F-042",
			want:   fmt.Sprintf("%.4f", -42.0/10.0*0.3048),
		},
		{
			name:    "missing water level",
			packet:  "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033F....",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse packet: %v", err)
			}

			wx := p.Wx
			if wx == nil {
				t.Fatalf("wx is nil")
			}

			if tc.wantNil {
				if wx.WaterLevel != nil {
					t.Errorf("water_level = %v, want nil", *wx.WaterLevel)
				}
				return
			}

			if wx.WaterLevel == nil {
				t.Fatalf("water_level is nil, want %s", tc.want)
			}
			if got := fmt.Sprintf("%.4f", *wx.WaterLevel); got != tc.want {
				t.Errorf("water_level = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestWxRadiation(t *testing.T) {
	tests := []struct {
		name    string
		packet  string
		want    float64
		wantNil bool
	}{
		{
			name:   "12 microsieverts/hr",
			packet: "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033X123",
			want:   12000.0, // 12 * 10^3 nSv/hr
		},
		{
			name:   "45 millisieverts/hr",
			packet: "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033X456",
			want:   45000000.0, // 45 * 10^6 nSv/hr
		},
		{
			name:   "zero exponent",
			packet: "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033X100",
			want:   10.0, // 10 * 10^0 nSv/hr
		},
		{
			name:    "missing radiation",
			packet:  "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033X...",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse packet: %v", err)
			}

			wx := p.Wx
			if wx == nil {
				t.Fatalf("wx is nil")
			}

			if tc.wantNil {
				if wx.Radiation != nil {
					t.Errorf("radiation = %v, want nil", *wx.Radiation)
				}
				return
			}

			if wx.Radiation == nil {
				t.Fatalf("radiation is nil, want %v", tc.want)
			}
			if got := fmt.Sprintf("%.4f", *wx.Radiation); got != fmt.Sprintf("%.4f", tc.want) {
				t.Errorf("radiation = %s, want %.4f", got, tc.want)
			}
		})
	}
}

func TestWxBatteryVoltage(t *testing.T) {
	tests := []struct {
		name    string
		packet  string
		want    string
		wantNil bool
	}{
		{
			name:   "battery voltage",
			packet: "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033V128",
			want:   fmt.Sprintf("%.1f", 128.0/10.0),
		},
		{
			name:    "missing battery voltage",
			packet:  "N0CALL>APRS,TCPIP*:_12032359c000s000g000t033V...",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse packet: %v", err)
			}

			wx := p.Wx
			if wx == nil {
				t.Fatalf("wx is nil")
			}

			if tc.wantNil {
				if wx.BatteryVoltage != nil {
					t.Errorf("battery_voltage = %v, want nil", *wx.BatteryVoltage)
				}
				return
			}

			if wx.BatteryVoltage == nil {
				t.Fatalf("battery_voltage is nil, want %s", tc.want)
			}
			if got := fmt.Sprintf("%.1f", *wx.BatteryVoltage); got != tc.want {
				t.Errorf("battery_voltage = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestIsSoftwareID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"XRSW", true},
		{"abc", true},
		{"Ab-_9", true},
		{"Z1234", true},
		// too short
		{"ab", false},
		{"", false},
		// too long
		{"abcdef", false},
		// invalid characters
		{"ab c", false},
		{"ab.c", false},
		{"ab/c", false},
		{"XR{W", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := isSoftwareID(tc.input); got != tc.want {
				t.Errorf("isSoftwareID(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestWxAllFieldsMissing(t *testing.T) {
	// Every weather field is dots (missing), triggering skipWxField for all fields.
	packet := "N0CALL>APRS,TCPIP*:_12032359c...s...g...t...r...p...P...h..b.....#...Os...L...F....X...V..."

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection != nil {
		t.Errorf("wind_direction = %v, want nil", *wx.WindDirection)
	}
	if wx.WindSpeed != nil {
		t.Errorf("wind_speed = %v, want nil", *wx.WindSpeed)
	}
	if wx.WindGust != nil {
		t.Errorf("wind_gust = %v, want nil", *wx.WindGust)
	}
	if wx.Temp != nil {
		t.Errorf("temp = %v, want nil", *wx.Temp)
	}
	if wx.Rain1h != nil {
		t.Errorf("rain_1h = %v, want nil", *wx.Rain1h)
	}
	if wx.Rain24h != nil {
		t.Errorf("rain_24h = %v, want nil", *wx.Rain24h)
	}
	if wx.RainMidnight != nil {
		t.Errorf("rain_midnight = %v, want nil", *wx.RainMidnight)
	}
	if wx.Humidity != nil {
		t.Errorf("humidity = %v, want nil", *wx.Humidity)
	}
	if wx.Pressure != nil {
		t.Errorf("pressure = %v, want nil", *wx.Pressure)
	}
	if wx.Snow24h != nil {
		t.Errorf("snow_24h = %v, want nil", *wx.Snow24h)
	}
	if wx.Luminosity != nil {
		t.Errorf("luminosity = %v, want nil", *wx.Luminosity)
	}
	if wx.WaterLevel != nil {
		t.Errorf("water_level = %v, want nil", *wx.WaterLevel)
	}
	if wx.Radiation != nil {
		t.Errorf("radiation = %v, want nil", *wx.Radiation)
	}
	if wx.BatteryVoltage != nil {
		t.Errorf("battery_voltage = %v, want nil", *wx.BatteryVoltage)
	}
}

func TestWxPositionlessWithSnowfall(t *testing.T) {
	packet := "JH9YVX>APU25N,TCPIP*,qAC,T2TOKYO3:_12032359c180s001g002t033r010p040P080#456b09860h98Os010L500"

	p, err := Parse(packet)
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

// TestWxPlaceholderLength tests missing weather fields whose placeholder
// is not as long as the field: "h..." and "L....". The fields after them
// must still be parsed. The packets are from APRS-IS, the expected values
// are from the Perl module.
func TestWxPlaceholderLength(t *testing.T) {
	tests := []struct {
		name     string
		packet   string
		pressure string
		humidity string
		software string
		comment  string
	}{
		{
			name:     "humidity",
			packet:   "TEST>APRS,TCPIP*,qAC,AMBCWOP-2:@192201z4627.22N/12403.24W_.../...g...t...r000p000P000h...b10109AmbientCWOP.com",
			pressure: "1010.9",
			comment:  "AmbientCWOP.com",
		},
		{
			name:     "all but pressure",
			packet:   "TEST-9>APRS,TCPIP*,qAC,AMBCWOP-2:@192147z4533.17N/12233.22W_.../...g...t...r...p...P...h...b10139AmbientCWOP.com",
			pressure: "1013.9",
			comment:  "AmbientCWOP.com",
		},
		{
			name:     "luminosity",
			packet:   "TEST>APRS,TCPIP*,qAC,SECOND:@192200z3512.00N/07748.00W_051/000g000t062r000p000P000h66b10287L....DsIP",
			pressure: "1028.7",
			humidity: "66",
			software: "DsIP",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if p.Wx == nil {
				t.Fatal("wx is nil")
			}
			if p.Wx.Pressure == nil {
				t.Fatal("pressure is nil")
			}
			if got := fmt.Sprintf("%.1f", *p.Wx.Pressure); got != tc.pressure {
				t.Errorf("pressure = %s, want %s", got, tc.pressure)
			}
			got := ""
			if p.Wx.Humidity != nil {
				got = fmt.Sprintf("%d", *p.Wx.Humidity)
			}
			if got != tc.humidity {
				t.Errorf("humidity = %q, want %q", got, tc.humidity)
			}
			if p.Wx.Luminosity != nil {
				t.Errorf("luminosity = %d, want nil", *p.Wx.Luminosity)
			}
			if p.Wx.Software != tc.software {
				t.Errorf("software = %q, want %q", p.Wx.Software, tc.software)
			}
			if p.Comment != tc.comment {
				t.Errorf("comment = %q, want %q", p.Comment, tc.comment)
			}
		})
	}
}

// TestWxShortAndLongValues tests weather values which do not have the
// nominal number of digits of the field: a 2-digit temperature and a
// 3-digit humidity. Humidity over 100% is ignored. The first two packets
// are from APRS-IS, the expected values are from the Perl module.
func TestWxShortAndLongValues(t *testing.T) {
	tests := []struct {
		name     string
		packet   string
		temp     string
		humidity string
		pressure string
	}{
		{
			name:     "2-digit temperature",
			packet:   "TEST-13>APRS,TCPIP*,qAS,TEST:!3553.76N/07854.92W_098/000g000t58r000p000P000h47b10178",
			temp:     "14.4",
			humidity: "47",
			pressure: "1017.8",
		},
		{
			name:     "humidity 100",
			packet:   "TEST-15>APRS,WIDE1-1,qAR,SK3W-10:=6000.88N/01626.37E_.../...g...t360h100b-976Batt=93%",
			temp:     "182.2",
			humidity: "100",
		},
		{
			name:     "humidity over 100",
			packet:   "TEST-15>APRS,WIDE1-1,qAR,SK3W-10:=6000.88N/01626.37E_.../...g...t060h101b10123",
			temp:     "15.6",
			pressure: "1012.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if p.Wx == nil {
				t.Fatal("wx is nil")
			}
			got := ""
			if p.Wx.Temp != nil {
				got = fmt.Sprintf("%.1f", *p.Wx.Temp)
			}
			if got != tc.temp {
				t.Errorf("temp = %q, want %q", got, tc.temp)
			}
			got = ""
			if p.Wx.Humidity != nil {
				got = fmt.Sprintf("%d", *p.Wx.Humidity)
			}
			if got != tc.humidity {
				t.Errorf("humidity = %q, want %q", got, tc.humidity)
			}
			got = ""
			if p.Wx.Pressure != nil {
				got = fmt.Sprintf("%.1f", *p.Wx.Pressure)
			}
			if got != tc.pressure {
				t.Errorf("pressure = %q, want %q", got, tc.pressure)
			}
		})
	}
}

// TestWxSnowfallAfterWind tests snowfall 's' in a position report, where
// the wind speed is given before it in the CCC/SSS wind field. The first
// packet is from APRS-IS, the second has a snowfall and a wind speed. The
// expected values are from the Perl module.
func TestWxSnowfallAfterWind(t *testing.T) {
	tests := []struct {
		name      string
		packet    string
		windSpeed string
		snow      string
		humidity  string
	}{
		{
			name:      "zero",
			packet:    "TEST-13>APPRPV,TCPIP*,qAS,TEST-1:/191653z3531.94N/07945.00W_270/000t059P000p000r000s000h62b10062",
			windSpeed: "0.0",
			snow:      "0.0",
			humidity:  "62",
		},
		{
			name:      "snowing",
			packet:    "TEST-13>APPRPV,TCPIP*,qAS,TEST-1:/191653z3531.94N/07945.00W_270/005t031P000p000r000s012h62b10062",
			windSpeed: "2.2",
			snow:      "3.0",
			humidity:  "62",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			wx := p.Wx
			if wx == nil {
				t.Fatal("wx is nil")
			}
			if wx.WindSpeed == nil || wx.Snow24h == nil || wx.Humidity == nil {
				t.Fatalf("wind_speed %v, snow_24h %v, humidity %v: want all set", wx.WindSpeed, wx.Snow24h, wx.Humidity)
			}
			if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != tc.windSpeed {
				t.Errorf("wind_speed = %s, want %s", got, tc.windSpeed)
			}
			if got := fmt.Sprintf("%.1f", *wx.Snow24h); got != tc.snow {
				t.Errorf("snow_24h = %s, want %s", got, tc.snow)
			}
			if got := fmt.Sprintf("%d", *wx.Humidity); got != tc.humidity {
				t.Errorf("humidity = %s, want %s", got, tc.humidity)
			}
		})
	}
}
