package fap

import (
	"fmt"
	"strconv"
	"time"
)

// dhmFutureLimit is how far into the future a day/hours/minutes timestamp
// may be, to allow for clock differences, like in the Perl module.
const dhmFutureLimit = 43400 * time.Second

// parseTimestamp parses a 7-character APRS timestamp.
// Formats:
//   - DDHHMMz - Day/Hours/Minutes in UTC
//   - DDHHMM/ - Day/Hours/Minutes in local time
//   - HHMMSSh - Hours/Minutes/Seconds in UTC
func parseTimestamp(s string) (*time.Time, error) {
	return parseTimestampAt(s, time.Now())
}

// dhmTimestamp returns the time of a day/hours/minutes timestamp, which
// does not have the month, relative to now, in the location of now. Like
// the Perl module, it picks the latest of the next, the current and the
// previous month in which the day exists and the time is less than
// dhmFutureLimit in the future.
func dhmTimestamp(dd, hh, mm int, now time.Time) (*time.Time, error) {
	loc := now.Location()
	for _, offset := range []int{1, 0, -1} {
		first := time.Date(now.Year(), now.Month()+time.Month(offset), 1, 0, 0, 0, 0, loc)
		daysInMonth := first.AddDate(0, 1, -1).Day()
		if dd > daysInMonth {
			continue
		}
		t := time.Date(first.Year(), first.Month(), dd, hh, mm, 0, 0, loc)
		if offset == -1 || t.Sub(now) < dhmFutureLimit {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("day %d does not exist in the previous or current month", dd)
}

// parseTimestampAt parses a 7-character APRS timestamp, which is relative
// to the time now.
func parseTimestampAt(s string, now time.Time) (*time.Time, error) {
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

		return dhmTimestamp(dd, hh, mm, now.UTC())

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

		return dhmTimestamp(dd, hh, mm, now.Local())

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

		now := now.UTC()
		t := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, 0, time.UTC)

		return &t, nil

	default:
		return nil, fmt.Errorf("unknown timestamp indicator: %c", indicator)
	}
}
