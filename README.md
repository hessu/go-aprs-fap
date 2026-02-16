# go-aprs-fap

Finnish APRS Parser (Fabulous APRS Parser) - Go edition

This is a Go port of the [Ham::APRS::FAP](https://metacpan.org/pod/Ham::APRS::FAP)
Perl module, a fairly complete APRS (Automatic Packet Reporting System)
parser. It parses normal, Mic-E and compressed location packets, NMEA
location packets, objects, items, messages, telemetry and most weather
packets. It is implemented in pure Go with no external dependencies
(stdlib only).

The original Perl module is stable and fast enough to parse the APRS-IS
stream in real time, and is used to power the <http://aprs.fi/> web site.

But nobody actually tested or used this Go port yet. I have no clue whether
it works.

## Performance

The Go port parses packets 19 times faster than the original Perl version.
For a constant incoming stream of packets, it uses about 5% of the CPU used
by the Perl parser.  This was tested by parsing the same 24-hour log of
APRS-IS packets using both parsers.

On an old Intel Xeon L5520 CPU @ 2.27GHz (8 MB cache), it does about 230k
packets/second, which is sufficient for my current needs.

## AI-assisted port

This Go module was created using the Claude Code AI agent, by letting it
inspect the old Perl code.  The original Perl module's test packets and
expected output values have been retained in the Go test suite, which should
provide a good level of confidence that the port produces correct results.

## Supported packet types

- Position (uncompressed, compressed, Mic-E)
- Objects and items
- Messages, acks, and rejects
- Weather reports
- Telemetry
- Status reports
- Station capabilities
- NMEA (GPRMC, GPGGA, GPGLL)
- DX spots

## Not handled

- Special objects (area, signpost, etc)
- Network tunneling / third party packets
- Direction finding
- Station capability queries
- User defined data formats

This module is based (on those parts that are implemented) on APRS
specification 1.0.1.

## Installation

```
go get github.com/hessu/go-aprs-fap
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/hessu/go-aprs-fap"
)

func main() {
    packet := "N0CALL>APRS,WIDE1-1,WIDE2-1,qAo,IGATE:!6128.23N/02353.52E-PHG2360/Testing"
    p, err := fap.Parse(packet)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Source: %s\n", p.SrcCallsign)
    fmt.Printf("Type: %s\n", p.Type)
    fmt.Printf("Lat: %.4f\n", *p.Latitude)
    fmt.Printf("Lon: %.4f\n", *p.Longitude)
}
```

## Error handling

Parse errors are returned as `*fap.ParseError` values, which carry a
machine-readable `Code` (e.g. `"pos_short"`) and a human-readable `Msg`.
Sentinel error variables (e.g. `fap.ErrPosShort`) are provided for use
with `errors.Is()`:

```go
import "errors"

p, err := fap.Parse(raw)
if errors.Is(err, fap.ErrLocInvalid) {
    // invalid location
} else if err != nil {
    // some other parse failure
}
```

To access the error code and message directly, use `errors.As()`:

```go
var parseErr *fap.ParseError
if errors.As(err, &parseErr) {
    fmt.Printf("code=%s msg=%s\n", parseErr.Code, parseErr.Msg)
}
```

## APRS-IS client

The package includes an APRS-IS TCP client for connecting to APRS-IS
servers. It is a Go port of the
[Ham::APRS::IS](https://metacpan.org/pod/Ham::APRS::IS) Perl module.

Do provide the name of your application in place of "myapp", and
version number in place of "0.1" in the following example.

```go
package main

import (
    "fmt"
    "time"
    "github.com/hessu/go-aprs-fap"
)

func main() {
    c, err := fap.Dial("rotate.aprs2.net:14580", "N0CALL", "-1", "myapp", "0.1", "r/60.18/24.94/100")
    if err != nil {
        fmt.Printf("Connect error: %v\n", err)
        return
    }
    defer c.Close()

    for {
        line, err := c.ReadPacket(30 * time.Second)
        if err != nil {
            fmt.Printf("Read error: %v\n", err)
            return
        }

        p, err := fap.Parse(line)
        if err != nil {
            fmt.Printf("Parse error: %v\n", err)
            continue
        }

        fmt.Printf("%s> type=%s\n", p.SrcCallsign, p.Type)
    }
}
```

### Functions

- `fap.Dial(addr, callsign, passcode, appName, appVer, filter...)` — connect, authenticate, and return a `*Conn`
- `Conn.ReadLine(timeout)` — read one line (strips CR/LF)
- `Conn.ReadPacket(timeout)` — read one non-comment line (skips `#` keepalives)
- `Conn.SendLine(line)` — send a line (appends CR/LF)
- `Conn.Close()` — close the connection
- `fap.AprsPasscode(callsign)` — compute the APRS-IS passcode for a callsign

## See also

- [Ham::APRS::FAP](https://metacpan.org/pod/Ham::APRS::FAP) - the original Perl module
- [libfap](http://pakettiradio.net/libfap/) - C library port of Ham::APRS::FAP
- [python-libfap](http://github.com/kd7lxl/python-libfap) - Python bindings for libfap

## Copyright and licence

Copyright (C) 2005-2026 Tapio Sokura, OH2KKU

Copyright (C) 2007-2026 Heikki Hannikainen, OH7LZB

This library is free software; you can redistribute it and/or modify
it under the same terms Perl itself.
