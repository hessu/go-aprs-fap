package fap

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// defaultLoginTimeout is how long Login waits for the server's logresp.
const defaultLoginTimeout = 5 * time.Second

// Conn represents a connection to an APRS-IS server.
type Conn struct {
	conn     net.Conn
	reader   *bufio.Reader
	callsign string
	passcode string
	appName  string
	appVer   string
	filter   string

	loginLine    string
	loginTimeout time.Duration
	onComment    func(line string)
	banner       string
}

// ConnOptions configures Connect. The zero value, with the login fields
// filled in, behaves like Dial.
type ConnOptions struct {
	Callsign string
	Passcode string
	AppName  string
	AppVer   string
	Filter   string

	// DialTimeout limits connection establishment. Zero means no limit.
	DialTimeout time.Duration
	// LoginTimeout limits the wait for the server's logresp in Login.
	// Zero means 5 seconds.
	LoginTimeout time.Duration
	// TLS, if non-nil, makes the connection use TLS.
	TLS *tls.Config
	// Dialer, if non-nil, is used to open the underlying connection
	// instead of net.Dial. DialTimeout and TLS are not applied to it.
	Dialer func(network, addr string) (net.Conn, error)

	// NoLogin makes Connect return right after connecting, without
	// logging in. The caller can then use SendLine, SendRaw and ReadLine
	// to drive the handshake, or call Login.
	NoLogin bool
	// LoginLine, if set, is sent verbatim by Login instead of the
	// generated "user ... pass ... vers ..." line.
	LoginLine string
	// OnComment, if set, is called with every line starting with '#'
	// that is read from the server, including the banner and logresp.
	OnComment func(line string)
}

// LogResp is a parsed "# logresp" line.
type LogResp struct {
	Raw      string // the whole line
	Callsign string
	Verified bool
	Server   string // the "server" field, if present
}

// Dial connects to an APRS-IS server, sends the login line, and waits
// for a "# logresp" reply. An optional filter string can be provided.
func Dial(addr, callsign, passcode, appName, appVer string, filter ...string) (*Conn, error) {
	o := ConnOptions{
		Callsign: callsign,
		Passcode: passcode,
		AppName:  appName,
		AppVer:   appVer,
	}
	if len(filter) > 0 {
		o.Filter = filter[0]
	}

	return Connect(addr, o)
}

// Connect opens a connection to an APRS-IS server as configured by o. Unless
// o.NoLogin is set, it also logs in and waits for the logresp, like Dial.
func Connect(addr string, o ConnOptions) (*Conn, error) {
	var tc net.Conn
	var err error

	switch {
	case o.Dialer != nil:
		tc, err = o.Dialer("tcp", addr)
	case o.TLS != nil:
		d := &net.Dialer{Timeout: o.DialTimeout}
		tc, err = tls.DialWithDialer(d, "tcp", addr, o.TLS)
	default:
		tc, err = net.DialTimeout("tcp", addr, o.DialTimeout)
	}
	if err != nil {
		return nil, err
	}

	c := &Conn{
		conn:         tc,
		reader:       bufio.NewReader(tc),
		callsign:     o.Callsign,
		passcode:     o.Passcode,
		appName:      o.AppName,
		appVer:       o.AppVer,
		filter:       o.Filter,
		loginLine:    o.LoginLine,
		loginTimeout: o.LoginTimeout,
		onComment:    o.OnComment,
	}

	if o.NoLogin {
		return c, nil
	}

	if _, err := c.Login(); err != nil {
		tc.Close()
		return nil, err
	}

	return c, nil
}

// Login sends the login line and waits for the server's "# logresp" reply,
// returning it parsed. A rejected or unverified login is not an error; check
// the returned LogResp. Other comment lines read while waiting (such as the
// server banner) are passed to OnComment.
func (c *Conn) Login() (*LogResp, error) {
	login := c.loginLine
	if login == "" {
		login = fmt.Sprintf("user %s pass %s vers %s %s", c.callsign, c.passcode, c.appName, c.appVer)
		if c.filter != "" {
			login += " filter " + c.filter
		}
	}

	if err := c.SendLine(login); err != nil {
		return nil, fmt.Errorf("failed to send login: %w", err)
	}

	timeout := c.loginTimeout
	if timeout == 0 {
		timeout = defaultLoginTimeout
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		line, err := c.ReadLine(time.Until(deadline))
		if err != nil {
			return nil, fmt.Errorf("failed to read login response: %w", err)
		}
		if strings.HasPrefix(line, "# logresp") {
			return parseLogResp(line), nil
		}
	}

	return nil, errors.New("login timed out waiting for logresp")
}

// parseLogResp parses "# logresp CALL verified, server NAME".
func parseLogResp(line string) *LogResp {
	r := &LogResp{Raw: line}

	rest := strings.TrimSpace(strings.TrimPrefix(line, "# logresp"))
	fields := strings.Fields(rest)
	if len(fields) > 0 {
		r.Callsign = fields[0]
	}
	if len(fields) > 1 {
		r.Verified = strings.TrimSuffix(fields[1], ",") == "verified"
	}
	for i, f := range fields {
		if f == "server" && i+1 < len(fields) {
			r.Server = strings.TrimSuffix(fields[i+1], ",")
		}
	}

	return r
}

// Banner returns the server's greeting line (the first comment line, usually
// "# aprsc 2.x.x ..."), or "" if none was seen yet. It is set while reading.
func (c *Conn) Banner() string {
	return c.banner
}

// ReadLine reads a single line from the connection, stripping the
// trailing CR/LF. The provided timeout sets a read deadline.
func (c *Conn) ReadLine(timeout time.Duration) (string, error) {
	err := c.conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", err
	}
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")

	if strings.HasPrefix(line, "#") {
		if c.banner == "" {
			c.banner = line
		}
		if c.onComment != nil {
			c.onComment(line)
		}
	}

	return line, nil
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

// SendRaw writes the bytes to the connection as they are, without adding a
// line terminator.
func (c *Conn) SendRaw(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

// CloseWrite shuts down the sending half of the connection (TCP half-close),
// so the server sees EOF while the client can still read. It returns an
// error if the underlying connection does not support it, as with TLS.
func (c *Conn) CloseWrite() error {
	cw, ok := c.conn.(interface{ CloseWrite() error })
	if !ok {
		return errors.New("connection does not support CloseWrite")
	}
	return cw.CloseWrite()
}

// NetConn returns the underlying network connection.
func (c *Conn) NetConn() net.Conn {
	return c.conn
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
