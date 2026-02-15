package fap

import (
	"testing"
)

// Tests ported from perl-aprs-fap/t/90check_ax25_call.t

func TestCheckAX25Call(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no SSID", "OH7LZB", "OH7LZB"},
		{"SSID 9", "OH7LZB-9", "OH7LZB-9"},
		{"SSID 15", "OH7LZB-15", "OH7LZB-15"},
		{"SSID 16 invalid", "OH7LZB-16", ""},
		{"SSID 166 invalid", "OH7LZB-166", ""},
		{"too long invalid", "OH7LZBXXX", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CheckAX25Call(tc.input)
			if got != tc.want {
				t.Errorf("CheckAX25Call(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
