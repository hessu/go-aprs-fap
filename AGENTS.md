# AGENTS.md

This file provides guidance to agentic coding agents such as Claude Code
(claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
go build ./...                              # Build
go test -v ./...                            # Run all tests
go test -run TestParseUncompressedNortheast -v ./...  # Run a single test
go test -cover ./...                        # Tests with coverage
```

## Architecture

This is a Go port of the Perl `Ham::APRS::FAP` module — an APRS (Automatic
Packet Reporting System) packet parser.  Package name is `fap`, module path
`github.com/hessu/go-aprs-fap`.  No external dependencies (stdlib only).

The original perl module is available at ../perl-aprs-fap
for inspection at any time.

### Entry Point

```go
func ParseAPRS(raw string, opts ...Options) (*Packet, error)
```

Parses a TNC2/APRS-IS format packet string into a `Packet` struct. Options: `IsAX25`, `AcceptBrokenMicE`, `RawTimestamp`.

### Parsing Flow

1. **Header parsing** (fap.go): Split at first `:`, extract `SOURCE>DEST,DIGI1,DIGI2,...` and body
2. **Body dispatch** (fap.go): First character determines packet type:
   - `!` `=` → uncompressed/compressed position (position.go)
   - `/` `@` → position with timestamp (position.go + timestamp.go)
   - `` ` `` `'` → Mic-E encoded position (mice.go)
   - `:` → message/ack/reject (message.go)
   - `;` → object, `)` → item (object.go)
   - `>` → status, `<` → capabilities (status.go)
   - `_` → positionless weather (weather.go)
   - `$` → NMEA (nmea.go) or ULTW weather (weather.go)
   - `T` → telemetry (telemetry.go)

### Key Design Patterns

- **Optional fields use pointers**: `*float64`, `*int`, `*time.Time` — nil means not present in packet
- **Telemetry values**: `[]*float64` where nil elements represent undefined channels
- **Error reporting**: Parse failures return `*ParseError` (defined in errors.go) with `Code` and `Msg` fields. Use `errors.Is(err, fap.ErrXxx)` to check for specific error codes
- **Comment parsing** extracts embedded data (altitude, DAO, weather, base-91 telemetry) then stores the remainder in `Packet.Comment`
- **Base-91 telemetry** in Mic-E comments uses `|...|` delimiters with LSB-first bit order (matching Perl's `unpack('b8')`)

The original module is in ../perl-aprs-fap/FAP.pm .
When implementing tests, examples are found in the original module tests:
../perl-aprs-fap/t 
Original sample packets in the original tests MUST be used without
modification. Original expected parsed outcome values must also be used.

### Test Style

Individual test functions or table-driven tests with `t.Run` subtests,
using `t.Fatalf`/`t.Errorf` directly.  Table-driven tests are preferred
when multiple cases share the same assertion logic.
Helper: `approxEqual(a, b, tolerance)` for float comparison.
Use `new(val)` for pointer creation in tests.
