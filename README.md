# go-aprs-fap

Finnish APRS Parser (Fabulous APRS Parser) - Go edition

This is a Go port of the [Ham::APRS::FAP](https://metacpan.org/pod/Ham::APRS::FAP)
Perl module, a fairly complete APRS (Automatic Packet Reporting System)
parser. It parses normal, Mic-E and compressed location packets, NMEA
location packets, objects, items, messages, telemetry and most weather
packets. It is implemented in pure Go; the only external dependency is
the Go Protocol Buffers module, for the IS2 protocol of the APRS-IS
client.

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

### Warnings

Some issues are non-fatal: parsing succeeds but the packet has problems
worth noting (e.g. an invalid timestamp). These are collected in the
`Warnings` slice on the returned `Packet`:

```go
p, err := fap.Parse(raw)
if err != nil {
    // hard failure
}
for _, w := range p.Warnings {
    fmt.Printf("warning: code=%s msg=%s\n", w.Code, w.Msg)
}
```

Warnings use the same `ParseError` type as errors, so you can match
specific codes with `errors.Is()`:

```go
for i := range p.Warnings {
    if errors.Is(&p.Warnings[i], fap.ErrTimestampInvalid) {
        // timestamp could not be parsed, but the rest of the packet is fine
    }
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
- `fap.Connect(addr, fap.ConnOptions{...})` — like `Dial`, with dial/login timeouts, TLS, a custom dialer, a `NoLogin` mode for driving the handshake by hand, a verbatim `LoginLine`, and an `OnComment` hook for `#` lines
- `Conn.Login()` — send the login line and return the parsed `*fap.LogResp` (callsign, verified, server); a rejected login is not an error
- `Conn.Banner()` — the server's first `#` greeting line
- `Conn.SendRaw(b)`, `Conn.CloseWrite()`, `Conn.NetConn()` — raw writes, TCP half-close, and the underlying connection
- `Conn.ReadLine(timeout)` — read one line (strips CR/LF)
- `Conn.ReadPacket(timeout)` — read one non-comment line (skips `#` keepalives)
- `Conn.SendLine(line)` — send a line (appends CR/LF)
- `Conn.Close()` — close the connection
- `Conn.SetFilter(filter)` — change the filter after login
- `fap.AprsPasscode(callsign)` — compute the APRS-IS passcode for a callsign

### IS2

The client also speaks IS2, the binary protocol of
[aprsc](https://github.com/hessu/aprsc) which carries the packets in
Protocol Buffers messages, together with metadata: which servers received
the packet and when, and how the receiver heard it (RSSI, SNR, modem, bit
rate, frequency). The Go types of the protocol are in the `aprsis2`
package, generated from a copy of `aprsis2.proto` of aprsc.

A connection gets to IS2 in one of two ways:

- **Upgrade, by default.** When logging in on an ordinary plaintext
  APRS-IS port, the client offers an upgrade to IS2 on a comment line
  before its login line. A server which supports it switches the connection
  to IS2; other servers ignore the comment and the connection stays on
  plaintext. Set `ConnOptions.NoIS2Upgrade` to not offer it. It is not
  offered with a custom `LoginLine` either.
- **An IS2 port.** Set `ConnOptions.IS2` to connect to a port of the server
  which speaks IS2 from the start.

`Conn.IS2()` tells which protocol the connection ended up with. The same
methods work with both, so a client does not need to care: `ReadPacket`
returns packets, and `SendLine` sends them. On IS2, `ReadLine` returns
packets too, since there are no comment lines; comment lines given to
`SendLine` are not sent, except for `#filter ...`, which is sent as the IS2
request to set the filter.

```go
c, err := fap.Connect("aprsc.example.org:14580", fap.ConnOptions{
    Callsign: "N0CALL", Passcode: "-1", AppName: "myapp", AppVer: "0.1",
    Filter: "r/60.18/24.94/100",
})
if err != nil {
    return err
}
defer c.Close()

for {
    isp, err := c.ReadISPacket(30 * time.Second)
    if err != nil {
        return err
    }

    // the parsed packet, with the ISPacket in p.IS2
    p, err := fap.ParseISPacket(isp)
    if err != nil {
        fmt.Printf("Parse error: %v\n", err)
        continue
    }

    fmt.Printf("%s> type=%s\n", p.SrcCallsign, p.Type)
    if _, ok := p.IS2.GetOptionalRxRssiDbm().(*aprsis2.ISPacket_RxRssiDbm); ok {
        fmt.Printf("  RSSI %.1f dBm\n", p.IS2.GetRxRssiDbm())
    }
    for _, r := range p.IS2.GetReceived() {
        fmt.Printf("  received by %s from %s\n", r.GetServerId(), r.GetUsername())
    }
}
```

`fap.ParseISPacket` works on a plaintext connection as well: `ReadISPacket`
then returns the packet without metadata, so the same loop serves both
protocols. The getters of `p.IS2` are safe to call on a packet parsed with
`fap.Parse`, where `p.IS2` is nil.

- `Conn.IS2()` — whether the connection speaks IS2
- `fap.ParseISPacket(p, opts...)` — parse the APRS packet of an `*aprsis2.ISPacket` like `fap.Parse`, keeping the `ISPacket` and its metadata in `Packet.IS2`
- `Conn.ReadISPacket(timeout)` — read one packet as an `*aprsis2.ISPacket`, with its metadata; on plaintext, the packet alone
- `Conn.SendISPacket(p)` — send a packet with its reception metadata; on plaintext, the metadata goes on an `#is2-meta` comment line, which aprsc accepts on its igate and client ports
- `Conn.ServerSignature()`, `Conn.LoginReply()` — the server's IS2 signature and login reply
- `Conn.SendIS2Message(m)` — send any IS2 message
- `ConnOptions.OnIS2Message` — a hook for the IS2 messages received other than packets, like `OnComment` for plaintext
- `LogResp.IS2`, `LogResp.FilterRejected` — the login was done over IS2; the server refused the filter given at login
- A login the server rejects over IS2 is an `*fap.IS2LoginError`, with the reason code of the reply; `SetFilter` returns `fap.ErrFilterRejected` if the server does not apply the filter

The server's IS2 keepalive pings are answered while reading, so a client
must keep reading from the connection, as it does anyway. A `Conn` can be
read by one goroutine and written to by another, but not read by several.

## Position encoding

`EncodePosition` creates an uncompressed APRS position body string.
Coordinates are in decimal degrees, speed in km/h, course in degrees,
altitude in meters. Symbol is a 2-character string (table + code).

```go
lat := 60.4525
lon := 24.9842
speed := 45.0
course := 180.0
alt := 120.0

body, err := fap.EncodePosition(lat, lon, &speed, &course, &alt, "/>", nil)
// body is "!6027.15N/02459.05E/>180/024/A=000394"
```

Optional parameters are passed via `EncodePositionOpts`:

```go
body, err := fap.EncodePosition(lat, lon, nil, nil, nil, "/-", &fap.EncodePositionOpts{
    Ambiguity: 2,     // 0-4, reduces coordinate precision
    Comment:   "Test",
})
```

Include an HMS UTC timestamp with the `Timestamp` option. This changes the
data type identifier from `!` to `/` and prepends an `HHMMSSh` timestamp:

```go
body, err := fap.EncodePosition(lat, lon, nil, nil, nil, "/-", &fap.EncodePositionOpts{
    Timestamp: time.Date(2024, 3, 15, 12, 30, 45, 0, time.UTC),
})
// body is "/123045h6027.15N/02459.05E-"
```

Enable DAO for extra precision using the `!DAO!` extension:

```go
body, err := fap.EncodePosition(lat, lon, nil, nil, nil, "/-", &fap.EncodePositionOpts{
    DAO: true,
})
```

## Message encoding

`EncodeMessage` encodes a `Message` struct into an APRS message body string
that can be parsed back by the packet parser.

```go
msg := &fap.Message{
    Destination: "N0CALL",
    Text:        "Hello world",
    ID:          "42",
}

body, err := fap.EncodeMessage(msg)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
// body is ":N0CALL   :Hello world{42"
```

It supports all message variants: regular messages, acks, rejects, and
reply-acks. A message with `ReplyAck` set and no `AckID` is encoded as
`{ID}`, telling that the sender supports reply-acks; the parser sets
`ReplyAck` on messages in the reply-ack format. Validation is performed
on all fields:

| Sentinel | Condition |
|---|---|
| `ErrMsgNoDst` | Destination is empty |
| `ErrMsgDstTooLong` | Destination exceeds 9 characters |
| `ErrMsgIDInvalid` | Message ID is not 1-5 alphanumeric characters |
| `ErrMsgReplyAck` | Reply-ack ID + '}' + ack ID exceeds 5 characters |
| `ErrMsgAckRej` | Both ack and reject set, or content with reject |
| `ErrMsgCRLF` | Any field contains CR or LF |

```go
// Send an ack
ack := &fap.Message{Destination: "N0CALL", AckID: "42"}
body, _ := fap.EncodeMessage(ack)

// Send a reject
rej := &fap.Message{Destination: "N0CALL", RejID: "42"}
body, _ = fap.EncodeMessage(rej)

// Reply-ack (embedded ack in message)
msg := &fap.Message{Destination: "N0CALL", Text: "Hi", ID: "1", AckID: "ab"}
body, err := fap.EncodeMessage(msg)
if errors.Is(err, fap.ErrMsgReplyAck) {
    // ack doesn't fit, send it separately
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
