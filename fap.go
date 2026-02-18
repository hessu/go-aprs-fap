// Package fap implements an APRS (Automatic Packet Reporting System) packet
// parser in pure Go. It parses APRS packets in the TNC2 / APRS-IS text format.
//
// This is a Go port of the Ham::APRS::FAP Perl module, which is used by
// the aprs.fi service.
//
// Supported packet types:
//   - Position (uncompressed, compressed, Mic-E)
//   - Objects and items
//   - Messages, acks, and rejects
//   - Weather reports
//   - Telemetry
//   - Status reports
//   - Station capabilities
//   - NMEA (GPRMC, GPGGA, GPGLL)
//   - DX spots
package fap

import (
	"strings"
	"time"
)

// PacketType represents the type of an APRS packet.
type PacketType string

const (
	PacketTypeLocation     PacketType = "location"
	PacketTypeObject       PacketType = "object"
	PacketTypeItem         PacketType = "item"
	PacketTypeMessage      PacketType = "message"
	PacketTypeWx           PacketType = "wx"
	PacketTypeTelemetry    PacketType = "telemetry"
	PacketTypeStatus       PacketType = "status"
	PacketTypeCapabilities PacketType = "capabilities"
)

// Format represents the position encoding format used in a packet.
type Format string

const (
	FormatUncompressed Format = "uncompressed"
	FormatCompressed   Format = "compressed"
	FormatMicE         Format = "mice"
	FormatNMEA         Format = "nmea"
)

// Digipeater represents a digipeater in the packet path.
type Digipeater struct {
	Call      string // Callsign of the digipeater
	WasDigied bool   // Whether this digipeater has already relayed the packet
}

// Weather contains weather data from a weather report packet.
type Weather struct {
	WindDirection  *float64 // Wind direction in degrees
	WindSpeed      *float64 // Wind speed in m/s
	WindGust       *float64 // Wind gust speed in m/s
	Temp           *float64 // Temperature in degrees Celsius
	TempIn         *float64 // Indoor temperature in degrees Celsius
	Humidity       *int     // Relative humidity in percent
	HumidityIn     *int     // Indoor humidity in percent
	Pressure       *float64 // Barometric pressure in millibars
	Rain1h         *float64 // Rain in the last hour in mm
	Rain24h        *float64 // Rain in the last 24 hours in mm
	RainMidnight   *float64 // Rain since midnight in mm
	Snow24h        *float64 // Snow in the last 24 hours in mm
	Luminosity     *int     // Luminosity in watts per square meter
	WaterLevel     *float64 // Water level above or below flood stage in meters
	Radiation      *float64 // Ionizing radiation in nSv/hour
	BatteryVoltage *float64 // Battery voltage, in V
	Software       string   // Software / device identifier
	commentAfterWx string   // internal: non-weather comment text after weather data
}

// Telemetry contains telemetry data.
type Telemetry struct {
	Seq  int        // Sequence number
	Vals []*float64 // Analog values (nil = undefined)
	Bits string     // Digital bits (8-bit string)
}

// Message contains data from an APRS message packet.
type Message struct {
	Destination string // Message destination callsign
	Text        string // Message text
	ID          string // Message ID
	AckID       string // Message acknowledgment ID
	RejID       string // Message reject ID
}

// Packet represents a parsed APRS packet.
type Packet struct {
	// Always present on successful parse
	OrigPacket  string       // Original packet string
	Header      string       // Raw packet header (before the first colon)
	Body        string       // Raw packet body (after the first colon)
	SrcCallsign string       // Source callsign
	DstCallsign string       // Destination callsign
	Digipeaters []Digipeater // Digipeater path

	// Packet type and format
	Type   PacketType // Packet type
	Format Format     // Position encoding format (for location packets)

	// Position data
	Latitude      *float64 // Latitude in decimal degrees (negative for south)
	Longitude     *float64 // Longitude in decimal degrees (negative for west)
	PosAmbiguity  *int     // Position ambiguity level (0-4)
	PosResolution *float64 // Position resolution in meters

	// Symbol
	SymbolTable byte // Symbol table identifier ('/', '\', or overlay)
	SymbolCode  byte // Symbol code character

	// Movement
	Speed    *float64 // Speed in km/h
	Course   *int     // Course/heading in degrees (0-360)
	Altitude *float64 // Altitude in meters

	// Flags
	Messaging *bool // Messaging capability (nil if unknown)

	// PHG and radio range
	PHG        string   // PHG data string (4 digits)
	RadioRange *float64 // Radio range in km

	// Timestamp
	Timestamp    *time.Time // Timestamp from the packet (when RawTimestamp is false)
	RawTimestamp string     // Raw timestamp string (when RawTimestamp option is true)

	// Objects and items
	ObjectName string // Name of object
	ItemName   string // Name of item
	Alive      *bool  // Object/item alive status

	// Messages
	Message *Message // Message data (nil if not a message packet)

	// Status
	Status string // Status text

	// Weather
	Wx *Weather // Weather data (nil if no weather)

	// Telemetry
	TelemetryData *Telemetry // Telemetry data

	// Capabilities
	Capabilities map[string]string // Station capabilities

	// Mic-E specifics
	MBits       string // Mic-E message bits
	MiceMangled bool   // True if mic-e packet was repaired

	// DAO
	DaoDatumByte byte // DAO datum byte

	// GPS fix
	GPSFixStatus *int // GPS fix status (0 or 1)

	// NMEA
	ChecksumOK *bool // NMEA checksum validation result

	// Comment
	Comment string // Packet comment text

	// Warnings collected during parsing (non-fatal issues)
	Warnings []ParseError
}

// options holds internal parsing configuration.
type options struct {
	isAX25           bool
	acceptBrokenMicE bool
	rawTimestamp     bool
}

// Option configures parsing behavior.
type Option func(*options)

// WithAX25 validates the packet against AX.25 rules.
func WithAX25() Option {
	return func(o *options) { o.isAX25 = true }
}

// WithAcceptBrokenMicE attempts to fix corrupted mic-e packets.
func WithAcceptBrokenMicE() Option {
	return func(o *options) { o.acceptBrokenMicE = true }
}

// WithRawTimestamp returns timestamps as raw strings instead of time.Time.
func WithRawTimestamp() Option {
	return func(o *options) { o.rawTimestamp = true }
}

// Parse parses an APRS packet in TNC2 / APRS-IS text format.
// It returns a Packet struct with all parsed fields populated.
// On failure, the returned error is a *ParseError with Code and Msg fields.
func Parse(raw string, opts ...Option) (*Packet, error) {
	var opt options
	for _, o := range opts {
		o(&opt)
	}

	p := &Packet{
		OrigPacket: raw,
	}

	// Split header and body at the first colon
	colonIdx := strings.IndexByte(raw, ':')
	if colonIdx < 0 {
		return p, p.fail(ErrPacketNoBody, "no packet body after header")
	}

	p.Header = raw[:colonIdx]
	p.Body = raw[colonIdx+1:]

	// Parse header: SRC>DST,DIGI1,DIGI2,...
	if err := p.parseHeader(&opt); err != nil {
		return p, err
	}

	// Determine packet type from the first character(s) of the body
	if err := p.parseBody(&opt); err != nil {
		return p, err
	}

	return p, nil
}

// fail returns a *ParseError with the given sentinel's code and the specific message.
func (p *Packet) fail(code *ParseError, msg string) error {
	return &ParseError{Code: code.Code, Msg: msg}
}

// warn records a non-fatal parsing warning on the packet.
func (p *Packet) warn(code *ParseError, msg string) {
	p.Warnings = append(p.Warnings, ParseError{Code: code.Code, Msg: msg})
}

// parseHeader parses the packet header into source, destination, and digipeaters.
func (p *Packet) parseHeader(opt *options) error {
	// Split at '>'
	gtIdx := strings.IndexByte(p.Header, '>')
	if gtIdx < 0 {
		return p.fail(ErrSrcCallNoGT, "no '>' in header")
	}

	p.SrcCallsign = p.Header[:gtIdx]
	if len(p.SrcCallsign) == 0 {
		return p.fail(ErrSrcCallEmpty, "source callsign is empty")
	}

	// Validate source callsign
	if opt.isAX25 {
		normalized := CheckAX25Call(p.SrcCallsign)
		if normalized == "" {
			return p.fail(ErrSrcCallNoAX25, "source callsign is not a valid AX.25 call")
		}
		p.SrcCallsign = normalized
	} else if !isValidSrcCall(p.SrcCallsign) {
		return p.fail(ErrSrcCallBadChars, "source callsign contains bad characters")
	}

	rest := p.Header[gtIdx+1:]
	if len(rest) == 0 {
		return p.fail(ErrDstCallEmpty, "destination callsign is empty")
	}

	// Split the rest by commas: first is destination, rest are digipeaters
	parts := strings.Split(rest, ",")

	// AX.25 limits path to 9 components (1 dst + 8 digipeaters)
	if opt.isAX25 && len(parts) > 9 {
		return p.fail(ErrDstPathTooMany, "too many path components for AX.25")
	}

	// Destination callsign is always validated as AX.25 — there should be
	// no need to use a non-AX.25 compatible destination callsign.
	p.DstCallsign = CheckAX25Call(parts[0])
	if p.DstCallsign == "" {
		return p.fail(ErrDstCallNoAX25, "destination callsign is not a valid AX.25 call")
	}

	// Parse digipeaters
	seenQConstr := false
	for _, d := range parts[1:] {
		digi := Digipeater{}
		if strings.HasSuffix(d, "*") {
			digi.WasDigied = true
			digi.Call = d[:len(d)-1]
		} else {
			digi.Call = d
		}
		if len(digi.Call) == 0 {
			return p.fail(ErrDigiEmpty, "empty digipeater callsign")
		}

		// Validate digipeater callsign
		if opt.isAX25 {
			normalized := CheckAX25Call(digi.Call)
			if normalized == "" {
				return p.fail(ErrDigiCallNoAX25, "digipeater callsign is not a valid AX.25 call")
			}
			digi.Call = normalized
		} else {
			if isValidDigiCall(digi.Call) {
				if isQConstruct(digi.Call) {
					seenQConstr = true
				}
			} else if seenQConstr && isIPv6Hex(digi.Call) {
				// Allow 32-char hex IPv6 addresses after q-construct
			} else {
				return p.fail(ErrDigiCallBadChars, "digipeater callsign contains bad characters")
			}
		}

		p.Digipeaters = append(p.Digipeaters, digi)
	}

	return nil
}

// parseBody dispatches body parsing based on the packet type identifier.
func (p *Packet) parseBody(opt *options) error {
	if len(p.Body) == 0 {
		return p.fail(ErrPacketNoBody, "packet body is empty")
	}

	typeChar := p.Body[0]

	switch typeChar {
	case '!':
		// Position without timestamp, or !! ULTW weather
		if len(p.Body) > 1 && p.Body[1] == '!' {
			return p.parseULTWLogging(opt)
		}
		return p.parsePositionNoTimestamp(opt, typeChar)
	case '=':
		// Position without timestamp (with messaging)
		return p.parsePositionNoTimestamp(opt, typeChar)
	case '/', '@':
		// Position with timestamp
		return p.parsePositionWithTimestamp(opt, typeChar)
	case '\'', '`':
		// Mic-E
		err := p.parseMicE(opt)
		if err != nil && opt.acceptBrokenMicE {
			// Reset fields that parseMicE may have partially set
			p.Speed = nil
			p.Course = nil
			p.Altitude = nil
			p.Comment = ""
			return p.parseMicEMangled(opt)
		}
		return err
	case ':':
		// Message
		return p.parseMessage(opt)
	case ';':
		// Object
		return p.parseObject(opt)
	case ')':
		// Item
		return p.parseItem(opt)
	case '>':
		// Status
		return p.parseStatus(opt)
	case '<':
		// Capabilities
		return p.parseCapabilities(opt)
	case '_':
		// Positionless weather
		return p.parseWeatherPositionless(opt)
	case '$':
		// NMEA or $ULTW weather
		if strings.HasPrefix(p.Body, "$ULTW") {
			return p.parseULTW(opt)
		}
		return p.parseNMEA(opt)
	case 'T':
		// Telemetry
		if len(p.Body) > 1 && p.Body[1] == '#' {
			return p.parseTelemetry(opt)
		}
		return p.parsePositionFallback(opt)
	case '{':
		// Experimental
		if len(p.Body) > 1 && p.Body[1] == '{' {
			return p.fail(ErrExpUnsupported, "unsupported experimental packet")
		}
		return p.parsePositionFallback(opt)
	default:
		// Try last-resort position parsing (look for ! in body)
		return p.parsePositionFallback(opt)
	}
}

// CheckAX25Call validates and normalizes an AX.25 callsign.
// Returns the normalized callsign (with SSID if present) or empty string if invalid.
func CheckAX25Call(call string) string {
	s := strings.ToUpper(call)

	// Split base callsign and optional SSID at dash
	base := s
	ssidStr := ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		base = s[:i]
		ssidStr = s[i+1:]
		// If there was a dash with nothing after it, that's invalid.
		if ssidStr == "" {
			return ""
		}
	}

	// Base must be 1-6 alphanumeric characters
	if len(base) < 1 || len(base) > 6 {
		return ""
	}
	for i := 0; i < len(base); i++ {
		c := base[i]
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return ""
		}
	}

	if ssidStr == "" {
		return base
	}

	// SSID must be 1-2 digits, value 0-15
	if len(ssidStr) < 1 || len(ssidStr) > 2 {
		return ""
	}
	ssid := 0
	for i := 0; i < len(ssidStr); i++ {
		c := ssidStr[i]
		if c < '0' || c > '9' {
			return ""
		}
		ssid = ssid*10 + int(c-'0')
	}
	if ssid > 15 {
		return ""
	}

	return base + "-" + ssidStr
}

// isAlnumDash reports whether s is 1–9 characters of [A-Za-z0-9-].
func isAlnumDash(s string, maxLen int) bool {
	if len(s) < 1 || len(s) > maxLen {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// isValidSrcCall reports whether s is a valid APRS-IS source callsign.
func isValidSrcCall(s string) bool {
	return isAlnumDash(s, 9)
}

// isValidDigiCall reports whether s is a valid APRS-IS digipeater callsign.
func isValidDigiCall(s string) bool {
	return isAlnumDash(s, 9)
}

// isQConstruct reports whether s is a q-construct (e.g. "qAR", "qAo").
func isQConstruct(s string) bool {
	return len(s) == 3 && s[0] == 'q'
}

// isIPv6Hex reports whether s is a 32-character uppercase hex string.
func isIPv6Hex(s string) bool {
	if len(s) != 32 {
		return false
	}
	for i := range 32 {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
