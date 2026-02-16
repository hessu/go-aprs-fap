package fap

import (
	"errors"
	"testing"
)

// Tests ported from perl-aprs-fap/t/53decode-tlm.t

func TestTelemetryClassic(t *testing.T) {
	// Classic packet with one floating point value
	p, err := Parse("SRCCALL>APRS:T#324,000,038,255,.12,50.12,01000001")
	if err != nil {
		t.Fatalf("failed to parse telemetry packet: %v", err)
	}
	if p.TelemetryData == nil {
		t.Fatal("no telemetry data")
	}

	tlm := p.TelemetryData
	if tlm.Seq != "324" {
		t.Errorf("seq = %q, want %q", tlm.Seq, "324")
	}
	if tlm.Bits != "01000001" {
		t.Errorf("bits = %q, want %q", tlm.Bits, "01000001")
	}
	if len(tlm.Vals) < 5 {
		t.Fatalf("vals length = %d, want >= 5", len(tlm.Vals))
	}

	wantVals := []float64{0, 38, 255, 0.12, 50.12}
	for i, want := range wantVals {
		if tlm.Vals[i] == nil {
			t.Errorf("vals[%d] = nil, want %v", i, want)
		} else if *tlm.Vals[i] != want {
			t.Errorf("vals[%d] = %v, want %v", i, *tlm.Vals[i], want)
		}
	}
}

func TestTelemetryRelaxed(t *testing.T) {
	// Floating-point and negative values (relaxed rules)
	p, err := Parse("SRCCALL>APRS:T#1,-1,2147483647,-2147483648,0.000001,-0.0000001,01000001 comment")
	if err != nil {
		t.Fatalf("failed to parse relaxed telemetry: %v", err)
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}
	if tlm.Seq != "1" {
		t.Errorf("seq = %q, want %q", tlm.Seq, "1")
	}
	if tlm.Bits != "01000001" {
		t.Errorf("bits = %q, want %q", tlm.Bits, "01000001")
	}

	wantVals := []float64{-1, 2147483647, -2147483648, 0.000001, -0.0000001}
	for i, want := range wantVals {
		if tlm.Vals[i] == nil {
			t.Errorf("vals[%d] = nil, want %v", i, want)
		} else if *tlm.Vals[i] != want {
			t.Errorf("vals[%d] = %v, want %v", i, *tlm.Vals[i], want)
		}
	}
}

func TestTelemetryShort(t *testing.T) {
	// Very short telemetry packet (only one value)
	p, err := Parse("SRCCALL>APRS:T#001,42")
	if err != nil {
		t.Fatalf("failed to parse short telemetry: %v", err)
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}
	if tlm.Seq != "001" {
		t.Errorf("seq = %q, want %q", tlm.Seq, "001")
	}
	if tlm.Bits != "" {
		t.Errorf("bits = %q, want empty", tlm.Bits)
	}
	if len(tlm.Vals) < 5 {
		t.Fatalf("vals length = %d, want >= 5", len(tlm.Vals))
	}
	if tlm.Vals[0] == nil || *tlm.Vals[0] != 42 {
		t.Errorf("vals[0] = %v, want 42", tlm.Vals[0])
	}
	for i := 1; i <= 4; i++ {
		if tlm.Vals[i] != nil {
			t.Errorf("vals[%d] = %v, want nil", i, *tlm.Vals[i])
		}
	}
}

func TestTelemetryUndefinedMiddle(t *testing.T) {
	// Undefined values in the middle
	p, err := Parse("SRCCALL>APRS:T#1,1,,3,,5")
	if err != nil {
		t.Fatalf("failed to parse telemetry with undefined middle: %v", err)
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}

	// vals[0]=1, vals[1]=nil, vals[2]=3, vals[3]=nil, vals[4]=5
	type check struct {
		idx  int
		want *float64
	}
	checks := []check{
		{0, new(1.0)},
		{1, nil},
		{2, new(3.0)},
		{3, nil},
		{4, new(5.0)},
	}
	for _, c := range checks {
		if c.want == nil {
			if tlm.Vals[c.idx] != nil {
				t.Errorf("vals[%d] = %v, want nil", c.idx, *tlm.Vals[c.idx])
			}
		} else {
			if tlm.Vals[c.idx] == nil {
				t.Errorf("vals[%d] = nil, want %v", c.idx, *c.want)
			} else if *tlm.Vals[c.idx] != *c.want {
				t.Errorf("vals[%d] = %v, want %v", c.idx, *tlm.Vals[c.idx], *c.want)
			}
		}
	}
}

func TestTelemetryPartiallyCorrect(t *testing.T) {
	// Parsing ends at 'f' since it may be a comment
	p, err := Parse("SRCCALL>APRS:T#1,1,f,3")
	if err != nil {
		t.Fatalf("failed to parse partially correct telemetry: %v", err)
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}

	if tlm.Vals[0] == nil || *tlm.Vals[0] != 1 {
		t.Errorf("vals[0] = %v, want 1", tlm.Vals[0])
	}
	for i := 1; i <= 4; i++ {
		if tlm.Vals[i] != nil {
			t.Errorf("vals[%d] = %v, want nil", i, *tlm.Vals[i])
		}
	}
}

func TestIsNumericValue(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// Valid integers
		{"0", true},
		{"1", true},
		{"42", true},
		{"123456789", true},

		// Valid negative integers
		{"-1", true},
		{"-42", true},

		// Valid floats
		{"0.5", true},
		{"3.14", true},
		{".5", true},
		{".123", true},
		{"0.000001", true},

		// Valid negative floats
		{"-0.5", true},
		{"-.5", true},
		{"-0.0000001", true},

		// Valid: leading digits before dot
		{"123.456", true},
		{"2147483647", true},

		// Invalid: empty
		{"", false},

		// Invalid: bare minus
		{"-", false},

		// Invalid: trailing dot, no digits after
		{"1.", false},
		{"-1.", false},
		{"0.", false},

		// Invalid: bare dot
		{".", false},
		{"-.", false},

		// Invalid: letters
		{"abc", false},
		{"1a", false},
		{"a1", false},
		{"-a", false},

		// Invalid: spaces
		{" 1", false},
		{"1 ", false},
		{" ", false},

		// Invalid: multiple dots
		{"1.2.3", false},

		// Invalid: multiple minuses
		{"--1", false},

		// Invalid: minus not at start
		{"1-2", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := isNumericValue(tc.input)
			if got != tc.want {
				t.Errorf("isNumericValue(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestTelemetryInvalidDash(t *testing.T) {
	// Invalid: bare '-' with no number
	_, err := Parse("SRCCALL>APRS:T#1,1,-,3")
	if err == nil {
		t.Fatal("expected error for invalid telemetry with bare dash")
	}
	if !errors.Is(err, ErrTlmInvalid) {
		t.Errorf("error = %v, want %v", err, ErrTlmInvalid)
	}
}

func TestTelemetryInvalidTrailingDot(t *testing.T) {
	// Invalid: trailing dot after number
	_, err := Parse("SRCCALL>APRS:T#1,1,-1.,3")
	if err == nil {
		t.Fatal("expected error for invalid telemetry with trailing dot")
	}
	if !errors.Is(err, ErrTlmInvalid) {
		t.Errorf("error = %v, want %v", err, ErrTlmInvalid)
	}
}
