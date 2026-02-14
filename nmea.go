package fap

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

	// Remove checksum if present
	if idx := strings.IndexByte(body, '*'); idx >= 0 {
		body = body[:idx]
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

	// Latitude
	lat, err := parseNMEACoord(parts[3], parts[4], false)
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("GPRMC: %v", err))
	}
	p.Latitude = &lat

	// Longitude
	lon, err := parseNMEACoord(parts[5], parts[6], true)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("GPRMC: %v", err))
	}
	p.Longitude = &lon

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
			c := int(course)
			p.Course = &c
		}
	}

	amb := 0
	p.PosAmbiguity = &amb
	res := posResolution(0)
	p.PosResolution = &res

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
	lat, err := parseNMEACoord(parts[2], parts[3], false)
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("GPGGA: %v", err))
	}
	p.Latitude = &lat

	// Longitude
	lon, err := parseNMEACoord(parts[4], parts[5], true)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("GPGGA: %v", err))
	}
	p.Longitude = &lon

	// Altitude
	if parts[9] != "" {
		alt, err := strconv.ParseFloat(parts[9], 64)
		if err == nil {
			p.Altitude = &alt
		}
	}

	amb := 0
	p.PosAmbiguity = &amb
	res := posResolution(0)
	p.PosResolution = &res

	return nil
}

// parseGPGLL parses a GPGLL sentence.
// Format: $GPGLL,DDMM.MMM,N,DDDMM.MMM,W,HHMMSS,A
func (p *Packet) parseGPGLL(parts []string) error {
	if len(parts) < 5 {
		return p.fail(ErrNMEAShort, "GPGLL sentence too short")
	}

	// Latitude
	lat, err := parseNMEACoord(parts[1], parts[2], false)
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("GPGLL: %v", err))
	}
	p.Latitude = &lat

	// Longitude
	lon, err := parseNMEACoord(parts[3], parts[4], true)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("GPGLL: %v", err))
	}
	p.Longitude = &lon

	amb := 0
	p.PosAmbiguity = &amb
	res := posResolution(0)
	p.PosResolution = &res

	return nil
}

// parseNMEACoord parses an NMEA coordinate (latitude or longitude).
func parseNMEACoord(coord, hemisphere string, isLon bool) (float64, error) {
	if coord == "" || hemisphere == "" {
		return 0, fmt.Errorf("empty coordinate or hemisphere")
	}

	var degLen int
	if isLon {
		degLen = 3
	} else {
		degLen = 2
	}

	if len(coord) < degLen+1 {
		return 0, fmt.Errorf("coordinate too short")
	}

	deg, err := strconv.ParseFloat(coord[:degLen], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid degrees: %v", err)
	}

	min, err := strconv.ParseFloat(coord[degLen:], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %v", err)
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
		return 0, fmt.Errorf("coordinate out of range: %f", result)
	}

	return result, nil
}
