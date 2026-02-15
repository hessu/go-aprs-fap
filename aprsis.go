package fap

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Conn represents a connection to an APRS-IS server.
type Conn struct {
	conn     net.Conn
	reader   *bufio.Reader
	callsign string
	passcode string
	appName  string
	appVer   string
	filter   string
}

// Dial connects to an APRS-IS server, sends the login line, and waits
// for a "# logresp" reply. An optional filter string can be provided.
func Dial(addr, callsign, passcode, appName, appVer string, filter ...string) (*Conn, error) {
	tc, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &Conn{
		conn:     tc,
		reader:   bufio.NewReader(tc),
		callsign: callsign,
		passcode: passcode,
		appName:  appName,
		appVer:   appVer,
	}

	if len(filter) > 0 {
		c.filter = filter[0]
	}

	// Build login line.
	login := fmt.Sprintf("user %s pass %s vers %s %s", callsign, passcode, appName, appVer)
	if c.filter != "" {
		login += " filter " + c.filter
	}

	if err := c.SendLine(login); err != nil {
		tc.Close()
		return nil, fmt.Errorf("failed to send login: %w", err)
	}

	// Wait for logresp with a timeout.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		line, err := c.ReadLine(time.Until(deadline))
		if err != nil {
			tc.Close()
			return nil, fmt.Errorf("failed to read login response: %w", err)
		}
		if strings.HasPrefix(line, "# logresp") {
			return c, nil
		}
	}

	tc.Close()
	return nil, fmt.Errorf("login timed out waiting for logresp")
}

// ReadLine reads a single line from the connection, stripping the
// trailing CR/LF. The provided timeout sets a read deadline.
func (c *Conn) ReadLine(timeout time.Duration) (string, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// ReadPacket reads lines until a non-comment line (one that does not
// start with '#') is found, and returns it. Comment lines are silently
// skipped. The timeout applies to each individual read.
func (c *Conn) ReadPacket(timeout time.Duration) (string, error) {
	for {
		line, err := c.ReadLine(timeout)
		if err != nil {
			return "", err
		}
		if !strings.HasPrefix(line, "#") {
			return line, nil
		}
	}
}

// SendLine writes a line followed by CR/LF to the connection.
func (c *Conn) SendLine(line string) error {
	_, err := fmt.Fprintf(c.conn, "%s\r\n", line)
	return err
}

// Close closes the underlying TCP connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// AprsPasscode calculates the APRS-IS passcode for a given callsign.
// The SSID suffix is stripped and the callsign is uppercased before
// computing the hash.
func AprsPasscode(callsign string) int16 {
	// Strip SSID (everything from first '-' onward).
	if i := strings.IndexByte(callsign, '-'); i >= 0 {
		callsign = callsign[:i]
	}
	callsign = strings.ToUpper(callsign)

	var hash uint16 = 0x73E2 // 29666
	for i := 0; i < len(callsign); i += 2 {
		hash ^= uint16(callsign[i]) << 8
		if i+1 < len(callsign) {
			hash ^= uint16(callsign[i+1])
		}
	}

	return int16(hash & 0x7FFF)
}
