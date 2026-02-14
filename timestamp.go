package fap

import (
	"fmt"
	"strconv"
	"time"
)

// parseTimestamp parses a 7-character APRS timestamp.
// Formats:
//   - DDHHMMz - Day/Hours/Minutes in UTC
//   - DDHHMM/ - Day/Hours/Minutes in local time
//   - HHMMSSh - Hours/Minutes/Seconds in UTC
func parseTimestamp(s string) (*time.Time, error) {
	if len(s) != 7 {
		return nil, fmt.Errorf("timestamp must be 7 characters, got %d", len(s))
	}

	indicator := s[6]

	switch indicator {
	case 'z':
		// DDHHMMz - day/hours/minutes UTC
		dd, err := strconv.Atoi(s[0:2])
		if err != nil || dd < 1 || dd > 31 {
			return nil, fmt.Errorf("invalid day: %s", s[0:2])
		}
		hh, err := strconv.Atoi(s[2:4])
		if err != nil || hh > 23 {
			return nil, fmt.Errorf("invalid hours: %s", s[2:4])
		}
		mm, err := strconv.Atoi(s[4:6])
		if err != nil || mm > 59 {
			return nil, fmt.Errorf("invalid minutes: %s", s[4:6])
		}

		now := time.Now().UTC()
		t := time.Date(now.Year(), now.Month(), dd, hh, mm, 0, 0, time.UTC)

		// If the timestamp is in the future, assume it's from last month
		if t.After(now) {
			t = t.AddDate(0, -1, 0)
		}

		return &t, nil

	case '/':
		// DDHHMM/ - day/hours/minutes local time
		dd, err := strconv.Atoi(s[0:2])
		if err != nil || dd < 1 || dd > 31 {
			return nil, fmt.Errorf("invalid day: %s", s[0:2])
		}
		hh, err := strconv.Atoi(s[2:4])
		if err != nil || hh > 23 {
			return nil, fmt.Errorf("invalid hours: %s", s[2:4])
		}
		mm, err := strconv.Atoi(s[4:6])
		if err != nil || mm > 59 {
			return nil, fmt.Errorf("invalid minutes: %s", s[4:6])
		}

		now := time.Now()
		loc := now.Location()
		t := time.Date(now.Year(), now.Month(), dd, hh, mm, 0, 0, loc)

		// If the timestamp is in the future, assume it's from last month
		if t.After(now) {
			t = t.AddDate(0, -1, 0)
		}

		return &t, nil

	case 'h':
		// HHMMSSh - hours/minutes/seconds UTC
		hh, err := strconv.Atoi(s[0:2])
		if err != nil || hh > 23 {
			return nil, fmt.Errorf("invalid hours: %s", s[0:2])
		}
		mm, err := strconv.Atoi(s[2:4])
		if err != nil || mm > 59 {
			return nil, fmt.Errorf("invalid minutes: %s", s[2:4])
		}
		ss, err := strconv.Atoi(s[4:6])
		if err != nil || ss > 59 {
			return nil, fmt.Errorf("invalid seconds: %s", s[4:6])
		}

		now := time.Now().UTC()
		t := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, 0, time.UTC)

		return &t, nil

	default:
		return nil, fmt.Errorf("unknown timestamp indicator: %c", indicator)
	}
}
