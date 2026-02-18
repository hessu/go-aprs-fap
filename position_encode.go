package fap

import (
	"fmt"
	"strconv"
	"time"
)

// EncodePositionOpts contains optional parameters for EncodePosition.
type EncodePositionOpts struct {
	Ambiguity        int       // 0-4
	Timestamp        time.Time // if non-zero, include HHMMSSh UTC timestamp
	MessagingCapable bool      // report that the station can receive text messages
	DAO              bool      // enable !DAO! extension for extra precision
	Comment          string    // comment to append
}

// formatMinutes converts fractional minutes to a string for APRS position encoding.
// With dao=true, returns 4-digit minutes string and 2-digit DAO extension.
// With dao=false, returns 4-digit minutes string and empty DAO.
// Handles rounding to 60 minutes by clamping to 5999.
func formatMinutes(min float64, dao bool) (minS, daoS string) {
	if dao {
		minS = fmt.Sprintf("%06.0f", min*10000)
		if len(minS) > 4 {
			daoS = minS[4:6]
		}
	} else {
		minS = fmt.Sprintf("%04.0f", min*100)
	}
	if len(minS) >= 2 && minS[0] == '6' && minS[1] == '0' {
		minS = "5999"
		daoS = "99"
	}
	return minS, daoS
}

// EncodePosition creates an uncompressed APRS position string.
// lat/lon are in decimal degrees, speed in km/h, course in degrees, altitude in meters.
// symbol is a 2-character string (table + code).
func EncodePosition(lat, lon float64, speed, course, altitude *float64, symbol string, opts *EncodePositionOpts) (string, error) {
	if opts == nil {
		opts = &EncodePositionOpts{}
	}

	// If ambiguity is set, DAO is not applicable
	if opts.Ambiguity > 0 {
		opts.DAO = false
	}

	// Validate coordinates
	if lat < -89.99999 || lat > 89.99999 || lon < -179.99999 || lon > 179.99999 {
		return "", &ParseError{Code: ErrPosEncInvalid.Code, Msg: fmt.Sprintf("invalid coordinates: lat=%f lon=%f", lat, lon)}
	}

	// Parse symbol
	var symbolTable, symbolCode byte
	if len(symbol) == 2 {
		symbolTable = symbol[0]
		symbolCode = symbol[1]
		if !isValidSymbolTable(symbolTable) {
			return "", &ParseError{Code: ErrPosEncInvalid.Code, Msg: fmt.Sprintf("invalid symbol table: %c", symbolTable)}
		}
		if symbolCode < 0x21 || symbolCode > 0x7b && symbolCode != 0x7d {
			return "", &ParseError{Code: ErrPosEncInvalid.Code, Msg: fmt.Sprintf("invalid symbol code: %c", symbolCode)}
		}
	} else {
		return "", &ParseError{Code: ErrPosEncInvalid.Code, Msg: fmt.Sprintf("invalid symbol length: %d", len(symbol))}
	}

	// Convert latitude to degrees and minutes
	isNorth := true
	if lat < 0 {
		lat = -lat
		isNorth = false
	}
	latDeg := int(lat)
	latMin := (lat - float64(latDeg)) * 60

	latMinS, latMinDAO := formatMinutes(latMin, opts.DAO)

	latString := fmt.Sprintf("%02d%s.%s", latDeg, latMinS[0:2], latMinS[2:4])

	// Apply position ambiguity
	if opts.Ambiguity > 0 && opts.Ambiguity <= 4 {
		latBytes := []byte(latString)
		if opts.Ambiguity <= 2 {
			for i := 0; i < opts.Ambiguity; i++ {
				latBytes[7-1-i] = ' '
			}
		} else if opts.Ambiguity == 3 {
			latString = latString[:3] + " .  "
		} else if opts.Ambiguity == 4 {
			latString = latString[:2] + "  .  "
		}
		if opts.Ambiguity <= 2 {
			latString = string(latBytes)
		}
	}

	if isNorth {
		latString += "N"
	} else {
		latString += "S"
	}

	// Convert longitude to degrees and minutes
	isEast := true
	if lon < 0 {
		lon = -lon
		isEast = false
	}
	lonDeg := int(lon)
	lonMin := (lon - float64(lonDeg)) * 60

	lonMinS, lonMinDAO := formatMinutes(lonMin, opts.DAO)

	lonString := fmt.Sprintf("%03d%s.%s", lonDeg, lonMinS[0:2], lonMinS[2:4])

	// Apply position ambiguity
	if opts.Ambiguity > 0 && opts.Ambiguity <= 4 {
		lonBytes := []byte(lonString)
		if opts.Ambiguity <= 2 {
			for i := 0; i < opts.Ambiguity; i++ {
				lonBytes[8-1-i] = ' '
			}
		} else if opts.Ambiguity == 3 {
			lonString = lonString[:4] + " .  "
		} else if opts.Ambiguity == 4 {
			lonString = lonString[:3] + "  .  "
		}
		if opts.Ambiguity <= 2 {
			lonString = string(lonBytes)
		}
	}

	if isEast {
		lonString += "E"
	} else {
		lonString += "W"
	}

	// Build result - data type identifier depends on timestamp and messaging capability:
	//   ! = no timestamp, no messaging
	//   = = no timestamp, messaging capable
	//   / = timestamp, no messaging
	//   @ = timestamp, messaging capable
	var result string
	if !opts.Timestamp.IsZero() {
		utc := opts.Timestamp.UTC()
		dtid := byte('/')
		if opts.MessagingCapable {
			dtid = '@'
		}
		result = fmt.Sprintf("%c%02d%02d%02dh", dtid, utc.Hour(), utc.Minute(), utc.Second())
	} else {
		if opts.MessagingCapable {
			result = "="
		} else {
			result = "!"
		}
	}
	result += latString + string(symbolTable) + lonString + string(symbolCode)

	// Add course/speed if both provided
	if speed != nil && course != nil && *speed >= 0 && *course >= 0 {
		speedKnots := *speed / 1.852
		if speedKnots > 999 {
			speedKnots = 999
		}
		c := *course
		if c > 360 {
			c = 0
		}
		result += fmt.Sprintf("%03.0f/%03.0f", c, speedKnots)
	}

	// Add altitude if provided
	if altitude != nil {
		altFeet := *altitude / 0.3048
		if altFeet >= 0 {
			result += fmt.Sprintf("/A=%06.0f", altFeet)
		} else {
			result += fmt.Sprintf("/A=-%05.0f", -altFeet)
		}
	}

	// Add comment
	if opts.Comment != "" {
		result += opts.Comment
	}

	// Add DAO extension
	if opts.DAO && latMinDAO != "" && lonMinDAO != "" {
		latDAO, _ := strconv.Atoi(latMinDAO)
		lonDAO, _ := strconv.Atoi(lonMinDAO)
		latChar := byte(int(float64(latDAO)/1.1+0.5) + 33)
		lonChar := byte(int(float64(lonDAO)/1.1+0.5) + 33)
		result += "!w" + string(latChar) + string(lonChar) + "!"
	}

	return result, nil
}
