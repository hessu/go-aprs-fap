package fap

import "testing"

func TestAprsPasscode(t *testing.T) {
	tests := []struct {
		callsign string
		expected int16
	}{
		{"N0CALL", 13023},
		{"N1CALL", 13022},
		// With SSID - should produce same result as without.
		{"N0CALL-9", 13023},
		// Lowercase - should produce same result as uppercase.
		{"n0call", 13023},
	}

	for _, tc := range tests {
		t.Run(tc.callsign, func(t *testing.T) {
			got := AprsPasscode(tc.callsign)
			if got != tc.expected {
				t.Errorf("AprsPasscode(%q) = %d, want %d", tc.callsign, got, tc.expected)
			}
		})
	}
}
