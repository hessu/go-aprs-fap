package fap

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// parsePositionNoTimestamp parses position packets without timestamps (! and =).
func (p *Packet) parsePositionNoTimestamp(opt Options, typeChar byte) error {
	p.Type = PacketTypeLocation

	// '=' indicates messaging capability
	if typeChar == '=' {
		t := true
		p.Messaging = &t
	} else {
		f := false
		p.Messaging = &f
	}

	body := p.Body[1:] // skip type character

	if len(body) == 0 {
		return p.fail(ErrPosShort, "position body too short")
	}

	// Determine if compressed or uncompressed
	if body[0] >= '0' && body[0] <= '9' || body[0] == ' ' {
		return p.parseUncompressedPosition(body, opt)
	}
	return p.parseCompressedPosition(body, opt)
}

// parsePositionWithTimestamp parses position packets with timestamps (/ and @).
func (p *Packet) parsePositionWithTimestamp(opt Options, typeChar byte) error {
	p.Type = PacketTypeLocation

	if typeChar == '@' {
		t := true
		p.Messaging = &t
	} else {
		f := false
		p.Messaging = &f
	}

	body := p.Body[1:] // skip type character

	if len(body) < 7 {
		return p.fail(ErrPosShort, "position body too short for timestamp")
	}

	// Parse timestamp (7 characters)
	ts, err := parseTimestamp(body[:7])
	if err != nil {
		return p.fail(ErrTimestampInvalid, fmt.Sprintf("invalid timestamp: %v", err))
	}
	p.Timestamp = ts

	posBody := body[7:]
	if len(posBody) == 0 {
		return p.fail(ErrPosShort, "no position data after timestamp")
	}

	if posBody[0] >= '0' && posBody[0] <= '9' || posBody[0] == ' ' {
		return p.parseUncompressedPosition(posBody, opt)
	}
	return p.parseCompressedPosition(posBody, opt)
}

// parseUncompressedPosition parses an uncompressed position report.
func (p *Packet) parseUncompressedPosition(body string, opt Options) error {
	p.Format = FormatUncompressed

	if len(body) < 19 {
		return p.fail(ErrPosShort, "uncompressed position too short")
	}

	// Parse latitude: DDMM.MMN
	lat, ambiguity, err := parseUncompressedLat(body[:8])
	if err != nil {
		return p.fail(ErrPosLatInvalid, fmt.Sprintf("invalid latitude: %v", err))
	}

	p.Latitude = &lat
	p.PosAmbiguity = &ambiguity

	// Symbol table character
	p.SymbolTable = body[8]

	// Parse longitude: DDDMM.MMW
	lon, err := parseUncompressedLon(body[9:18], ambiguity)
	if err != nil {
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("invalid longitude: %v", err))
	}
	p.Longitude = &lon

	// Symbol code
	p.SymbolCode = body[18]

	// Position resolution based on ambiguity
	res := posResolution(ambiguity)
	p.PosResolution = &res

	// Parse the rest (comment, PHG, altitude, etc.)
	if len(body) > 19 {
		p.parsePositionComment(body[19:])
	}

	return nil
}

// parseCompressedPosition parses a compressed position report.
func (p *Packet) parseCompressedPosition(body string, opt Options) error {
	p.Format = FormatCompressed

	if len(body) < 13 {
		return p.fail(ErrCompShort, "compressed position too short")
	}

	p.SymbolTable = body[0]

	// Decode latitude (4 bytes, base-91)
	lat := 90.0 - float64(
		(int(body[1])-33)*91*91*91+
			(int(body[2])-33)*91*91+
			(int(body[3])-33)*91+
			(int(body[4])-33),
	)/380926.0

	// Decode longitude (4 bytes, base-91)
	lon := -180.0 + float64(
		(int(body[5])-33)*91*91*91+
			(int(body[6])-33)*91*91+
			(int(body[7])-33)*91+
			(int(body[8])-33),
	)/190463.0

	p.Latitude = &lat
	p.Longitude = &lon

	p.SymbolCode = body[9]

	// Course/speed or altitude in bytes 10-11
	c1 := int(body[10]) - 33
	s1 := int(body[11]) - 33

	// Compression type byte
	compType := int(body[12]) - 33

	// GPS fix status
	if compType&0x20 != 0 {
		fix := 1
		p.GPSFixStatus = &fix
	} else {
		fix := 0
		p.GPSFixStatus = &fix
	}

	// Position resolution for compressed: 0.291 meters
	res := 0.291
	p.PosResolution = &res

	amb := 0
	p.PosAmbiguity = &amb

	// Decode course/speed or altitude
	if c1 >= 0 && c1 <= 89 {
		course := c1 * 4
		p.Course = &course
		speed := math.Pow(1.08, float64(s1)) - 1.0
		speed *= 1.852 // knots to km/h
		p.Speed = &speed
	} else if c1 == 90 {
		// Altitude
		alt := math.Pow(1.002, float64(c1*91+s1))
		alt *= 0.3048 // feet to meters
		p.Altitude = &alt
	}

	// Radio range
	if compType&0x18 == 0x10 && c1 >= 0 && c1 <= 89 {
		rng := 2.0 * math.Pow(1.08, float64(s1))
		rng *= 1.609344 // miles to km
		p.RadioRange = &rng
	}

	// Comment after compressed position
	if len(body) > 13 {
		p.Comment = strings.TrimSpace(body[13:])
	}

	return nil
}

// parsePositionComment parses the comment section of an uncompressed position.
func (p *Packet) parsePositionComment(comment string) {
	// Check for PHG data
	if strings.HasPrefix(comment, "PHG") && len(comment) >= 7 {
		p.PHG = comment[3:7]
		comment = comment[7:]
		// Skip separator after PHG
		if len(comment) > 0 && comment[0] == '/' {
			comment = comment[1:]
		}
	}

	// Check for course/speed: CCC/SSS
	if len(comment) >= 7 && comment[3] == '/' {
		courseStr := comment[0:3]
		speedStr := comment[4:7]
		course, cerr := strconv.Atoi(courseStr)
		speed, serr := strconv.Atoi(speedStr)
		if cerr == nil && serr == nil && course >= 0 && course <= 360 {
			p.Course = &course
			speedKmh := float64(speed) * 1.852 // knots to km/h
			p.Speed = &speedKmh
			comment = comment[7:]
		}
	}

	// Check for altitude: /A=NNNNNN
	if idx := strings.Index(comment, "/A="); idx >= 0 && len(comment) >= idx+9 {
		altStr := comment[idx+3 : idx+9]
		if alt, err := strconv.Atoi(altStr); err == nil {
			altM := float64(alt) * 0.3048 // feet to meters
			p.Altitude = &altM
			comment = comment[:idx] + comment[idx+9:]
		}
	}

	// If symbol is weather ('_'), try to parse weather data
	if p.SymbolCode == '_' {
		// Weather station with comment instead of weather data -
		// don't store the comment to avoid confusion
		return
	}

	p.Comment = strings.TrimSpace(comment)
}

// parsePositionFallback tries a last-resort position parse (looking for '!' in body).
func (p *Packet) parsePositionFallback(opt Options) error {
	idx := strings.IndexByte(p.Body, '!')
	if idx < 0 {
		return p.fail(ErrTypeNotSupported, "unsupported packet type")
	}

	p.Type = PacketTypeLocation
	f := false
	p.Messaging = &f

	body := p.Body[idx+1:]
	if len(body) == 0 {
		return p.fail(ErrPosShort, "position body too short")
	}

	if body[0] >= '0' && body[0] <= '9' || body[0] == ' ' {
		return p.parseUncompressedPosition(body, opt)
	}
	return p.parseCompressedPosition(body, opt)
}

// parseUncompressedLat parses an uncompressed latitude string "DDMM.MMN".
// Returns latitude in decimal degrees and ambiguity level.
func parseUncompressedLat(s string) (float64, int, error) {
	if len(s) != 8 {
		return 0, 0, fmt.Errorf("latitude must be 8 characters, got %d", len(s))
	}

	hemisphere := s[7]
	if hemisphere != 'N' && hemisphere != 'S' {
		return 0, 0, fmt.Errorf("invalid hemisphere: %c", hemisphere)
	}

	// Count ambiguity (spaces in the numeric portion)
	ambiguity := 0
	latStr := []byte(s[:7])
	// Check from right to left for spaces
	// Format: DDMM.MM - positions 6,5,(skip dot at 4),3,2 can be spaces
	positions := []int{6, 5, 3, 2}
	for _, pos := range positions {
		if latStr[pos] == ' ' {
			ambiguity++
			latStr[pos] = '0'
		} else {
			break
		}
	}

	// Also handle the dot position being a space (shouldn't happen but be safe)
	if latStr[4] == ' ' {
		latStr[4] = '.'
	}

	str := string(latStr)
	dd, err := strconv.ParseFloat(str[:2], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid degrees: %v", err)
	}
	mm, err := strconv.ParseFloat(str[2:], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minutes: %v", err)
	}

	// For ambiguous positions, center in the ambiguity box
	if ambiguity > 0 {
		switch ambiguity {
		case 1:
			mm = math.Floor(mm/0.1)*0.1 + 0.05
		case 2:
			mm = math.Floor(mm) + 0.5
		case 3:
			mm = math.Floor(mm/10)*10 + 5
		case 4:
			dd = math.Floor(dd)
			mm = 30
		}
	}

	lat := dd + mm/60.0

	if lat > 90.0 || lat < 0.0 {
		return 0, 0, fmt.Errorf("latitude out of range: %f", lat)
	}

	if hemisphere == 'S' {
		lat = -lat
	}

	return lat, ambiguity, nil
}

// parseUncompressedLon parses an uncompressed longitude string "DDDMM.MMW".
func parseUncompressedLon(s string, ambiguity int) (float64, error) {
	if len(s) != 9 {
		return 0, fmt.Errorf("longitude must be 9 characters, got %d", len(s))
	}

	hemisphere := s[8]
	if hemisphere != 'E' && hemisphere != 'W' {
		return 0, fmt.Errorf("invalid hemisphere: %c", hemisphere)
	}

	// Apply same ambiguity as latitude
	lonStr := []byte(s[:8])
	positions := []int{7, 6, 4, 3}
	for i := 0; i < ambiguity && i < len(positions); i++ {
		lonStr[positions[i]] = '0'
	}

	if lonStr[5] == ' ' {
		lonStr[5] = '.'
	}

	str := string(lonStr)
	ddd, err := strconv.ParseFloat(str[:3], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid degrees: %v", err)
	}
	mm, err := strconv.ParseFloat(str[3:], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %v", err)
	}

	if ambiguity > 0 {
		switch ambiguity {
		case 1:
			mm = math.Floor(mm/0.1)*0.1 + 0.05
		case 2:
			mm = math.Floor(mm) + 0.5
		case 3:
			mm = math.Floor(mm/10)*10 + 5
		case 4:
			ddd = math.Floor(ddd)
			mm = 30
		}
	}

	lon := ddd + mm/60.0

	if lon > 180.0 || lon < 0.0 {
		return 0, fmt.Errorf("longitude out of range: %f", lon)
	}

	if hemisphere == 'W' {
		lon = -lon
	}

	return lon, nil
}

// posResolution returns the position resolution in meters for a given ambiguity level.
func posResolution(ambiguity int) float64 {
	switch ambiguity {
	case 0:
		return 1852.0 / 100.0 // ~18.52m (0.01 minute)
	case 1:
		return 1852.0 / 10.0 // ~185.2m (0.1 minute)
	case 2:
		return 1852.0 // ~1852m (1 minute)
	case 3:
		return 1852.0 * 10.0 // ~18520m (10 minutes)
	case 4:
		return 1852.0 * 60.0 // ~111120m (1 degree)
	default:
		return 1852.0 / 100.0
	}
}
