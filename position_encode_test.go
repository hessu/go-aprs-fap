package fap

import (
	"errors"
	"testing"
	"time"
)

// Tests ported from perl-aprs-fap/t/80make-position.t

func TestEncodePosition(t *testing.T) {
	tests := []struct {
		name     string
		lat, lon float64
		speed    *float64
		course   *float64
		altitude *float64
		symbol   string
		opts     *EncodePositionOpts
		want     string
	}{
		{
			"basic NE, no speed/course/alt",
			63.06716666666667, 27.6605, nil, nil, nil, "/#", nil,
			"!6304.03N/02739.63E#",
		},
		{
			"basic SW, no speed/course/alt",
			-23.64266666666667, -46.797, nil, nil, nil, "/#", nil,
			"!2338.56S/04647.82W#",
		},
		{
			"minute rounding, no speed/course/alt",
			22.9999999, -177.9999999, nil, nil, nil, "/#", nil,
			"!2259.99N/17759.99W#",
		},
		{
			"NE, has speed/course/alt",
			52.364, 14.1045, new(83.34), new(353.0), new(95.7072), "/>", nil,
			"!5221.84N/01406.27E>353/045/A=000314",
		},
		{
			"NE, no speed/course, has alt",
			52.364, 14.1045, nil, nil, new(95.7072), "/>", nil,
			"!5221.84N/01406.27E>/A=000314",
		},
		{
			"NE, ambiguity 1",
			52.364, 14.1045, nil, nil, nil, "/>", &EncodePositionOpts{Ambiguity: 1},
			"!5221.8 N/01406.2 E>",
		},
		{
			"NE, ambiguity 2",
			52.364, 14.1045, nil, nil, nil, "/>", &EncodePositionOpts{Ambiguity: 2},
			"!5221.  N/01406.  E>",
		},
		{
			"NE, ambiguity 3",
			52.364, 14.1045, nil, nil, nil, "/>", &EncodePositionOpts{Ambiguity: 3},
			"!522 .  N/0140 .  E>",
		},
		{
			"NE, ambiguity 4",
			52.364, 14.1045, nil, nil, nil, "/>", &EncodePositionOpts{Ambiguity: 4},
			"!52  .  N/014  .  E>",
		},
		{
			"DAO position, US",
			39.15380036630037, -84.62208058608059, nil, nil, nil, "/>", &EncodePositionOpts{DAO: true},
			"!3909.22N/08437.32W>!wjM!",
		},
		{
			"DAO rounding",
			39.9999999, -84.9999999, nil, nil, nil, "/>", &EncodePositionOpts{DAO: true},
			"!3959.99N/08459.99W>!w{{!",
		},
		{
			"DAO with speed/course/alt/comment",
			48.37314835164835, 15.71477838827839, new(62.968), new(321.0), new(192.9384), "/>",
			&EncodePositionOpts{DAO: true, Comment: "Comment blah"},
			"!4822.38N/01542.88E>321/034/A=000633Comment blah!wr^!",
		},
		{
			"with timestamp",
			63.06716666666667, 27.6605, nil, nil, nil, "/#",
			&EncodePositionOpts{Timestamp: time.Date(2024, 3, 15, 12, 30, 45, 0, time.UTC)},
			"/123045h6304.03N/02739.63E#",
		},
		{
			"with timestamp and speed/course/alt",
			52.364, 14.1045, new(83.34), new(353.0), new(95.7072), "/>",
			&EncodePositionOpts{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			"/000000h5221.84N/01406.27E>353/045/A=000314",
		},
		{
			"messaging capable, no timestamp",
			63.06716666666667, 27.6605, nil, nil, nil, "/#",
			&EncodePositionOpts{MessagingCapable: true},
			"=6304.03N/02739.63E#",
		},
		{
			"messaging capable, with timestamp",
			63.06716666666667, 27.6605, nil, nil, nil, "/#",
			&EncodePositionOpts{MessagingCapable: true, Timestamp: time.Date(2024, 3, 15, 12, 30, 45, 0, time.UTC)},
			"@123045h6304.03N/02739.63E#",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EncodePosition(tc.lat, tc.lon, tc.speed, tc.course, tc.altitude, tc.symbol, tc.opts)
			if err != nil {
				t.Fatalf("EncodePosition failed: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEncodePositionErrors(t *testing.T) {
	tests := []struct {
		name   string
		lat    float64
		lon    float64
		symbol string
	}{
		{"lat too high", 91.0, 0.0, "/#"},
		{"lat too low", -91.0, 0.0, "/#"},
		{"lon too high", 0.0, 181.0, "/#"},
		{"lon too low", 0.0, -181.0, "/#"},
		{"invalid symbol table", 0.0, 0.0, "a#"},
		{"invalid symbol code", 0.0, 0.0, "/\x1f"},
		{"symbol too short", 0.0, 0.0, "/"},
		{"symbol too long", 0.0, 0.0, "//#"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := EncodePosition(tc.lat, tc.lon, nil, nil, nil, tc.symbol, nil)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrPosEncInvalid) {
				t.Errorf("expected ErrPosEncInvalid, got %v", err)
			}
		})
	}
}
