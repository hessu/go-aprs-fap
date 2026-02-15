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

## See also

- [Ham::APRS::FAP](https://metacpan.org/pod/Ham::APRS::FAP) - the original Perl module
- [libfap](http://pakettiradio.net/libfap/) - C library port of Ham::APRS::FAP
- [python-libfap](http://github.com/kd7lxl/python-libfap) - Python bindings for libfap

## Copyright and licence

Copyright (C) 2005-2026 Tapio Sokura, OH2KKU
Copyright (C) 2007-2026 Heikki Hannikainen, OH7LZB

This library is free software; you can redistribute it and/or modify
it under the same terms Perl itself.
