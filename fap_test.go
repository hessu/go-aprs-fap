package fap

import (
	"errors"
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

func TestIsIPv6Hex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// Valid: 32 uppercase hex chars
		{"00000000000000000000000000000000", true},
		{"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", true},
		{"0123456789ABCDEF0123456789ABCDEF", true},

		// Too short / too long / empty
		{"0123456789ABCDEF0123456789ABCDE", false},
		{"0123456789ABCDEF0123456789ABCDEF0", false},
		{"", false},

		// Lowercase hex not accepted
		{"0123456789abcdef0123456789abcdef", false},
		{"0123456789ABCDEf0123456789ABCDEF", false},

		// Non-hex characters
		{"G0000000000000000000000000000000", false},
		{"0000000000000000000000000000000Z", false},
		{"00000000000000000000000000000 00", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := isIPv6Hex(tc.input); got != tc.want {
				t.Errorf("isIPv6Hex(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name   string
		packet string
		opts   []Option
		errIs  error
	}{
		// Parse: header/body splitting
		{
			name:   "no colon",
			packet: "OH7LZB>APRS",
			errIs:  ErrPacketNoBody,
		},
		{
			name:   "empty body",
			packet: "OH7LZB>APRS:",
			errIs:  ErrPacketNoBody,
		},

		// parseHeader: source callsign
		{
			name:   "no > in header",
			packet: "OH7LZB:body",
			errIs:  ErrSrcCallNoGT,
		},
		{
			name:   "empty source callsign",
			packet: ">APRS:body",
			errIs:  ErrSrcCallEmpty,
		},
		{
			name:   "source callsign bad chars",
			packet: "OH7_LZB>APRS:body",
			errIs:  ErrSrcCallBadChars,
		},
		{
			name:   "source callsign not valid AX.25",
			packet: "TOOLONGCALL>APRS:body",
			opts:   []Option{WithAX25()},
			errIs:  ErrSrcCallNoAX25,
		},

		// parseHeader: destination callsign
		{
			name:   "empty destination",
			packet: "OH7LZB>:body",
			errIs:  ErrDstCallEmpty,
		},
		{
			name:   "destination not valid AX.25",
			packet: "OH7LZB>!!!:body",
			errIs:  ErrDstCallNoAX25,
		},

		// parseHeader: path too many (AX.25)
		{
			name:   "too many path components for AX.25",
			packet: "OH7LZB>APRS,A,B,C,D,E,F,G,H,I:body",
			opts:   []Option{WithAX25()},
			errIs:  ErrDstPathTooMany,
		},

		// parseHeader: digipeaters
		{
			name:   "empty digipeater callsign",
			packet: "OH7LZB>APRS,,WIDE:body",
			errIs:  ErrDigiEmpty,
		},
		{
			name:   "empty digipeater after star",
			packet: "OH7LZB>APRS,*:body",
			errIs:  ErrDigiEmpty,
		},
		{
			name:   "digipeater bad chars",
			packet: "OH7LZB>APRS,WI_DE:body",
			errIs:  ErrDigiCallBadChars,
		},
		{
			name:   "digipeater not valid AX.25",
			packet: "OH7LZB>APRS,TOOLONGCALL:body",
			opts:   []Option{WithAX25()},
			errIs:  ErrDigiCallNoAX25,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.packet, tc.opts...)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.errIs)
			}
			if !errors.Is(err, tc.errIs) {
				t.Errorf("error = %v, want %v", err, tc.errIs)
			}
		})
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
