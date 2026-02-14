// Package fap implements an APRS (Automatic Packet Reporting System) packet
// parser in pure Go. It parses APRS packets in the TNC2 / APRS-IS text format.
//
// This is a Go port of the Ham::APRS::FAP Perl module, originally written
// for the aprs.fi service.
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
	"fmt"
	"math"
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
	PacketTypeDX           PacketType = "dx"
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
	Call       string // Callsign of the digipeater
	WasDigied bool   // Whether this digipeater has already relayed the packet
}

// Weather contains weather data from a weather report packet.
type Weather struct {
	WindDirection *float64 // Wind direction in degrees
	WindSpeed     *float64 // Wind speed in m/s
	WindGust      *float64 // Wind gust speed in m/s
	Temp          *float64 // Temperature in degrees Celsius
	TempIn        *float64 // Indoor temperature in degrees Celsius
	Humidity      *int     // Relative humidity in percent
	HumidityIn    *int     // Indoor humidity in percent
	Pressure      *float64 // Barometric pressure in millibars
	Rain1h        *float64 // Rain in the last hour in mm
	Rain24h       *float64 // Rain in the last 24 hours in mm
	RainMidnight  *float64 // Rain since midnight in mm
	Snow24h       *float64 // Snow in the last 24 hours in mm
	Luminosity    *int     // Luminosity in watts per square meter
	Soft          string   // Software / device identifier
}

// Telemetry contains telemetry data.
type Telemetry struct {
	Seq  string    // Sequence number
	Vals []float64 // Analog values
	Bits string    // Digital bits (8-bit string)
}

// Packet represents a parsed APRS packet.
type Packet struct {
	// Always present on successful parse
	OrigPacket  string        // Original packet string
	Header      string        // Raw packet header (before the first colon)
	Body        string        // Raw packet body (after the first colon)
	SrcCallsign string        // Source callsign
	DstCallsign string        // Destination callsign
	Digipeaters []Digipeater  // Digipeater path

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
	Timestamp *time.Time // Timestamp from the packet

	// Objects and items
	ObjectName string // Name of object
	ItemName   string // Name of item
	Alive      *bool  // Object/item alive status

	// Messages
	Destination string // Message destination callsign
	Message     string // Message text
	MessageID   string // Message ID
	MessageAck  string // Message acknowledgment ID
	MessageRej  string // Message reject ID

	// Status
	Status string // Status text

	// Weather
	Wx *Weather // Weather data (nil if no weather)

	// Telemetry
	TelemetryData *Telemetry // Telemetry data

	// Capabilities
	Capabilities map[string]string // Station capabilities

	// Mic-E specifics
	MBits        string // Mic-E message bits
	MiceMangled  bool   // True if mic-e packet was repaired

	// DAO
	DaoDatumByte byte // DAO datum byte

	// GPS fix
	GPSFixStatus *int // GPS fix status (0 or 1)

	// Comment
	Comment string // Packet comment text

	// Error/warning info (populated on parse failure)
	ResultCode string // Machine-readable error code
	ResultMsg  string // Human-readable error message
}

// Options configures parsing behavior.
type Options struct {
	// IsAX25 validates the packet against AX.25 rules.
	IsAX25 bool

	// AcceptBrokenMicE attempts to fix corrupted mic-e packets.
	AcceptBrokenMicE bool

	// RawTimestamp returns timestamps as raw strings instead of time.Time.
	RawTimestamp bool
}

// ParseAPRS parses an APRS packet in TNC2 / APRS-IS text format.
// It returns a Packet struct with all parsed fields populated.
// If parsing fails, the returned Packet will have ResultCode and ResultMsg set.
func ParseAPRS(raw string, opts ...Options) (*Packet, error) {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	p := &Packet{
		OrigPacket: raw,
	}

	// Split header and body at the first colon
	colonIdx := strings.IndexByte(raw, ':')
	if colonIdx < 0 {
		return p, p.fail("packet_no_body", "no packet body after header")
	}

	p.Header = raw[:colonIdx]
	p.Body = raw[colonIdx+1:]

	if len(p.Body) == 0 {
		return p, p.fail("packet_no_body", "packet body is empty")
	}

	// Parse header: SRC>DST,DIGI1,DIGI2,...
	if err := p.parseHeader(opt); err != nil {
		return p, err
	}

	// Determine packet type from the first character(s) of the body
	if err := p.parseBody(opt); err != nil {
		return p, err
	}

	return p, nil
}

// fail sets error fields on the packet and returns an error.
func (p *Packet) fail(code, msg string) error {
	p.ResultCode = code
	p.ResultMsg = msg
	return fmt.Errorf("fap: %s: %s", code, msg)
}

// parseHeader parses the packet header into source, destination, and digipeaters.
func (p *Packet) parseHeader(opt Options) error {
	// Split at '>'
	gtIdx := strings.IndexByte(p.Header, '>')
	if gtIdx < 0 {
		return p.fail("srccall_nogt", "no '>' in header")
	}

	p.SrcCallsign = p.Header[:gtIdx]
	if len(p.SrcCallsign) == 0 {
		return p.fail("srccall_empty", "source callsign is empty")
	}

	rest := p.Header[gtIdx+1:]
	if len(rest) == 0 {
		return p.fail("dstcall_empty", "destination callsign is empty")
	}

	// Split the rest by commas: first is destination, rest are digipeaters
	parts := strings.Split(rest, ",")
	p.DstCallsign = parts[0]

	if len(p.DstCallsign) == 0 {
		return p.fail("dstcall_empty", "destination callsign is empty")
	}

	// Parse digipeaters
	for _, d := range parts[1:] {
		digi := Digipeater{}
		if strings.HasSuffix(d, "*") {
			digi.WasDigied = true
			digi.Call = d[:len(d)-1]
		} else {
			digi.Call = d
		}
		if len(digi.Call) == 0 {
			return p.fail("digi_empty", "empty digipeater callsign")
		}
		p.Digipeaters = append(p.Digipeaters, digi)
	}

	return nil
}

// parseBody dispatches body parsing based on the packet type identifier.
func (p *Packet) parseBody(opt Options) error {
	if len(p.Body) == 0 {
		return p.fail("packet_no_body", "packet body is empty")
	}

	typeChar := p.Body[0]

	switch typeChar {
	case '!', '=':
		// Position without timestamp
		return p.parsePositionNoTimestamp(opt, typeChar)
	case '/', '@':
		// Position with timestamp
		return p.parsePositionWithTimestamp(opt, typeChar)
	case '\'', '`':
		// Mic-E
		return p.parseMicE(opt)
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
		// NMEA
		return p.parseNMEA(opt)
	case 'T':
		// Telemetry
		if len(p.Body) > 1 && p.Body[1] == '#' {
			return p.parseTelemetry(opt)
		}
		return p.parsePositionFallback(opt)
	default:
		// Try last-resort position parsing (look for ! in body)
		return p.parsePositionFallback(opt)
	}
}

// Helper functions

// Distance calculates the great-circle distance in kilometers between two
// points specified in decimal degrees.
func Distance(lat0, lon0, lat1, lon1 float64) float64 {
	lat0r := lat0 * math.Pi / 180.0
	lon0r := lon0 * math.Pi / 180.0
	lat1r := lat1 * math.Pi / 180.0
	lon1r := lon1 * math.Pi / 180.0

	dlon := lon1r - lon0r
	dlat := lat1r - lat0r

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat0r)*math.Cos(lat1r)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6366.71 * c // Earth radius in km
}

// Direction calculates the initial bearing in degrees from point 0 to point 1,
// with both points specified in decimal degrees.
func Direction(lat0, lon0, lat1, lon1 float64) float64 {
	lat0r := lat0 * math.Pi / 180.0
	lon0r := lon0 * math.Pi / 180.0
	lat1r := lat1 * math.Pi / 180.0
	lon1r := lon1 * math.Pi / 180.0

	dlon := lon1r - lon0r

	direction := math.Atan2(
		math.Sin(dlon)*math.Cos(lat1r),
		math.Cos(lat0r)*math.Sin(lat1r)-math.Sin(lat0r)*math.Cos(lat1r)*math.Cos(dlon),
	) * 180.0 / math.Pi

	if direction < 0 {
		direction += 360.0
	}

	return direction
}
