package fap

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// parseNMEA parses NMEA GPS data packets.
// Supported: $GPRMC, $GPGGA, $GPGLL
func (p *Packet) parseNMEA(opt Options) error {
	p.Type = PacketTypeLocation
	p.Format = FormatNMEA

	body := p.Body

	// Must start with $GP
	if !strings.HasPrefix(body, "$GP") {
		return p.fail(ErrNMEAInvalid, "NMEA sentence must start with $GP")
	}

	// Verify and remove checksum if present
	body = strings.TrimRight(body, " \t\r\n")
	if idx := strings.IndexByte(body, '*'); idx >= 0 {
		checksumStr := body[idx+1:]
		checksumArea := body[:idx]
		if len(checksumStr) == 2 {
			given, err := strconv.ParseUint(checksumStr, 16, 8)
			if err == nil {
				var calculated uint8
				// NMEA checksum covers everything between $ and *, exclusive
				start := 0
				if len(checksumArea) > 0 && checksumArea[0] == '$' {
					start = 1
				}
				for i := start; i < len(checksumArea); i++ {
					calculated ^= checksumArea[i]
				}
				if uint8(given) != calculated {
					return p.fail(ErrNMEAInvalid, "NMEA checksum mismatch")
				}
				ok := true
				p.ChecksumOK = &ok
			}
		}
		body = checksumArea
	}

	parts := strings.Split(body, ",")
	if len(parts) < 2 {
		return p.fail(ErrNMEAShort, "NMEA sentence too short")
	}

	sentence := parts[0]

	switch sentence {
	case "$GPRMC":
		return p.parseGPRMC(parts)
	case "$GPGGA":
		return p.parseGPGGA(parts)
	case "$GPGLL":
		return p.parseGPGLL(parts)
	default:
		return p.fail(ErrNMEAInvalid, fmt.Sprintf("unsupported NMEA sentence: %s", sentence))
	}
}

// parseGPRMC parses a GPRMC sentence.
// Format: $GPRMC,HHMMSS,A,DDMM.MMM,N,DDDMM.MMM,W,speed,course,DDMMYY,...
func (p *Packet) parseGPRMC(parts []string) error {
	if len(parts) < 10 {
		return p.fail(ErrNMEAShort, "GPRMC sentence too short")
	}

	// Status check
	if parts[2] != "A" {
		return p.fail(ErrNMEAInvalid, "GPRMC: no valid fix")
	}

	// Timestamp from time (HHMMSS) and date (DDMMYY) fields
	if err := p.parseGPRMCTimestamp(parts[1], parts[9]); err != nil {
		return err
	}

	// Latitude
	lat, latRes, err := parseNMEACoordWithRes(parts[3], parts[4], false)
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("GPRMC: %v", err))
	}
	p.Latitude = &lat

	// Longitude
	lon, lonRes, err := parseNMEACoordWithRes(parts[5], parts[6], true)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("GPRMC: %v", err))
	}
	p.Longitude = &lon

	// Use the worse (larger) resolution of lat/lon
	res := latRes
	if lonRes > res {
		res = lonRes
	}
	p.PosResolution = &res

	// Speed (knots to km/h)
	if parts[7] != "" {
		speed, err := strconv.ParseFloat(parts[7], 64)
		if err == nil {
			speed *= 1.852
			p.Speed = &speed
		}
	}

	// Course
	if parts[8] != "" {
		course, err := strconv.ParseFloat(parts[8], 64)
		if err == nil {
			c := int(course + 0.5)
			if c == 0 {
				c = 360
			} else if c > 360 {
				c = 0
			}
			p.Course = &c
		}
	} else {
		c := 0
		p.Course = &c
	}

	return nil
}

// parseGPRMCTimestamp parses GPRMC time (HHMMSS) and date (DDMMYY) into a timestamp.
func (p *Packet) parseGPRMCTimestamp(timeStr, dateStr string) error {
	// Parse time: HHMMSS (possibly with decimal seconds)
	timeStr = strings.TrimSpace(timeStr)
	// Remove decimal part if present
	if idx := strings.IndexByte(timeStr, '.'); idx >= 0 {
		timeStr = timeStr[:idx]
	}
	if len(timeStr) != 6 {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid time")
	}
	hour, err := strconv.Atoi(timeStr[0:2])
	if err != nil || hour > 23 {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid time")
	}
	minute, err := strconv.Atoi(timeStr[2:4])
	if err != nil || minute > 59 {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid time")
	}
	second, err := strconv.Atoi(timeStr[4:6])
	if err != nil || second > 59 {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid time")
	}

	// Parse date: DDMMYY
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) != 6 {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid date")
	}
	day, err := strconv.Atoi(dateStr[0:2])
	if err != nil {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid date")
	}
	month, err := strconv.Atoi(dateStr[2:4])
	if err != nil {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid date")
	}
	yy, err := strconv.Atoi(dateStr[4:6])
	if err != nil {
		return p.fail(ErrNMEAInvalid, "GPRMC: invalid date")
	}

	year := 2000 + yy
	if yy >= 70 {
		year = 1900 + yy
	}

	ts := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	p.Timestamp = &ts

	return nil
}

// parseGPGGA parses a GPGGA sentence.
// Format: $GPGGA,HHMMSS,DDMM.MMM,N,DDDMM.MMM,W,quality,sats,HDOP,alt,M,...
func (p *Packet) parseGPGGA(parts []string) error {
	if len(parts) < 11 {
		return p.fail(ErrNMEAShort, "GPGGA sentence too short")
	}

	// Fix quality check
	if parts[6] == "0" {
		return p.fail(ErrNMEAInvalid, "GPGGA: no valid fix")
	}

	// Latitude
	lat, latRes, err := parseNMEACoordWithRes(parts[2], parts[3], false)
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("GPGGA: %v", err))
	}
	p.Latitude = &lat

	// Longitude
	lon, lonRes, err := parseNMEACoordWithRes(parts[4], parts[5], true)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("GPGGA: %v", err))
	}
	p.Longitude = &lon

	// Position resolution
	res := latRes
	if lonRes > res {
		res = lonRes
	}
	p.PosResolution = &res

	// Altitude
	if parts[9] != "" {
		alt, err := strconv.ParseFloat(parts[9], 64)
		if err == nil {
			p.Altitude = &alt
		}
	}

	return nil
}

// parseGPGLL parses a GPGLL sentence.
// Format: $GPGLL,DDMM.MMM,N,DDDMM.MMM,W,HHMMSS,A
func (p *Packet) parseGPGLL(parts []string) error {
	if len(parts) < 5 {
		return p.fail(ErrNMEAShort, "GPGLL sentence too short")
	}

	// Latitude
	lat, latRes, err := parseNMEACoordWithRes(parts[1], parts[2], false)
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("GPGLL: %v", err))
	}
	p.Latitude = &lat

	// Longitude
	lon, lonRes, err := parseNMEACoordWithRes(parts[3], parts[4], true)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("GPGLL: %v", err))
	}
	p.Longitude = &lon

	// Position resolution
	res := latRes
	if lonRes > res {
		res = lonRes
	}
	p.PosResolution = &res

	return nil
}

// nmeaPosResolution returns position resolution in meters based on the number
// of minute decimal digits. Matches Perl's _get_posresolution().
func nmeaPosResolution(decimals int) float64 {
	base := 1000.0
	if decimals <= -2 {
		base = 600.0
	}
	return 1.852 * base * math.Pow(10, float64(-decimals))
}

// parseNMEACoordWithRes parses an NMEA coordinate and returns the position resolution.
func parseNMEACoordWithRes(coord, hemisphere string, isLon bool) (float64, float64, error) {
	if coord == "" || hemisphere == "" {
		return 0, 0, fmt.Errorf("empty coordinate or hemisphere")
	}

	var degLen int
	if isLon {
		degLen = 3
	} else {
		degLen = 2
	}

	if len(coord) < degLen+1 {
		return 0, 0, fmt.Errorf("coordinate too short")
	}

	deg, err := strconv.ParseFloat(coord[:degLen], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid degrees: %v", err)
	}

	min, err := strconv.ParseFloat(coord[degLen:], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minutes: %v", err)
	}

	result := deg + min/60.0

	if hemisphere == "S" || hemisphere == "W" {
		result = -result
	}

	// Range check
	maxVal := 90.0
	if isLon {
		maxVal = 180.0
	}
	if math.Abs(result) > maxVal {
		return 0, 0, fmt.Errorf("coordinate out of range: %f", result)
	}

	// Calculate position resolution based on decimal places in minutes
	decimals := 0
	if dotIdx := strings.IndexByte(coord[degLen:], '.'); dotIdx >= 0 {
		decimals = len(coord[degLen:]) - dotIdx - 1
	}
	res := nmeaPosResolution(decimals)

	return result, res, nil
}
