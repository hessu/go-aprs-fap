package fap

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hessu/go-aprs-fap/aprsis2"
)

// defaultLoginTimeout is how long Login waits for the server's logresp.
const defaultLoginTimeout = 5 * time.Second

// Conn represents a connection to an APRS-IS server. It speaks either the
// traditional plaintext protocol or the binary IS2 protocol; see IS2.
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

	// IS2 state, see aprsis_is2.go
	is2          bool // the connection speaks IS2
	is2Upgrade   bool // Login offers the upgrade to IS2
	signature    *aprsis2.IS2ServerSignature
	loginReply   *aprsis2.IS2LoginReply
	queue        []*aprsis2.ISPacket // packets of a frame not returned yet
	requestID    int32
	onIS2Message func(m *aprsis2.IS2Message)
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
	// generated "user ... pass ... vers ..." line. The upgrade to IS2 is
	// not offered with it, and it is not used on an IS2 port.
	LoginLine string
	// OnComment, if set, is called with every line starting with '#'
	// that is read from the server, including the banner and logresp.
	OnComment func(line string)

	// IS2 makes the connection speak the binary IS2 protocol from the
	// start, for connecting to an IS2 port of the server.
	IS2 bool
	// NoIS2Upgrade disables the upgrade of a plaintext connection to IS2.
	// By default Login offers the upgrade to the server, and the
	// connection switches to IS2 if the server supports it; servers which
	// do not know of IS2 ignore the offer. See Conn.IS2.
	NoIS2Upgrade bool
	// OnIS2Message, if set, is called with every IS2 message read from
	// the server other than the packets: the server signature, login
	// reply, keepalive pings (which are answered automatically) and
	// parameter replies.
	OnIS2Message func(m *aprsis2.IS2Message)
}

// LogResp is a parsed "# logresp" line, or the equivalent result of an IS2
// login.
type LogResp struct {
	Raw      string // the whole line; empty on IS2
	Callsign string
	Verified bool
	Server   string // the "server" field, if present

	// IS2 is set if the login was done over IS2: on an IS2 port, or after
	// the server upgraded the connection.
	IS2 bool
	// FilterRejected is set if the server did not apply the filter given
	// in ConnOptions, as on a port which does not allow user filters.
	// Only known on IS2; a plaintext server does not tell.
	FilterRejected bool
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
//
// The connection is upgraded to the IS2 protocol if the server supports
// that, unless o.NoIS2Upgrade is set; with o.IS2 it is IS2 from the start.
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

	// an IS2 frame has to fit in the read buffer
	bufSize := 4096
	if o.IS2 || !o.NoIS2Upgrade {
		bufSize = is2ReadBufferSize
	}

	c := &Conn{
		conn:         tc,
		reader:       bufio.NewReaderSize(tc, bufSize),
		callsign:     o.Callsign,
		passcode:     o.Passcode,
		appName:      o.AppName,
		appVer:       o.AppVer,
		filter:       o.Filter,
		loginLine:    o.LoginLine,
		loginTimeout: o.LoginTimeout,
		onComment:    o.OnComment,
		is2:          o.IS2,
		is2Upgrade:   !o.IS2 && !o.NoIS2Upgrade,
		onIS2Message: o.OnIS2Message,
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
//
// Unless disabled with NoIS2Upgrade, the login line is preceded by an offer
// to upgrade the connection to IS2. If the server accepts it, or the
// connection is to an IS2 port, the login is done over IS2: the LogResp is
// then built from the server's login reply, and a login which the server
// rejects, closing the connection, is an *IS2LoginError.
func (c *Conn) Login() (*LogResp, error) {
	timeout := c.loginTimeout
	if timeout == 0 {
		timeout = defaultLoginTimeout
	}
	deadline := time.Now().Add(timeout)

	if c.is2 {
		return c.loginIS2(deadline)
	}

	// The upgrade is offered on a comment line right before the login
	// line, without waiting for an answer: a server which does not
	// upgrade ignores it and answers the login line as usual.
	upgrade := c.is2Upgrade && c.loginLine == ""

	login := c.loginLine
	if login == "" {
		login = fmt.Sprintf("user %s pass %s vers %s %s", c.callsign, c.passcode, c.appName, c.appVer)
		if c.filter != "" {
			login += " filter " + c.filter
		}
	}

	if upgrade {
		login = is2UpgradeHello + "\r\n" + login
	}
	if err := c.SendLine(login); err != nil {
		return nil, fmt.Errorf("failed to send login: %w", err)
	}

	for time.Now().Before(deadline) {
		line, err := c.ReadLine(time.Until(deadline))
		if err != nil {
			return nil, fmt.Errorf("failed to read login response: %w", err)
		}
		if upgrade && isIS2HelloAnswer(line) {
			// the last plaintext line from the server
			if err := checkIS2HelloAnswer(line); err != nil {
				return nil, err
			}
			c.is2 = true
			return c.loginIS2(deadline)
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
//
// On an IS2 connection there are no lines or comments, and ReadLine
// returns the next packet, like ReadPacket.
func (c *Conn) ReadLine(timeout time.Duration) (string, error) {
	if c.is2 {
		return c.ReadPacket(timeout)
	}

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
//
// On an IS2 connection it returns the data of the next packet, which is the
// same line. The timeout then applies to each frame read, so the server's
// keepalive pings keep it from expiring like the keepalive comments of a
// plaintext connection do.
func (c *Conn) ReadPacket(timeout time.Duration) (string, error) {
	if c.is2 {
		p, err := c.ReadISPacket(timeout)
		if err != nil {
			return "", err
		}
		return string(p.GetIsPacketData()), nil
	}

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
//
// On an IS2 connection the line is sent as a packet. Comment lines (starting
// with '#') do not exist in IS2 and are not sent, except that a
// "#filter ..." command is sent as the equivalent parameter request, so that
// a client written for plaintext works the same after an upgrade to IS2.
func (c *Conn) SendLine(line string) error {
	if c.is2 {
		return c.sendLineIS2(line)
	}

	_, err := fmt.Fprintf(c.conn, "%s\r\n", line)
	return err
}

// SendRaw writes the bytes to the connection as they are, without adding a
// line terminator or IS2 framing.
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
