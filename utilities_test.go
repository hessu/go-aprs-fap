package fap

import "testing"

// TestDistance tests great-circle distance calculation.
func TestDistance(t *testing.T) {
	// Helsinki to Tampere, approximately 161 km
	d := Distance(60.17, 24.94, 61.50, 23.78)
	if d < 155 || d > 165 {
		t.Errorf("distance Helsinki-Tampere: got %.1f, want ~161", d)
	}
}

// TestDirection tests bearing calculation.
func TestDirection(t *testing.T) {
	// Due north
	d := Direction(60.0, 25.0, 61.0, 25.0)
	if d < 355 && d > 5 {
		t.Errorf("direction north: got %.1f, want ~0", d)
	}

	// Due east
	d = Direction(60.0, 25.0, 60.0, 26.0)
	if d < 85 || d > 95 {
		t.Errorf("direction east: got %.1f, want ~90", d)
	}
}
