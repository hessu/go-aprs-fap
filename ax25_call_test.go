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
		{"no SSID", "N0CAL", "N0CAL"},
		{"SSID 9", "N0CAL-9", "N0CAL-9"},
		{"SSID 15", "N0CAL-15", "N0CAL-15"},
		{"SSID characters invalid", "N0CAL-IS", ""},
		{"SSID dash without SSID invalid", "N0CAL-", ""},
		{"space in front invalid", " N0CAL", ""},
		{"space in back invalid", "N0CAL ", ""},
		{"two dashes SSID invalid", "N0CAL--1", ""},
		{"post-dash SSID invalid", "N0CA-1-", ""},
		{"only SSID invalid", "-1", ""},
		{"underscore SSID invalid", "N0CAL_1", ""},
		{"SSID 16 invalid", "N0CAL-16", ""},
		{"SSID 166 invalid", "N0CAL-166", ""},
		{"too long invalid", "N0CALXXXX", ""},
		{"characters invalid", "N0CÄL-1", ""},
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
