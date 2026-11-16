package fap

// The IS2 side of the APRS-IS client: the binary protocol of aprsc, with
// Protocol Buffers messages in frames. A Conn speaks it either from the
// start, on an IS2 port of the server (ConnOptions.IS2), or after the server
// has accepted the upgrade of a plaintext connection, which Login offers by
// default.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/hessu/go-aprs-fap/aprsis2"
)

// IS2 framing: STX, a reserved zero byte, the length of the message as a
// 16-bit big-endian integer, the encoded IS2Message, ETX.
const (
	is2STX = 0x02
	is2ETX = 0x03

	is2HeadLen = 4
	is2TailLen = 1

	// is2MaxMessageLen is the longest message the framing can carry.
	// aprsc accepts frames of up to 4096 bytes in all.
	is2MaxMessageLen = 0xffff

	// is2ReadBufferSize has room for the longest frame.
	is2ReadBufferSize = is2HeadLen + is2MaxMessageLen + is2TailLen

	// is2UpgradeHello offers the upgrade of a plaintext connection to
	// IS2, listing the versions of the upgrade supported. The server
	// answers with the same line, with the version it picked.
	is2UpgradeHello  = "#is2 hello v=1"
	is2HelloPrefix   = "#is2 hello"
	is2UpgradeVerMax = 1
)

// ErrIS2Framing is returned when the data received is not a valid IS2 frame.
// The connection cannot be used after it.
var ErrIS2Framing = errors.New("invalid IS2 frame received")

// ErrFilterRejected is returned by SetFilter when the server does not apply
// the filter, as on a port which does not allow user filters.
var ErrFilterRejected = errors.New("the server did not apply the filter")

// IS2LoginError is an IS2 login which the server rejected. The server closes
// the connection after it.
type IS2LoginError struct {
	Reply *aprsis2.IS2LoginReply
}

func (e *IS2LoginError) Error() string {
	return fmt.Sprintf("IS2 login rejected: %v: %s", e.Reply.GetResultCode(), e.Reply.GetResultMessage())
}

// IS2 reports whether the connection speaks the IS2 protocol: it was opened
// with ConnOptions.IS2, or the server upgraded it during Login.
//
// The methods of Conn work the same in both cases, as far as the protocols
// allow: ReadPacket and SendLine carry packets, and ReadISPacket and
// SendISPacket give access to the metadata IS2 carries with them.
func (c *Conn) IS2() bool {
	return c.is2
}

// ServerSignature returns the signature the server sent at the start of an
// IS2 connection, with its ID and software version, or nil on plaintext.
func (c *Conn) ServerSignature() *aprsis2.IS2ServerSignature {
	return c.signature
}

// LoginReply returns the server's reply to the IS2 login, or nil on
// plaintext.
func (c *Conn) LoginReply() *aprsis2.IS2LoginReply {
	return c.loginReply
}

// isIS2HelloAnswer tells whether a line is the server's answer to the hello.
func isIS2HelloAnswer(line string) bool {
	return line == is2HelloPrefix || strings.HasPrefix(line, is2HelloPrefix+" ")
}

// checkIS2HelloAnswer checks that the server picked a version of the upgrade
// which was offered. The line has space-separated key=value pairs, and keys
// which are not known are ignored.
func checkIS2HelloAnswer(line string) error {
	for _, kv := range strings.Fields(line[len(is2HelloPrefix):]) {
		v, ok := strings.CutPrefix(kv, "v=")
		if !ok {
			continue
		}
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= is2UpgradeVerMax {
			return nil
		}
	}

	return fmt.Errorf("server answered the IS2 upgrade with an unsupported version: %q", line)
}

// loginIS2 does the IS2 login: the server sends its signature first, then
// the client its login request, and the server answers it.
func (c *Conn) loginIS2(deadline time.Time) (*LogResp, error) {
	m, err := c.readMessage(time.Until(deadline))
	if err != nil {
		return nil, fmt.Errorf("failed to read the IS2 server signature: %w", err)
	}
	if m.GetType() != aprsis2.IS2Message_SERVER_SIGNATURE || m.GetServerSignature() == nil {
		return nil, fmt.Errorf("expected an IS2 server signature, got %v", m.GetType())
	}
	c.signature = m.GetServerSignature()

	err = c.sendMessage(&aprsis2.IS2Message{
		Type: aprsis2.IS2Message_LOGIN_REQUEST,
		LoginRequest: &aprsis2.IS2LoginRequest{
			Username:   c.callsign,
			Password:   c.passcode,
			AppName:    c.appName,
			AppVersion: c.appVer,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send login: %w", err)
	}

	m, err = c.readMessage(time.Until(deadline))
	if err != nil {
		return nil, fmt.Errorf("failed to read login response: %w", err)
	}
	if m.GetType() != aprsis2.IS2Message_LOGIN_REPLY || m.GetLoginReply() == nil {
		return nil, fmt.Errorf("expected an IS2 login reply, got %v", m.GetType())
	}
	c.loginReply = m.GetLoginReply()
	if c.loginReply.GetResult() != aprsis2.IS2LoginReply_OK {
		return nil, &IS2LoginError{Reply: c.loginReply}
	}

	r := &LogResp{
		Callsign: c.callsign,
		Verified: c.loginReply.GetVerified() != aprsis2.VerificationStatus_NONE,
		Server:   c.signature.GetUsername(),
		IS2:      true,
	}

	// There is no filter in the IS2 login request, it is set separately.
	// A server refusing it is not a login failure, as on plaintext.
	if c.filter != "" {
		err := c.setFilterIS2(c.filter, deadline)
		if errors.Is(err, ErrFilterRejected) {
			r.FilterRejected = true
		} else if err != nil {
			return nil, fmt.Errorf("failed to set the filter: %w", err)
		}
	}

	return r, nil
}

// readFrame reads one IS2 frame. The frame is only consumed from the buffer
// when all of it has arrived, so a timeout does not lose a part of one.
func (c *Conn) readFrame(timeout time.Duration) (*aprsis2.IS2Message, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	hdr, err := c.reader.Peek(is2HeadLen)
	if err != nil {
		return nil, err
	}
	if hdr[0] != is2STX || hdr[1] != 0 {
		return nil, fmt.Errorf("%w: no STX and zero reserved byte at the start", ErrIS2Framing)
	}
	length := int(binary.BigEndian.Uint16(hdr[2:4]))

	frame, err := c.reader.Peek(is2HeadLen + length + is2TailLen)
	if err != nil {
		return nil, err
	}
	if frame[is2HeadLen+length] != is2ETX {
		return nil, fmt.Errorf("%w: no ETX at the end", ErrIS2Framing)
	}

	m := &aprsis2.IS2Message{}
	err = proto.Unmarshal(frame[is2HeadLen:is2HeadLen+length], m)
	c.reader.Discard(len(frame))
	if err != nil {
		return nil, fmt.Errorf("failed to decode IS2 message: %w", err)
	}

	return m, nil
}

// readMessage reads frames until one which is not a keepalive ping, and
// answers the ping requests of the server. The timeout applies to each
// frame. All messages other than packets are passed to OnIS2Message.
func (c *Conn) readMessage(timeout time.Duration) (*aprsis2.IS2Message, error) {
	for {
		m, err := c.readFrame(timeout)
		if err != nil {
			return nil, err
		}

		if m.GetType() != aprsis2.IS2Message_IS_PACKET && c.onIS2Message != nil {
			c.onIS2Message(m)
		}
		if m.GetType() != aprsis2.IS2Message_KEEPALIVE_PING {
			return m, nil
		}

		// A ping request is sent back as a reply, with the same
		// request ID and data.
		ping := m.GetKeepalivePing()
		if ping.GetPingType() != aprsis2.IS2KeepalivePing_REQUEST {
			continue
		}
		err = c.sendMessage(&aprsis2.IS2Message{
			Type: aprsis2.IS2Message_KEEPALIVE_PING,
			KeepalivePing: &aprsis2.IS2KeepalivePing{
				PingType:    aprsis2.IS2KeepalivePing_REPLY,
				RequestId:   ping.GetRequestId(),
				RequestData: ping.GetRequestData(),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to answer an IS2 ping: %w", err)
		}
	}
}

// ReadISPacket returns the next packet as an IS2 ISPacket, which has, in
// addition to the packet itself (IsPacketData), the metadata the server
// passes with it: the servers which received it and when, and how the
// receiver heard it, where known. On a dupe feed port the Type of the
// packets is IS_PACKET_DUPLICATE.
//
// On a plaintext connection it returns the line read by ReadPacket in an
// ISPacket with no metadata. The timeout works like that of ReadPacket.
func (c *Conn) ReadISPacket(timeout time.Duration) (*aprsis2.ISPacket, error) {
	if !c.is2 {
		line, err := c.ReadPacket(timeout)
		if err != nil {
			return nil, err
		}
		return &aprsis2.ISPacket{Type: aprsis2.ISPacket_IS_PACKET, IsPacketData: []byte(line)}, nil
	}

	// a frame may carry several packets
	for len(c.queue) == 0 {
		m, err := c.readMessage(timeout)
		if err != nil {
			return nil, err
		}
		if m.GetType() == aprsis2.IS2Message_IS_PACKET {
			c.queue = m.GetIsPacket()
		}
	}

	p := c.queue[0]
	c.queue = c.queue[1:]

	return p, nil
}

// sendMessage encodes an IS2Message and sends it in a frame, in one write.
func (c *Conn) sendMessage(m *aprsis2.IS2Message) error {
	frame := make([]byte, is2HeadLen, is2HeadLen+proto.Size(m)+is2TailLen)
	frame, err := proto.MarshalOptions{}.MarshalAppend(frame, m)
	if err != nil {
		return err
	}

	length := len(frame) - is2HeadLen
	if length > is2MaxMessageLen {
		return fmt.Errorf("IS2 message of %d bytes is too long for a frame", length)
	}
	frame[0] = is2STX
	binary.BigEndian.PutUint16(frame[2:4], uint16(length))

	return c.SendRaw(append(frame, is2ETX))
}

// SendIS2Message sends an IS2 message of any type in a frame. It returns an
// error on a plaintext connection.
func (c *Conn) SendIS2Message(m *aprsis2.IS2Message) error {
	if !c.is2 {
		return errors.New("not an IS2 connection")
	}
	return c.sendMessage(m)
}

// SendISPacket sends a packet with its metadata: how the receiver heard
// the packet, such as the RSSI and SNR of a LoRa packet. A Type which is
// not set is taken as IS_PACKET.
//
// On a plaintext connection the reception metadata (the rx fields) is sent
// on an "#is2-meta" comment line before the packet, which aprsc understands
// and other servers ignore.
func (c *Conn) SendISPacket(p *aprsis2.ISPacket) error {
	if !c.is2 {
		line := string(p.GetIsPacketData()) + "\r\n"
		if meta := is2MetaLine(p); meta != "" {
			line = meta + "\r\n" + line
		}
		return c.SendRaw([]byte(line))
	}

	if p.GetType() == aprsis2.ISPacket_UNDEFINED {
		p = proto.Clone(p).(*aprsis2.ISPacket)
		p.Type = aprsis2.ISPacket_IS_PACKET
	}

	return c.sendMessage(&aprsis2.IS2Message{
		Type:     aprsis2.IS2Message_IS_PACKET,
		IsPacket: []*aprsis2.ISPacket{p},
	})
}

// is2MetaLine returns the "#is2-meta" line with the reception metadata of a
// packet for a plaintext connection, or "" if it has none.
func is2MetaLine(p *aprsis2.ISPacket) string {
	var b strings.Builder

	add := func(key, value string) {
		if b.Len() == 0 {
			b.WriteString("#is2-meta")
		}
		b.WriteString(" " + key + "=" + value)
	}

	if v, ok := p.GetOptionalRxRssiDbm().(*aprsis2.ISPacket_RxRssiDbm); ok {
		add("rx_rssi_dbm", strconv.FormatFloat(float64(v.RxRssiDbm), 'f', -1, 32))
	}
	if v, ok := p.GetOptionalRxSnrDb().(*aprsis2.ISPacket_RxSnrDb); ok {
		add("rx_snr_db", strconv.FormatFloat(float64(v.RxSnrDb), 'f', -1, 32))
	}
	if v := p.GetRxModem(); v != aprsis2.ISPacket_MODEM_UNDEFINED {
		add("rx_modem", strconv.Itoa(int(v)))
	}
	if v, ok := p.GetOptionalRxBitrate().(*aprsis2.ISPacket_RxBitrate); ok {
		add("rx_bitrate", strconv.FormatUint(uint64(v.RxBitrate), 10))
	}
	if v, ok := p.GetOptionalRxFrequency().(*aprsis2.ISPacket_RxFrequency); ok {
		add("rx_frequency", strconv.FormatFloat(v.RxFrequency, 'f', -1, 64))
	}

	return b.String()
}

// sendLineIS2 is SendLine on an IS2 connection.
func (c *Conn) sendLineIS2(line string) error {
	if strings.HasPrefix(line, "#") {
		// the plaintext filter command: "#filter r/60/25/100"
		if len(line) > 8 && strings.EqualFold(line[:8], "#filter ") {
			_, err := c.sendFilter(strings.TrimSpace(line[8:]))
			return err
		}
		return nil
	}

	return c.SendISPacket(&aprsis2.ISPacket{
		Type:         aprsis2.ISPacket_IS_PACKET,
		IsPacketData: []byte(line),
	})
}

// sendFilter sends a request to set the filter, and returns its request ID.
func (c *Conn) sendFilter(filter string) (int32, error) {
	c.requestID++
	id := c.requestID

	err := c.sendMessage(&aprsis2.IS2Message{
		Type: aprsis2.IS2Message_PARAMETER,
		Parameter: &aprsis2.IS2Parameter{
			Type:         aprsis2.IS2Parameter_PARAMETER_SET,
			RequestId:    id,
			FilterString: filter,
		},
	})

	return id, err
}

// setFilterIS2 sets the filter and waits for the server's reply to the
// request. Packets received while waiting are queued for ReadISPacket.
func (c *Conn) setFilterIS2(filter string, deadline time.Time) error {
	id, err := c.sendFilter(filter)
	if err != nil {
		return err
	}

	var queue []*aprsis2.ISPacket
	defer func() { c.queue = append(c.queue, queue...) }()

	for {
		m, err := c.readMessage(time.Until(deadline))
		if err != nil {
			return fmt.Errorf("no reply to the filter request: %w", err)
		}

		switch m.GetType() {
		case aprsis2.IS2Message_IS_PACKET:
			queue = append(queue, m.GetIsPacket()...)
		case aprsis2.IS2Message_PARAMETER:
			rep := m.GetParameter()
			if rep.GetRequestId() != id {
				continue
			}
			if rep.GetType() != aprsis2.IS2Parameter_PARAMETER_APPLIED {
				return ErrFilterRejected
			}
			return nil
		}
	}
}

// SetFilter changes the filter of the connection after login.
//
// On an IS2 connection it waits for the server's reply, for at most the
// login timeout, and returns ErrFilterRejected if the server did not apply
// the filter. It reads from the connection to do that, so it must not be
// called while another goroutine is reading; packets received while waiting
// are returned by the following reads. To change the filter while another
// goroutine reads, send it with SendLine("#filter ...") instead: the reply
// is then passed to OnIS2Message.
//
// On a plaintext connection it sends the "#filter" command and returns; the
// server answers with a comment line.
func (c *Conn) SetFilter(filter string) error {
	if !c.is2 {
		return c.SendLine("#filter " + filter)
	}

	timeout := c.loginTimeout
	if timeout == 0 {
		timeout = defaultLoginTimeout
	}

	return c.setFilterIS2(filter, time.Now().Add(timeout))
}
