package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	fap "github.com/hessu/go-aprs-fap"
)

func main() {
	var input string

	if len(os.Args) > 1 {
		input = strings.Join(os.Args[1:], " ")
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		if scanner.Scan() {
			input = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			os.Exit(1)
		}
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Fprintf(os.Stderr, "Usage: aprs-decode-single <packet>\n")
		fmt.Fprintf(os.Stderr, "   or: echo '<packet>' | aprs-decode-single\n")
		os.Exit(1)
	}

	p, err := fap.Parse(input)

	if p != nil {
		printPacket(p)
	}

	if len(p.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range p.Warnings {
			fmt.Printf("  [%s] %s\n", w.Code, w.Msg)
		}
	}

	if err != nil {
		fmt.Printf("\nError: %s\n", err)
		os.Exit(1)
	}
}

func printPacket(p *fap.Packet) {
	fmt.Printf("Original:     %s\n", p.OrigPacket)
	fmt.Printf("Header:       %s\n", p.Header)
	fmt.Printf("Body:         %s\n", p.Body)
	fmt.Printf("Source:       %s\n", p.SrcCallsign)
	fmt.Printf("Destination:  %s\n", p.DstCallsign)

	if len(p.Digipeaters) > 0 {
		digis := make([]string, len(p.Digipeaters))
		for i, d := range p.Digipeaters {
			if d.WasDigied {
				digis[i] = d.Call + "*"
			} else {
				digis[i] = d.Call
			}
		}
		fmt.Printf("Digipeaters:  %s\n", strings.Join(digis, ","))
	}

	if p.Type != "" {
		fmt.Printf("Type:         %s\n", p.Type)
	}
	if p.Format != "" {
		fmt.Printf("Format:       %s\n", p.Format)
	}

	if p.Latitude != nil {
		fmt.Printf("Latitude:     %.6f\n", *p.Latitude)
	}
	if p.Longitude != nil {
		fmt.Printf("Longitude:    %.6f\n", *p.Longitude)
	}
	if p.PosAmbiguity != nil {
		fmt.Printf("PosAmbiguity: %d\n", *p.PosAmbiguity)
	}
	if p.PosResolution != nil {
		fmt.Printf("PosResolution: %.1f m\n", *p.PosResolution)
	}

	if p.SymbolTable != 0 {
		fmt.Printf("SymbolTable:  %c\n", p.SymbolTable)
	}
	if p.SymbolCode != 0 {
		fmt.Printf("SymbolCode:   %c\n", p.SymbolCode)
	}

	if p.Speed != nil {
		fmt.Printf("Speed:        %.1f km/h\n", *p.Speed)
	}
	if p.Course != nil {
		fmt.Printf("Course:       %d°\n", *p.Course)
	}
	if p.Altitude != nil {
		fmt.Printf("Altitude:     %.1f m\n", *p.Altitude)
	}

	if p.Messaging != nil {
		fmt.Printf("Messaging:    %v\n", *p.Messaging)
	}

	if p.PHG != "" {
		fmt.Printf("PHG:          %s\n", p.PHG)
	}
	if p.RadioRange != nil {
		fmt.Printf("RadioRange:   %.1f km\n", *p.RadioRange)
	}

	if p.Timestamp != nil {
		fmt.Printf("Timestamp:    %s\n", p.Timestamp.Format(time.RFC3339))
	}
	if p.RawTimestamp != "" {
		fmt.Printf("RawTimestamp:  %s\n", p.RawTimestamp)
	}

	if p.ObjectName != "" {
		fmt.Printf("ObjectName:   %s\n", p.ObjectName)
	}
	if p.ItemName != "" {
		fmt.Printf("ItemName:     %s\n", p.ItemName)
	}
	if p.Alive != nil {
		fmt.Printf("Alive:        %v\n", *p.Alive)
	}

	if p.Message != nil {
		fmt.Printf("Message:\n")
		fmt.Printf("  Destination: %s\n", p.Message.Destination)
		if p.Message.Text != "" {
			fmt.Printf("  Text:        %s\n", p.Message.Text)
		}
		if p.Message.ID != "" {
			fmt.Printf("  ID:          %s\n", p.Message.ID)
		}
		if p.Message.AckID != "" {
			fmt.Printf("  AckID:       %s\n", p.Message.AckID)
		}
		if p.Message.RejID != "" {
			fmt.Printf("  RejID:       %s\n", p.Message.RejID)
		}
	}

	if p.Status != "" {
		fmt.Printf("Status:       %s\n", p.Status)
	}

	if p.Wx != nil {
		printWeather(p.Wx)
	}

	if p.TelemetryData != nil {
		printTelemetry(p.TelemetryData)
	}

	if p.Capabilities != nil {
		fmt.Printf("Capabilities:\n")
		for k, v := range p.Capabilities {
			if v != "" {
				fmt.Printf("  %s=%s\n", k, v)
			} else {
				fmt.Printf("  %s\n", k)
			}
		}
	}

	if p.MBits != "" {
		fmt.Printf("MicE Bits:    %s\n", p.MBits)
	}
	if p.MiceMangled {
		fmt.Printf("MicE Mangled: true\n")
	}

	if p.DaoDatumByte != 0 {
		fmt.Printf("DAO Datum:    %c\n", p.DaoDatumByte)
	}

	if p.GPSFixStatus != nil {
		fmt.Printf("GPS Fix:      %d\n", *p.GPSFixStatus)
	}
	if p.ChecksumOK != nil {
		fmt.Printf("Checksum OK:  %v\n", *p.ChecksumOK)
	}

	if p.Comment != "" {
		fmt.Printf("Comment:      %s\n", p.Comment)
	}
}

func printWeather(wx *fap.Weather) {
	fmt.Printf("Weather:\n")
	if wx.WindDirection != nil {
		fmt.Printf("  Wind Dir:     %.0f°\n", *wx.WindDirection)
	}
	if wx.WindSpeed != nil {
		fmt.Printf("  Wind Speed:   %.1f m/s\n", *wx.WindSpeed)
	}
	if wx.WindGust != nil {
		fmt.Printf("  Wind Gust:    %.1f m/s\n", *wx.WindGust)
	}
	if wx.Temp != nil {
		fmt.Printf("  Temp:         %.1f °C\n", *wx.Temp)
	}
	if wx.TempIn != nil {
		fmt.Printf("  Temp Indoor:  %.1f °C\n", *wx.TempIn)
	}
	if wx.Humidity != nil {
		fmt.Printf("  Humidity:     %d%%\n", *wx.Humidity)
	}
	if wx.HumidityIn != nil {
		fmt.Printf("  Humidity In:  %d%%\n", *wx.HumidityIn)
	}
	if wx.Pressure != nil {
		fmt.Printf("  Pressure:     %.1f mbar\n", *wx.Pressure)
	}
	if wx.Rain1h != nil {
		fmt.Printf("  Rain 1h:      %.1f mm\n", *wx.Rain1h)
	}
	if wx.Rain24h != nil {
		fmt.Printf("  Rain 24h:     %.1f mm\n", *wx.Rain24h)
	}
	if wx.RainMidnight != nil {
		fmt.Printf("  Rain Today:   %.1f mm\n", *wx.RainMidnight)
	}
	if wx.Snow24h != nil {
		fmt.Printf("  Snow 24h:     %.1f mm\n", *wx.Snow24h)
	}
	if wx.Luminosity != nil {
		fmt.Printf("  Luminosity:   %d W/m²\n", *wx.Luminosity)
	}
	if wx.WaterLevel != nil {
		fmt.Printf("  Water Level:  %.2f m\n", *wx.WaterLevel)
	}
	if wx.Radiation != nil {
		fmt.Printf("  Radiation:    %.1f nSv/h\n", *wx.Radiation)
	}
	if wx.BatteryVoltage != nil {
		fmt.Printf("  Battery:      %.1f V\n", *wx.BatteryVoltage)
	}
	if wx.Software != "" {
		fmt.Printf("  Software:     %s\n", wx.Software)
	}
}

func printTelemetry(t *fap.Telemetry) {
	fmt.Printf("Telemetry:\n")
	fmt.Printf("  Seq:    %d\n", t.Seq)
	if len(t.Vals) > 0 {
		for i, v := range t.Vals {
			if v != nil {
				fmt.Printf("  Val %d:  %.2f\n", i+1, *v)
			} else {
				fmt.Printf("  Val %d:  (undefined)\n", i+1)
			}
		}
	}
	if t.Bits != "" {
		fmt.Printf("  Bits:   %s\n", t.Bits)
	}
}
