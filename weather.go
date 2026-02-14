package fap

import (
	"strconv"
	"strings"
)

// parseWeatherPositionless parses a positionless weather report.
// Format: _MMDDHHMM weather data...
func (p *Packet) parseWeatherPositionless(opt Options) error {
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

// parseWeatherFields parses weather data fields from a string.
// Weather data uses single-character field identifiers followed by values.
func parseWeatherFields(data string, wx *Weather) {
	i := 0
	for i < len(data) {
		if i+1 >= len(data) {
			break
		}

		switch data[i] {
		case 'c': // Wind direction (degrees)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val)
				wx.WindDirection = &v
				i += 1 + n
			} else {
				i++
			}
		case 's': // Wind speed (mph, convert to m/s)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.44704
				wx.WindSpeed = &v
				i += 1 + n
			} else {
				i++
			}
		case 'g': // Wind gust (mph, convert to m/s)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.44704
				wx.WindGust = &v
				i += 1 + n
			} else {
				i++
			}
		case 't': // Temperature (Fahrenheit, convert to Celsius)
			if val, n := parseWxSignedInt(data[i+1:], 3); n > 0 {
				v := (float64(val) - 32.0) * 5.0 / 9.0
				wx.Temp = &v
				i += 1 + n
			} else {
				i++
			}
		case 'r': // Rain last hour (hundredths of inch, convert to mm)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.254
				wx.Rain1h = &v
				i += 1 + n
			} else {
				i++
			}
		case 'p': // Rain last 24h (hundredths of inch, convert to mm)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.254
				wx.Rain24h = &v
				i += 1 + n
			} else {
				i++
			}
		case 'P': // Rain since midnight (hundredths of inch, convert to mm)
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 0.254
				wx.RainMidnight = &v
				i += 1 + n
			} else {
				i++
			}
		case 'h': // Humidity (%)
			if val, n := parseWxInt(data[i+1:], 2); n > 0 {
				if val == 0 {
					val = 100
				}
				wx.Humidity = &val
				i += 1 + n
			} else {
				i++
			}
		case 'b': // Barometric pressure (tenths of millibar)
			if val, n := parseWxInt(data[i+1:], 5); n > 0 {
				v := float64(val) / 10.0
				wx.Pressure = &v
				i += 1 + n
			} else {
				i++
			}
		case 'L', 'l': // Luminosity
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				if data[i] == 'l' {
					val += 1000
				}
				wx.Luminosity = &val
				i += 1 + n
			} else {
				i++
			}
		case '#': // Snowfall (inches, convert to mm) - raw rain counter in some implementations
			if val, n := parseWxInt(data[i+1:], 3); n > 0 {
				v := float64(val) * 25.4
				wx.Snow24h = &v
				i += 1 + n
			} else {
				i++
			}
		default:
			i++
		}
	}

	// Look for software identifier at the end
	if idx := strings.LastIndexByte(data, '{'); idx >= 0 {
		wx.Soft = data[idx+1:]
	}
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
