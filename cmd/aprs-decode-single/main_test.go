package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestArgInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"N0CALL-1>APRS:;LEADER   *092345z4903.50N/07201.75W>088/036"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Source:       N0CALL-1") {
		t.Errorf("output missing source callsign:\n%s", output)
	}
	if !strings.Contains(output, "ObjectName:   LEADER") {
		t.Errorf("output missing object name:\n%s", output)
	}
}

func TestStdinInput(t *testing.T) {
	tests := []struct {
		name     string
		packet   string
		wantStrs []string
	}{
		{
			name:   "item with digipeaters",
			packet: "N0CALL-1>APRS,DIGI1,DIGI2*:)AID #2!4903.50N/07201.75WA",
			wantStrs: []string{
				"ItemName:     AID #2",
				"Digipeaters:  DIGI1,DIGI2*",
			},
		},
		{
			name:   "telemetry",
			packet: "N0CALL>APRS:T#324,000,038,255,.12,50.12,01000001",
			wantStrs: []string{
				"Telemetry:",
				"Seq:    324",
				"Bits:   01000001",
			},
		},
		{
			name:   "weather",
			packet: "N0CALL-1>BEACON-15,WIDE2-1,qAo,N0CALL-2:=6030.35N/02443.91E_150/002g004t039r001P002p004h00b10125L500F0123X123V128XRSW",
			wantStrs: []string{
				"Weather:",
				"Wind Dir:     150",
				"Wind Speed:",
				"Wind Gust:",
				"Temp:",
				"Humidity:     100%",
				"Pressure:     1012.5 mbar",
				"Rain 1h:",
				"Rain 24h:",
				"Rain Today:",
				"Luminosity:   500",
				"Water Level:",
				"Radiation:",
				"Battery:      12.8 V",
				"Software:     XRSW",
			},
		},
		{
			name:   "message",
			packet: "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :Testing, 1 2 3{1",
			wantStrs: []string{
				"Message:",
				"Destination: N0CALL",
				"Text:        Testing, 1 2 3",
				"ID:          1",
			},
		},
		{
			name:   "message ack",
			packet: "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :ack1",
			wantStrs: []string{
				"Message:",
				"Destination: N0CALL",
				"AckID:       1",
			},
		},
		{
			name:   "message reject",
			packet: "OH7AA-1>APRS,WIDE1-1,WIDE2-2,qAo,OH7AA::N0CALL   :rej1",
			wantStrs: []string{
				"Message:",
				"Destination: N0CALL",
				"RejID:       1",
			},
		},
		{
			name:   "status",
			packet: "N0CALL-14>APU25N,WIDE2-2,qAR,LANSNG:>051421>>Nashville,TN>>Toronto,ON",
			wantStrs: []string{
				"Type:         status",
				"Status:",
				"Nashville,TN>>Toronto,ON",
			},
		},
		{
			name:   "mic-e position",
			packet: "N0CALL-2>TQ4W2V,WIDE2-1,qAo,IGATE:`c51!f?>/]\"3x}=",
			wantStrs: []string{
				"Format:       mice",
				"Latitude:     41.787",
				"Longitude:    -71.420",
				"Speed:",
				"Course:       35",
				"Altitude:     6.0 m",
				"MicE Bits:    110",
			},
		},
		{
			name:   "position with PHG",
			packet: "N0CALL-1>BEACON-15,N0CALL-2*,WIDE:!6028.51N/02505.68E#PHG7220/RELAY,WIDE, OH2AP Jarvenpaa",
			wantStrs: []string{
				"Latitude:     60.475",
				"Longitude:    25.094",
				"PHG:          7220",
				"Comment:      RELAY,WIDE, OH2AP Jarvenpaa",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(nil, strings.NewReader(tc.packet+"\n"), &stdout, &stderr)
			if code != 0 {
				t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
			}
			output := stdout.String()
			for _, want := range tc.wantStrs {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q:\n%s", want, output)
				}
			}
		})
	}
}

func TestEmptyInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}

	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("stderr missing usage message:\n%s", stderr.String())
	}
}

func TestParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"INVALID"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}

	if !strings.Contains(stdout.String(), "Error:") {
		t.Errorf("stdout missing error message:\n%s", stdout.String())
	}
}
