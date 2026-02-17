package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	fap "github.com/hessu/go-aprs-fap"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var input string

	if len(args) > 0 {
		input = strings.Join(args, " ")
	} else {
		scanner := bufio.NewScanner(stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		if scanner.Scan() {
			input = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(stderr, "read error: %v\n", err)
			return 1
		}
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Fprintf(stderr, "Usage: aprs-decode-single <packet>\n")
		fmt.Fprintf(stderr, "   or: echo '<packet>' | aprs-decode-single\n")
		return 1
	}

	p, err := fap.Parse(input)

	if p != nil {
		printPacket(stdout, p)
	}

	if len(p.Warnings) > 0 {
		fmt.Fprintf(stdout, "\nWarnings:\n")
		for _, w := range p.Warnings {
			fmt.Fprintf(stdout, "  [%s] %s\n", w.Code, w.Msg)
		}
	}

	if err != nil {
		fmt.Fprintf(stdout, "\nError: %s\n", err)
		return 1
	}

	return 0
}

func printPacket(w io.Writer, p *fap.Packet) {
	fmt.Fprintf(w, "Original:     %s\n", p.OrigPacket)
	fmt.Fprintf(w, "Header:       %s\n", p.Header)
	fmt.Fprintf(w, "Body:         %s\n", p.Body)
	fmt.Fprintf(w, "Source:       %s\n", p.SrcCallsign)
	fmt.Fprintf(w, "Destination:  %s\n", p.DstCallsign)

	if len(p.Digipeaters) > 0 {
		digis := make([]string, len(p.Digipeaters))
		for i, d := range p.Digipeaters {
			if d.WasDigied {
				digis[i] = d.Call + "*"
			} else {
				digis[i] = d.Call
			}
		}
		fmt.Fprintf(w, "Digipeaters:  %s\n", strings.Join(digis, ","))
	}

	if p.Type != "" {
		fmt.Fprintf(w, "Type:         %s\n", p.Type)
	}
	if p.Format != "" {
		fmt.Fprintf(w, "Format:       %s\n", p.Format)
	}

	if p.Latitude != nil {
		fmt.Fprintf(w, "Latitude:     %.6f\n", *p.Latitude)
	}
	if p.Longitude != nil {
		fmt.Fprintf(w, "Longitude:    %.6f\n", *p.Longitude)
	}
	if p.PosAmbiguity != nil {
		fmt.Fprintf(w, "PosAmbiguity: %d\n", *p.PosAmbiguity)
	}
	if p.PosResolution != nil {
		fmt.Fprintf(w, "PosResolution: %.1f m\n", *p.PosResolution)
	}

	if p.SymbolTable != 0 {
		fmt.Fprintf(w, "SymbolTable:  %c\n", p.SymbolTable)
	}
	if p.SymbolCode != 0 {
		fmt.Fprintf(w, "SymbolCode:   %c\n", p.SymbolCode)
	}

	if p.Speed != nil {
		fmt.Fprintf(w, "Speed:        %.1f km/h\n", *p.Speed)
	}
	if p.Course != nil {
		fmt.Fprintf(w, "Course:       %d°\n", *p.Course)
	}
	if p.Altitude != nil {
		fmt.Fprintf(w, "Altitude:     %.1f m\n", *p.Altitude)
	}

	if p.Messaging != nil {
		fmt.Fprintf(w, "Messaging:    %v\n", *p.Messaging)
	}

	if p.PHG != "" {
		fmt.Fprintf(w, "PHG:          %s\n", p.PHG)
	}
	if p.RadioRange != nil {
		fmt.Fprintf(w, "RadioRange:   %.1f km\n", *p.RadioRange)
	}

	if p.Timestamp != nil {
		fmt.Fprintf(w, "Timestamp:    %s\n", p.Timestamp.Format(time.RFC3339))
	}
	if p.RawTimestamp != "" {
		fmt.Fprintf(w, "RawTimestamp:  %s\n", p.RawTimestamp)
	}

	if p.ObjectName != "" {
		fmt.Fprintf(w, "ObjectName:   %s\n", p.ObjectName)
	}
	if p.ItemName != "" {
		fmt.Fprintf(w, "ItemName:     %s\n", p.ItemName)
	}
	if p.Alive != nil {
		fmt.Fprintf(w, "Alive:        %v\n", *p.Alive)
	}

	if p.Message != nil {
		fmt.Fprintf(w, "Message:\n")
		fmt.Fprintf(w, "  Destination: %s\n", p.Message.Destination)
		if p.Message.Text != "" {
			fmt.Fprintf(w, "  Text:        %s\n", p.Message.Text)
		}
		if p.Message.ID != "" {
			fmt.Fprintf(w, "  ID:          %s\n", p.Message.ID)
		}
		if p.Message.AckID != "" {
			fmt.Fprintf(w, "  AckID:       %s\n", p.Message.AckID)
		}
		if p.Message.RejID != "" {
			fmt.Fprintf(w, "  RejID:       %s\n", p.Message.RejID)
		}
	}

	if p.Status != "" {
		fmt.Fprintf(w, "Status:       %s\n", p.Status)
	}

	if p.Wx != nil {
		printWeather(w, p.Wx)
	}

	if p.TelemetryData != nil {
		printTelemetry(w, p.TelemetryData)
	}

	if p.Capabilities != nil {
		fmt.Fprintf(w, "Capabilities:\n")
		for k, v := range p.Capabilities {
			if v != "" {
				fmt.Fprintf(w, "  %s=%s\n", k, v)
			} else {
				fmt.Fprintf(w, "  %s\n", k)
			}
		}
	}

	if p.MBits != "" {
		fmt.Fprintf(w, "MicE Bits:    %s\n", p.MBits)
	}
	if p.MiceMangled {
		fmt.Fprintf(w, "MicE Mangled: true\n")
	}

	if p.DaoDatumByte != 0 {
		fmt.Fprintf(w, "DAO Datum:    %c\n", p.DaoDatumByte)
	}

	if p.GPSFixStatus != nil {
		fmt.Fprintf(w, "GPS Fix:      %d\n", *p.GPSFixStatus)
	}
	if p.ChecksumOK != nil {
		fmt.Fprintf(w, "Checksum OK:  %v\n", *p.ChecksumOK)
	}

	if p.Comment != "" {
		fmt.Fprintf(w, "Comment:      %s\n", p.Comment)
	}
}

func printWeather(w io.Writer, wx *fap.Weather) {
	fmt.Fprintf(w, "Weather:\n")
	if wx.WindDirection != nil {
		fmt.Fprintf(w, "  Wind Dir:     %.0f°\n", *wx.WindDirection)
	}
	if wx.WindSpeed != nil {
		fmt.Fprintf(w, "  Wind Speed:   %.1f m/s\n", *wx.WindSpeed)
	}
	if wx.WindGust != nil {
		fmt.Fprintf(w, "  Wind Gust:    %.1f m/s\n", *wx.WindGust)
	}
	if wx.Temp != nil {
		fmt.Fprintf(w, "  Temp:         %.1f °C\n", *wx.Temp)
	}
	if wx.TempIn != nil {
		fmt.Fprintf(w, "  Temp Indoor:  %.1f °C\n", *wx.TempIn)
	}
	if wx.Humidity != nil {
		fmt.Fprintf(w, "  Humidity:     %d%%\n", *wx.Humidity)
	}
	if wx.HumidityIn != nil {
		fmt.Fprintf(w, "  Humidity In:  %d%%\n", *wx.HumidityIn)
	}
	if wx.Pressure != nil {
		fmt.Fprintf(w, "  Pressure:     %.1f mbar\n", *wx.Pressure)
	}
	if wx.Rain1h != nil {
		fmt.Fprintf(w, "  Rain 1h:      %.1f mm\n", *wx.Rain1h)
	}
	if wx.Rain24h != nil {
		fmt.Fprintf(w, "  Rain 24h:     %.1f mm\n", *wx.Rain24h)
	}
	if wx.RainMidnight != nil {
		fmt.Fprintf(w, "  Rain Today:   %.1f mm\n", *wx.RainMidnight)
	}
	if wx.Snow24h != nil {
		fmt.Fprintf(w, "  Snow 24h:     %.1f mm\n", *wx.Snow24h)
	}
	if wx.Luminosity != nil {
		fmt.Fprintf(w, "  Luminosity:   %d W/m²\n", *wx.Luminosity)
	}
	if wx.WaterLevel != nil {
		fmt.Fprintf(w, "  Water Level:  %.2f m\n", *wx.WaterLevel)
	}
	if wx.Radiation != nil {
		fmt.Fprintf(w, "  Radiation:    %.1f nSv/h\n", *wx.Radiation)
	}
	if wx.BatteryVoltage != nil {
		fmt.Fprintf(w, "  Battery:      %.1f V\n", *wx.BatteryVoltage)
	}
	if wx.Software != "" {
		fmt.Fprintf(w, "  Software:     %s\n", wx.Software)
	}
}

func printTelemetry(w io.Writer, t *fap.Telemetry) {
	fmt.Fprintf(w, "Telemetry:\n")
	fmt.Fprintf(w, "  Seq:    %d\n", t.Seq)
	if len(t.Vals) > 0 {
		for i, v := range t.Vals {
			if v != nil {
				fmt.Fprintf(w, "  Val %d:  %.2f\n", i+1, *v)
			} else {
				fmt.Fprintf(w, "  Val %d:  (undefined)\n", i+1)
			}
		}
	}
	if t.Bits != "" {
		fmt.Fprintf(w, "  Bits:   %s\n", t.Bits)
	}
}
