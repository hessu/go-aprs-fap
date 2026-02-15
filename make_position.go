package fap

import (
	"fmt"
	"strconv"
)

// MakePositionOpts contains optional parameters for MakePosition.
type MakePositionOpts struct {
	Ambiguity int    // 0-4
	DAO       bool   // enable !DAO! extension for extra precision
	Comment   string // comment to append
}

// MakePosition creates an uncompressed APRS position string.
// lat/lon are in decimal degrees, speed in km/h, course in degrees, altitude in meters.
// symbol is a 2-character string (table + code).
func MakePosition(lat, lon float64, speed, course, altitude *float64, symbol string, opts *MakePositionOpts) (string, error) {
	if opts == nil {
		opts = &MakePositionOpts{}
	}

	// If ambiguity is set, DAO is not applicable
	if opts.Ambiguity > 0 {
		opts.DAO = false
	}

	// Validate coordinates
	if lat < -89.99999 || lat > 89.99999 || lon < -179.99999 || lon > 179.99999 {
		return "", fmt.Errorf("invalid coordinates: lat=%f lon=%f", lat, lon)
	}

	// Parse symbol
	var symbolTable, symbolCode byte
	if len(symbol) == 0 {
		symbolTable = '/'
		symbolCode = '/'
	} else if len(symbol) == 2 {
		symbolTable = symbol[0]
		symbolCode = symbol[1]
		if !isValidSymbolTable(symbolTable) {
			return "", fmt.Errorf("invalid symbol table: %c", symbolTable)
		}
		if symbolCode < 0x21 || symbolCode > 0x7b && symbolCode != 0x7d {
			return "", fmt.Errorf("invalid symbol code: %c", symbolCode)
		}
	} else {
		return "", fmt.Errorf("invalid symbol length: %d", len(symbol))
	}

	// Convert latitude to degrees and minutes
	isNorth := true
	if lat < 0 {
		lat = -lat
		isNorth = false
	}
	latDeg := int(lat)
	latMin := (lat - float64(latDeg)) * 60

	var latMinS string
	var latMinDAO string
	if opts.DAO {
		latMinS = fmt.Sprintf("%06.0f", latMin*10000)
		if len(latMinS) > 4 {
			latMinDAO = latMinS[4:6]
		}
	} else {
		latMinS = fmt.Sprintf("%04.0f", latMin*100)
	}

	// Check for rounding to 60 minutes
	if len(latMinS) >= 2 && latMinS[0] == '6' && latMinS[1] == '0' {
		latMinS = "5999"
		latMinDAO = "99"
	}

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

	var lonMinS string
	var lonMinDAO string
	if opts.DAO {
		lonMinS = fmt.Sprintf("%06.0f", lonMin*10000)
		if len(lonMinS) > 4 {
			lonMinDAO = lonMinS[4:6]
		}
	} else {
		lonMinS = fmt.Sprintf("%04.0f", lonMin*100)
	}

	// Check for rounding to 60 minutes
	if len(lonMinS) >= 2 && lonMinS[0] == '6' && lonMinS[1] == '0' {
		lonMinS = "5999"
		lonMinDAO = "99"
	}

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

	// Build result
	result := "!" + latString + string(symbolTable) + lonString + string(symbolCode)

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
