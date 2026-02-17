package fap

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// errAmbiguityInvalid is an internal sentinel for ambiguity validation failures.
var errAmbiguityInvalid = errors.New("invalid position ambiguity")

// parsePositionNoTimestamp parses position packets without timestamps (! and =).
func (p *Packet) parsePositionNoTimestamp(opt *options, typeChar byte) error {
	p.Type = PacketTypeLocation

	// '=' indicates messaging capability
	if typeChar == '=' {
		p.Messaging = new(true)
	} else {
		p.Messaging = new(false)
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
func (p *Packet) parsePositionWithTimestamp(opt *options, typeChar byte) error {
	p.Type = PacketTypeLocation

	if typeChar == '@' {
		p.Messaging = new(true)
	} else {
		p.Messaging = new(false)
	}

	body := p.Body[1:] // skip type character

	if len(body) < 7 {
		return p.fail(ErrPosShort, "position body too short for timestamp")
	}

	// Parse timestamp (7 characters)
	if opt.rawTimestamp {
		p.RawTimestamp = body[:6] // strip the indicator char
	} else {
		ts, err := parseTimestamp(body[:7])
		if err != nil {
			p.warn(ErrTimestampInvalid, fmt.Sprintf("invalid timestamp: %v", err))
		} else {
			p.Timestamp = ts
		}
	}

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
func (p *Packet) parseUncompressedPosition(body string, opt *options) error {
	p.Format = FormatUncompressed

	if len(body) < 19 {
		return p.fail(ErrPosShort, "uncompressed position too short")
	}

	// Parse latitude: DDMM.MMN
	lat, ambiguity, err := parseUncompressedLat(body[:8])
	if err != nil {
		if errors.Is(err, errAmbiguityInvalid) {
			return p.fail(ErrPosAmbiguity, err.Error())
		}
		return p.fail(ErrLocInvalid, fmt.Sprintf("invalid latitude: %v", err))
	}

	p.Latitude = &lat
	p.PosAmbiguity = &ambiguity

	// Symbol table character
	p.SymbolTable = body[8]

	// Validate symbol table
	if !isValidSymbolTable(p.SymbolTable) {
		return p.fail(ErrSymInvTable, fmt.Sprintf("invalid symbol table: %c", p.SymbolTable))
	}

	// Parse longitude: DDDMM.MMW
	lon, err := parseUncompressedLon(body[9:18], ambiguity)
	if err != nil {
		if errors.Is(err, errAmbiguityInvalid) {
			return p.fail(ErrPosAmbiguity, err.Error())
		}
		return p.fail(ErrPosLonInvalid, fmt.Sprintf("invalid longitude: %v", err))
	}
	p.Longitude = &lon

	// Symbol code
	p.SymbolCode = body[18]

	// Position resolution based on ambiguity
	res := posResolution(ambiguity)
	p.PosResolution = &res

	// Parse the rest (comment, PHG, altitude, weather, etc.)
	if len(body) > 19 {
		p.parsePositionComment(body[19:])
	}

	return nil
}

// isCompressedTableChar checks if the first character after '!' could be a
// compressed position symbol table identifier: /\A-Za-j
func isCompressedTableChar(c byte) bool {
	return c == '/' || c == '\\' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'j')
}

// isValidSymbolTable checks if a symbol table character is valid.
func isValidSymbolTable(c byte) bool {
	return c == '/' || c == '\\' || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// parseCompressedPosition parses a compressed position report.
func (p *Packet) parseCompressedPosition(body string, opt *options) error {
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

	// GPS fix status — only set if csT bytes are used (c1 != -1)
	if c1 != -1 {
		if compType&0x20 != 0 {
			fix := 1
			p.GPSFixStatus = &fix
		} else {
			fix := 0
			p.GPSFixStatus = &fix
		}
	}

	// Position resolution for compressed: 0.291 meters
	res := 0.291
	p.PosResolution = &res

	// Decode course/speed, altitude, or radio range.
	// c1 == -1 (space) or s1 == -1 (space) means csT is not used.
	if c1 == -1 || s1 == -1 {
		// csT not used — no speed/course/altitude/range
	} else if compType&0x18 == 0x10 {
		// Altitude mode (GPGGA source)
		cs := c1*91 + s1
		alt := math.Pow(1.002, float64(cs)) * 0.3048
		p.Altitude = &alt
	} else if c1 >= 0 && c1 <= 89 {
		// Course/speed
		course := c1 * 4
		if c1 == 0 {
			course = 360 // north (0 means unknown in APRS, 360 means north)
		}
		p.Course = &course
		speed := (math.Pow(1.08, float64(s1)) - 1.0) * 1.852 // knots to km/h
		p.Speed = &speed
	} else if c1 == 90 {
		// Radio range
		rng := 2.0 * math.Pow(1.08, float64(s1)) * 1.609344 // miles to km
		p.RadioRange = &rng
	}

	// Comment after compressed position
	if len(body) > 13 {
		comment := body[13:]

		// If symbol is weather, parse weather from comment.
		if p.SymbolCode == '_' {
			p.Type = PacketTypeWx
			wx := &Weather{}
			p.Wx = wx
			wxComment := parseWeatherFromComment(comment, wx)
			if wx.hasData() && wxComment != "" {
				p.Comment = wxComment
			}
			return nil
		}

		// Strip inline telemetry |...|
		comment = stripInlineTelemetry(comment)

		// Check for DAO extension
		comment = p.parseDAO(comment)

		p.Comment = strings.TrimSpace(comment)
	}

	return nil
}

// stripInlineTelemetry removes |...| inline telemetry from comments.
func stripInlineTelemetry(comment string) string {
	// Look for |...| at the end of the comment
	if idx := strings.LastIndex(comment, "|"); idx > 0 {
		firstIdx := strings.Index(comment, "|")
		if firstIdx < idx {
			// Remove the telemetry section
			comment = comment[:firstIdx] + comment[idx+1:]
		}
	}
	return comment
}

// parsePositionComment parses the comment section of an uncompressed position.
func (p *Packet) parsePositionComment(comment string) {
	// If symbol is weather ('_'), parse weather data from the comment
	if p.SymbolCode == '_' {
		p.Type = PacketTypeWx
		wx := &Weather{}
		p.Wx = wx
		wxComment := parseWeatherFromComment(comment, wx)
		// Only store remaining comment text if weather data was actually
		// found. If no weather fields matched, the entire text is garbage
		// and should be discarded (matching Perl FAP behavior).
		if wx.hasData() && wxComment != "" {
			p.Comment = wxComment
		}
		return
	}

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

	// Check for DAO extension: !Wxx! or similar
	comment = p.parseDAO(comment)

	p.Comment = strings.TrimSpace(comment)
}

// parseDAO parses DAO extensions from comments.
// Returns the comment with the DAO extension removed.
func (p *Packet) parseDAO(comment string) string {
	// Look for !Dxx! pattern (DAO extension)
	for i := 0; i+4 < len(comment); i++ {
		if comment[i] == '!' && comment[i+4] == '!' {
			datumByte := comment[i+1]
			d1 := comment[i+2]
			d2 := comment[i+3]

			if datumByte >= 'A' && datumByte <= 'Z' {
				// Human-readable DAO (digits 0-9)
				if d1 >= '0' && d1 <= '9' && d2 >= '0' && d2 <= '9' {
					p.DaoDatumByte = datumByte
					p.applyHumanDAO(d1, d2)
					comment = comment[:i] + comment[i+5:]
					break
				}
			}
			if datumByte >= 'a' && datumByte <= 'z' {
				// Base-91 DAO
				if d1 >= '!' && d1 <= '{' && d2 >= '!' && d2 <= '{' {
					p.DaoDatumByte = datumByte - 32 // uppercase
					p.applyBase91DAO(d1, d2)
					comment = comment[:i] + comment[i+5:]
					break
				}
			}
		}
	}
	return comment
}

// applyHumanDAO applies human-readable DAO adjustments to position.
func (p *Packet) applyHumanDAO(d1, d2 byte) {
	if p.Latitude != nil && p.Longitude != nil {
		latAdd := float64(d1-'0') * 0.001 / 60.0
		lonAdd := float64(d2-'0') * 0.001 / 60.0
		if *p.Latitude < 0 {
			*p.Latitude -= latAdd
		} else {
			*p.Latitude += latAdd
		}
		if *p.Longitude < 0 {
			*p.Longitude -= lonAdd
		} else {
			*p.Longitude += lonAdd
		}
		res := 1.852 // 0.001 minute = 1.852m
		p.PosResolution = &res
	}
}

// applyBase91DAO applies base-91 DAO adjustments to position.
func (p *Packet) applyBase91DAO(d1, d2 byte) {
	if p.Latitude != nil && p.Longitude != nil {
		latAdd := float64(d1-33) / 91.0 * 0.01 / 60.0
		lonAdd := float64(d2-33) / 91.0 * 0.01 / 60.0
		if *p.Latitude < 0 {
			*p.Latitude -= latAdd
		} else {
			*p.Latitude += latAdd
		}
		if *p.Longitude < 0 {
			*p.Longitude -= lonAdd
		} else {
			*p.Longitude += lonAdd
		}
		res := 0.1852 // 0.0001 minute
		p.PosResolution = &res
	}
}

// parsePositionFallback tries a last-resort position parse (looking for '!' in body).
func (p *Packet) parsePositionFallback(opt *options) error {
	idx := strings.IndexByte(p.Body, '!')
	if idx < 0 || idx > 39 {
		return p.fail(ErrTypeNotSupported, "unsupported packet type")
	}

	body := p.Body[idx+1:]

	// Check minimum length requirements and dispatch to the right parser.
	// Compressed positions need at least 13 characters, uncompressed need 19.
	if len(body) > 0 && (body[0] >= '0' && body[0] <= '9' || body[0] == ' ') {
		if len(body) < 19 {
			return p.fail(ErrTypeNotSupported, "unsupported packet type")
		}
		p.Type = PacketTypeLocation
		p.Messaging = new(false)
		return p.parseUncompressedPosition(body, opt)
	}

	if len(body) < 13 || !isCompressedTableChar(body[0]) {
		return p.fail(ErrTypeNotSupported, "unsupported packet type")
	}
	p.Type = PacketTypeLocation
	p.Messaging = new(false)
	return p.parseCompressedPosition(body, opt)
}

// parseUncompressedLat parses an uncompressed latitude string "DDMM.MMN".
// Returns latitude in decimal degrees and ambiguity level.
// digitOrSpace reads a digit from s[i], treating space as 0.
// Returns the digit value and whether it was a space (for ambiguity counting).
func digitOrSpace(s string, i int) (int, bool, error) {
	c := s[i]
	if c == ' ' {
		return 0, true, nil
	}
	if c >= '0' && c <= '9' {
		return int(c - '0'), false, nil
	}
	return 0, false, fmt.Errorf("invalid character at position %d: %c", i, c)
}

// parseDegreesMinutes parses a DDMM.MM or DDDMM.MM string directly from characters,
// counting ambiguity (trailing spaces in minute digits, right to left).
// degDigits is 2 for latitude, 3 for longitude.
// For longitude, ambiguity is passed in from latitude; for latitude it is computed.
func parseDegreesMinutes(s string, degDigits int, computeAmbiguity bool, knownAmbiguity int) (float64, float64, int, error) {
	// Parse degree digits
	deg := 0.0
	for i := 0; i < degDigits; i++ {
		d, _, err := digitOrSpace(s, i)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid degrees: %v", err)
		}
		deg = deg*10 + float64(d)
	}

	// Format after degrees: MM.MM
	// Minute digit positions relative to degDigits: +0, +1, +2(dot), +3, +4
	dotPos := degDigits + 2
	if s[dotPos] != '.' && s[dotPos] != ' ' {
		return 0, 0, 0, fmt.Errorf("expected dot at position %d, got %c", dotPos, s[dotPos])
	}

	// Parse the 4 minute digits (2 before dot, 2 after)
	// Positions: degDigits, degDigits+1, degDigits+3, degDigits+4
	minPositions := []int{degDigits, degDigits + 1, degDigits + 3, degDigits + 4}
	var minDigits [4]int
	var isSpace [4]bool
	for i, pos := range minPositions {
		d, sp, err := digitOrSpace(s, pos)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid minutes: %v", err)
		}
		minDigits[i] = d
		isSpace[i] = sp
	}

	// Count ambiguity: spaces from right
	ambiguity := knownAmbiguity
	if computeAmbiguity {
		ambiguity = 0
		for i := 3; i >= 0; i-- {
			if isSpace[i] {
				ambiguity++
			} else {
				break
			}
		}
		// Verify no spaces exist before the trailing ambiguity block
		for i := 0; i < 4-ambiguity; i++ {
			if isSpace[i] {
				return 0, 0, 0, fmt.Errorf("%w: space in non-trailing position", errAmbiguityInvalid)
			}
		}
	} else {
		// Longitude: the non-ambiguous portion must not contain spaces.
		// The ambiguous portion (trailing digits) may or may not have spaces —
		// APRS101 says ambiguity from latitude applies automatically to longitude.
		for i := 0; i < 4-ambiguity; i++ {
			if isSpace[i] {
				return 0, 0, 0, fmt.Errorf("%w: longitude has spaces in non-ambiguous digits", errAmbiguityInvalid)
			}
		}
	}

	// Build minutes as MM.MM
	mm := float64(minDigits[0])*10 + float64(minDigits[1]) + float64(minDigits[2])*0.1 + float64(minDigits[3])*0.01

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
			deg = math.Floor(deg)
			mm = 30
		}
	}

	return deg, mm, ambiguity, nil
}

func parseUncompressedLat(s string) (float64, int, error) {
	if len(s) != 8 {
		return 0, 0, fmt.Errorf("latitude must be 8 characters, got %d", len(s))
	}

	hemisphere := s[7]
	if hemisphere != 'N' && hemisphere != 'S' {
		return 0, 0, fmt.Errorf("invalid hemisphere: %c", hemisphere)
	}

	dd, mm, ambiguity, err := parseDegreesMinutes(s[:7], 2, true, 0)
	if err != nil {
		return 0, 0, err
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

	ddd, mm, _, err := parseDegreesMinutes(s[:8], 3, false, ambiguity)
	if err != nil {
		return 0, err
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
