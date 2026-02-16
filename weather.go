package fap

import (
	"math"
	"strconv"
	"strings"
)

// parseWeatherPositionless parses a positionless weather report.
// Format: _MMDDHHMM weather data...
func (p *Packet) parseWeatherPositionless(opt *options) error {
	p.Type = PacketTypeWx

	body := p.Body[1:] // skip '_'

	if len(body) < 8 {
		return p.fail(ErrWxInvalid, "positionless weather report too short")
	}

	// Skip the timestamp (8 characters: MMDDHHMM)
	wxData := body[8:]

	wx := &Weather{}
	p.Wx = wx

	parseWeatherFields(wxData, wx)

	return nil
}

// hasData returns true if any weather field has been populated.
func (wx *Weather) hasData() bool {
	return wx.WindDirection != nil || wx.WindSpeed != nil || wx.WindGust != nil ||
		wx.Temp != nil || wx.TempIn != nil || wx.Humidity != nil || wx.HumidityIn != nil ||
		wx.Pressure != nil || wx.Rain1h != nil || wx.Rain24h != nil || wx.RainMidnight != nil ||
		wx.Snow24h != nil || wx.Luminosity != nil
}

// parseWeatherFromComment parses weather data from a position packet's comment field.
// The comment starts with wind direction/speed (CCC/SSS) followed by weather fields.
// Returns any remaining non-weather comment text.
func parseWeatherFromComment(comment string, wx *Weather) string {
	// Weather data starts with wind direction/speed: CCC/SSS
	if len(comment) < 7 {
		return ""
	}

	// Parse wind direction (3 chars) and speed (3 chars) separated by /
	if comment[3] == '/' {
		dirStr := comment[0:3]
		spdStr := comment[4:7]

		// Parse direction - can be dots/spaces for missing
		if dir, n := parseWxInt(dirStr, 3); n > 0 {
			v := float64(dir)
			wx.WindDirection = &v
		}

		if spd, n := parseWxInt(spdStr, 3); n > 0 {
			v := float64(spd) * 0.44704
			wx.WindSpeed = &v
		}

		comment = comment[7:]
	}

	// Now parse the remaining weather fields
	parseWeatherFields(comment, wx)

	// Return the non-weather comment (everything after the software identifier or unrecognized fields)
	return wx.commentAfterWx
}

// parseWeatherFields parses weather data fields from a string.
// Weather data uses single-character field identifiers followed by values.
func parseWeatherFields(data string, wx *Weather) {
	i := 0
	for i < len(data) {
		if i+1 >= len(data) {
			break
		}

		consumed := 0

		switch data[i] {
		case 'c': // Wind direction (degrees)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val)
				wx.WindDirection = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 's': // Wind speed (mph, convert to m/s)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.44704
				wx.WindSpeed = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'g': // Wind gust (mph, convert to m/s)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.44704
				wx.WindGust = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 't': // Temperature (Fahrenheit, convert to Celsius)
			if val, n := parseWxSignedInt(data[i+1:], 3); n > 0 {
				v := (float64(val) - 32.0) * 5.0 / 9.0
				wx.Temp = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'r': // Rain last hour (hundredths of inch, convert to mm)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.254
				wx.Rain1h = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'p': // Rain last 24h (hundredths of inch, convert to mm)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.254
				wx.Rain24h = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'P': // Rain since midnight (hundredths of inch, convert to mm)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.254
				wx.RainMidnight = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'h': // Humidity (%)
			if val, n := parseWxInt(data[i+1:], 2); n > 0 {
				if val == 0 {
					val = 100
				}
				wx.Humidity = &val
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 2); n > 0 {
				consumed = 1 + n
			}
		case 'b': // Barometric pressure (tenths of millibar)
			if val, n := parseWxInt(data[i+1:], 5); n > 0 {
				v := float64(val) / 10.0
				wx.Pressure = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 5); n > 0 {
				consumed = 1 + n
			}
		case 'L', 'l': // Luminosity
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				if data[i] == 'l' {
					val += 1000
				}
				wx.Luminosity = &val
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'O': // Snowfall: s followed by 3 digits (hundredths of inches, convert to mm)
			if i+1 < len(data) && data[i+1] == 's' {
				if val, n := parseWxInt(data[i+2:], 3); n > 0 {
					v := float64(val) * 0.254 // hundredths of inch to mm
					wx.Snow24h = &v
					consumed = 2 + n
				} else if n := skipWxField(data[i+2:], 3); n > 0 {
					consumed = 2 + n
				}
			}
		// Water level, radiation and battery voltage extensions are defined
		// in: https://www.aprs.org/aprs12/weather-new.txt
		case 'F': // Water level (signed, tenths of a foot, convert to meters)
			if val, n := parseWxSignedInt(data[i+1:], 4); n > 0 {
				v := float64(val) / 10.0 * 0.3048
				wx.WaterLevel = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 4); n > 0 {
				consumed = 1 + n
			}
		case 'X': // Ionizing radiation (nanosieverts/hr, resistor code: AB*10^C)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				sig := val / 10 // first two digits
				exp := val % 10 // last digit is exponent
				v := float64(sig) * math.Pow(10, float64(exp))
				wx.Radiation = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case 'V': // Battery voltage (tenths of volts)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) / 10.0
				wx.BatteryVoltage = &v
				consumed = 1 + n
			} else if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		case '#': // Raw rain counter
			if n := skipWxField(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			} else if _, n := parseWxInt(data[i+1:], 3); n > 0 {
				consumed = 1 + n
			}
		}

		if consumed > 0 {
			i += consumed
		} else {
			// Not a recognized weather field - remaining text is comment or software ID
			remaining := strings.TrimSpace(data[i:])
			if isSoftwareID(remaining) {
				wx.Software = remaining
			} else {
				wx.commentAfterWx = remaining
			}
			return
		}
	}
}

// isSoftwareID checks if a string looks like a weather station software identifier
// (3-5 alphanumeric/dash/underscore characters).
func isSoftwareID(s string) bool {
	if len(s) < 3 || len(s) > 5 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// skipWxField returns how many characters to skip for a missing weather field
// (containing only dots or spaces) of maxLen. Returns 0 if the field contains
// other characters (indicating it's not a weather field placeholder).
func skipWxField(s string, maxLen int) int {
	if len(s) < maxLen {
		return 0
	}
	for i := range maxLen {
		if s[i] != '.' && s[i] != ' ' {
			return 0
		}
	}
	return maxLen
}

// parseWxInt extracts an integer of up to maxDigits from the string.
// Returns the value and number of characters consumed.
// Returns 0, 0 if the field contains only dots/spaces (missing data).
func parseWxInt(s string, maxDigits int) (int, int) {
	if len(s) < maxDigits {
		return 0, 0
	}

	field := s[:maxDigits]

	// Check for missing data (dots or spaces)
	allMissing := true
	for _, c := range field {
		if c != '.' && c != ' ' {
			allMissing = false
			break
		}
	}
	if allMissing {
		return 0, 0
	}

	val, err := strconv.Atoi(strings.TrimSpace(field))
	if err != nil {
		return 0, 0
	}

	return val, maxDigits
}

// parseWxSignedInt is like parseWxInt but handles negative values.
func parseWxSignedInt(s string, maxDigits int) (int, int) {
	if len(s) < maxDigits {
		return 0, 0
	}

	field := s[:maxDigits]

	allMissing := true
	for _, c := range field {
		if c != '.' && c != ' ' {
			allMissing = false
			break
		}
	}
	if allMissing {
		return 0, 0
	}

	// Handle negative with leading minus
	trimmed := strings.TrimSpace(field)
	val, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, 0
	}

	return val, maxDigits
}

// parseULTWFields parses sequential 4-hex-character fields from ULTW data.
// Each field is a signed 16-bit integer. "----" means undefined.
func parseULTWFields(s string) []*int {
	var vals []*int
	for len(s) >= 4 {
		field := s[:4]
		s = s[4:]
		if field == "----" {
			vals = append(vals, nil)
			continue
		}
		v64, err := strconv.ParseUint(field, 16, 16)
		if err != nil {
			break
		}
		var v int
		if v64 < 32768 {
			v = int(v64)
		} else {
			v = int(v64) - 65536
		}
		vals = append(vals, &v)
	}
	return vals
}

// ultwShift removes and returns the first element from the slice.
func ultwShift(vals *[]*int) *int {
	if len(*vals) == 0 {
		return nil
	}
	v := (*vals)[0]
	*vals = (*vals)[1:]
	return v
}

// ultwWindSpeed converts a raw ULTW wind value to m/s.
// Formula: val * kmh_to_ms / 10, where kmh_to_ms = 10/36, so val/36.
func ultwWindSpeed(val int) float64 {
	return math.Round(float64(val)/36.0*10) / 10
}

// ultwDirection converts a raw ULTW direction value to degrees.
func ultwDirection(val int) float64 {
	return math.Round(float64(val&0xff) * 1.41176)
}

// ultwTemp converts a raw ULTW temperature (0.1 F) to Celsius.
func ultwTemp(val int) float64 {
	f := float64(val) / 10.0
	c := (f - 32.0) / 1.8
	return math.Round(c*10) / 10
}

// parseULTW parses a $ULTW weather packet.
// Field order: wind_gust, wind_direction, temp, rain_midnight, pressure,
// skip(baro_delta), skip(baro_corr_lsw), skip(baro_corr_msw),
// humidity, skip(date), skip(time), rain_midnight(overwrite), wind_speed
func (p *Packet) parseULTW(opt *options) error {
	p.Type = PacketTypeWx

	body := p.Body[5:] // skip '$ULTW'
	vals := parseULTWFields(body)
	if len(vals) == 0 {
		return p.fail(ErrWxInvalid, "ULTW weather report has no data")
	}

	wx := &Weather{}

	// wind_gust
	if t := ultwShift(&vals); t != nil {
		v := ultwWindSpeed(*t)
		wx.WindGust = &v
	}
	// wind_direction
	if t := ultwShift(&vals); t != nil {
		v := ultwDirection(*t)
		wx.WindDirection = &v
	}
	// temp (outdoor)
	if t := ultwShift(&vals); t != nil {
		v := ultwTemp(*t)
		wx.Temp = &v
	}
	// rain_midnight (may be overwritten later)
	if t := ultwShift(&vals); t != nil {
		v := math.Round(float64(*t)*0.254*10) / 10
		wx.RainMidnight = &v
	}
	// pressure (only if val >= 10)
	t := ultwShift(&vals)
	if t != nil && *t >= 10 {
		v := float64(*t) / 10.0
		wx.Pressure = &v
	}
	// skip: baro delta, baro corr lsw, baro corr msw
	ultwShift(&vals)
	ultwShift(&vals)
	ultwShift(&vals)
	// humidity
	if t := ultwShift(&vals); t != nil {
		h := *t / 10
		if h >= 1 && h <= 100 {
			wx.Humidity = &h
		}
	}
	// skip: date, time
	ultwShift(&vals)
	ultwShift(&vals)
	// rain_midnight (overwrite)
	if t := ultwShift(&vals); t != nil {
		v := math.Round(float64(*t)*0.254*10) / 10
		wx.RainMidnight = &v
	}
	// wind_speed
	if t := ultwShift(&vals); t != nil {
		v := ultwWindSpeed(*t)
		wx.WindSpeed = &v
	}

	p.Wx = wx
	return nil
}

// parseULTWLogging parses a !! ULTW logging format weather packet.
// Field order: wind_speed(instant), wind_direction, temp, rain_midnight,
// pressure, temp_in, humidity, humidity_in, skip(date), skip(time),
// rain_midnight(overwrite), wind_speed(avg, overwrites instant)
func (p *Packet) parseULTWLogging(opt *options) error {
	p.Type = PacketTypeWx

	body := p.Body[2:] // skip '!!'
	vals := parseULTWFields(body)
	if len(vals) == 0 {
		return p.fail(ErrWxInvalid, "ULTW logging weather report has no data")
	}

	wx := &Weather{}

	// wind_speed (instant, may be overwritten by avg later)
	if t := ultwShift(&vals); t != nil {
		v := ultwWindSpeed(*t)
		wx.WindSpeed = &v
	}
	// wind_direction
	if t := ultwShift(&vals); t != nil {
		v := ultwDirection(*t)
		wx.WindDirection = &v
	}
	// temp (outdoor)
	if t := ultwShift(&vals); t != nil {
		v := ultwTemp(*t)
		wx.Temp = &v
	}
	// rain_midnight (may be overwritten later)
	if t := ultwShift(&vals); t != nil {
		v := math.Round(float64(*t)*0.254*10) / 10
		wx.RainMidnight = &v
	}
	// pressure (only if val >= 10)
	t := ultwShift(&vals)
	if t != nil && *t >= 10 {
		v := float64(*t) / 10.0
		wx.Pressure = &v
	}
	// temp_in
	if t := ultwShift(&vals); t != nil {
		v := ultwTemp(*t)
		wx.TempIn = &v
	}
	// humidity
	if t := ultwShift(&vals); t != nil {
		h := *t / 10
		if h >= 1 && h <= 100 {
			wx.Humidity = &h
		}
	}
	// humidity_in
	if t := ultwShift(&vals); t != nil {
		h := *t / 10
		if h >= 1 && h <= 100 {
			wx.HumidityIn = &h
		}
	}
	// skip: date, time
	ultwShift(&vals)
	ultwShift(&vals)
	// rain_midnight (overwrite)
	if t := ultwShift(&vals); t != nil {
		v := math.Round(float64(*t)*0.254*10) / 10
		wx.RainMidnight = &v
	}
	// wind_speed (avg, overwrites instant)
	if t := ultwShift(&vals); t != nil {
		v := ultwWindSpeed(*t)
		wx.WindSpeed = &v
	}

	// if inside temperature exists but no outside, use inside
	if wx.TempIn != nil && wx.Temp == nil {
		wx.Temp = wx.TempIn
	}

	p.Wx = wx
	return nil
}
