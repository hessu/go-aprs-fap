package fap

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/hessu/go-aprs-fap/aprsis2"
)

// fakeIS2 is the server side of an IS2 connection in a test.
type fakeIS2 struct {
	conn net.Conn
	r    *bufio.Reader
}

func (f *fakeIS2) send(m *aprsis2.IS2Message) {
	b, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	frame := []byte{is2STX, 0, byte(len(b) >> 8), byte(len(b))}
	frame = append(append(frame, b...), is2ETX)
	f.conn.Write(frame)
}

func (f *fakeIS2) sendPackets(ps ...*aprsis2.ISPacket) {
	f.send(&aprsis2.IS2Message{Type: aprsis2.IS2Message_IS_PACKET, IsPacket: ps})
}

func (f *fakeIS2) read() (*aprsis2.IS2Message, error) {
	hdr := make([]byte, is2HeadLen)
	if _, err := io.ReadFull(f.r, hdr); err != nil {
		return nil, err
	}
	if hdr[0] != is2STX || hdr[1] != 0 {
		return nil, fmt.Errorf("bad frame header % x", hdr)
	}
	body := make([]byte, int(binary.BigEndian.Uint16(hdr[2:]))+is2TailLen)
	if _, err := io.ReadFull(f.r, body); err != nil {
		return nil, err
	}
	if body[len(body)-1] != is2ETX {
		return nil, errors.New("no ETX at the end of the frame")
	}
	m := &aprsis2.IS2Message{}
	return m, proto.Unmarshal(body[:len(body)-1], m)
}

// login does the server side of the login: signature, request, reply. It
// returns the login request.
func (f *fakeIS2) login(result aprsis2.IS2LoginReply_LoginResult, verified aprsis2.VerificationStatus) (*aprsis2.IS2LoginRequest, error) {
	f.send(&aprsis2.IS2Message{
		Type: aprsis2.IS2Message_SERVER_SIGNATURE,
		ServerSignature: &aprsis2.IS2ServerSignature{
			Username: "T2TEST", AppName: "fake", AppVersion: "1.0",
		},
	})
	// a ping may come before the login reply, and has to be answered
	f.send(&aprsis2.IS2Message{
		Type: aprsis2.IS2Message_KEEPALIVE_PING,
		KeepalivePing: &aprsis2.IS2KeepalivePing{
			PingType: aprsis2.IS2KeepalivePing_REQUEST, RequestId: 42, RequestData: []byte("abc"),
		},
	})

	m, err := f.read()
	if err != nil {
		return nil, err
	}
	req := m.GetLoginRequest()
	if m.GetType() != aprsis2.IS2Message_LOGIN_REQUEST || req == nil {
		return nil, fmt.Errorf("expected a login request, got %v", m)
	}

	m, err = f.read()
	if err != nil {
		return nil, err
	}
	ping := m.GetKeepalivePing()
	if ping.GetPingType() != aprsis2.IS2KeepalivePing_REPLY || ping.GetRequestId() != 42 || string(ping.GetRequestData()) != "abc" {
		return nil, fmt.Errorf("expected a ping reply, got %v", m)
	}

	reply := &aprsis2.IS2LoginReply{Result: result, Verified: verified}
	if result != aprsis2.IS2LoginReply_OK {
		reply.ResultCode = aprsis2.IS2LoginReply_INVALID_USERNAME
		reply.ResultMessage = "Invalid username"
	}
	f.send(&aprsis2.IS2Message{Type: aprsis2.IS2Message_LOGIN_REPLY, LoginReply: reply})

	return req, nil
}

// fakeIS2Server accepts one connection and runs fn on it. The error fn
// returns is reported by the returned function, which waits for fn.
func fakeIS2Server(t *testing.T, fn func(f *fakeIS2) error) (string, func()) {
	t.Helper()

	done := make(chan error, 1)
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		done <- fn(&fakeIS2{conn: conn, r: r})
	})

	return addr, func() {
		t.Helper()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("server side: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("server side did not finish")
		}
	}
}

var testOptions = ConnOptions{Callsign: "N0CALL", Passcode: "13023", AppName: "gotest", AppVer: "1.0"}

const testPacket = "N0CALL-1>APRS,qAR,N0CALL:>status"

func TestIS2Direct(t *testing.T) {
	addr, wait := fakeIS2Server(t, func(f *fakeIS2) error {
		req, err := f.login(aprsis2.IS2LoginReply_OK, aprsis2.VerificationStatus_WEAK)
		if err != nil {
			return err
		}
		if req.GetUsername() != "N0CALL" || req.GetPassword() != "13023" ||
			req.GetAppName() != "gotest" || req.GetAppVersion() != "1.0" {
			return fmt.Errorf("unexpected login request: %v", req)
		}

		// two packets in a frame, a ping, and one with metadata
		f.sendPackets(
			&aprsis2.ISPacket{Type: aprsis2.ISPacket_IS_PACKET, IsPacketData: []byte(testPacket + "1")},
			&aprsis2.ISPacket{Type: aprsis2.ISPacket_IS_PACKET, IsPacketData: []byte(testPacket + "2")})
		f.send(&aprsis2.IS2Message{
			Type:          aprsis2.IS2Message_KEEPALIVE_PING,
			KeepalivePing: &aprsis2.IS2KeepalivePing{PingType: aprsis2.IS2KeepalivePing_REQUEST, RequestId: 7},
		})
		f.sendPackets(&aprsis2.ISPacket{
			Type:              aprsis2.ISPacket_IS_PACKET,
			IsPacketData:      []byte(testPacket + "3"),
			OptionalRxRssiDbm: &aprsis2.ISPacket_RxRssiDbm{RxRssiDbm: -97.5},
			Received:          []*aprsis2.IS2Received{{ServerId: "T2TEST", Username: "N0CALL"}},
		})

		if m, err := f.read(); err != nil || m.GetKeepalivePing().GetRequestId() != 7 ||
			m.GetKeepalivePing().GetPingType() != aprsis2.IS2KeepalivePing_REPLY {
			return fmt.Errorf("expected a ping reply, got %v, %v", m, err)
		}

		// what the client sends: a packet, nothing for a comment, a
		// packet with metadata
		m, err := f.read()
		if err != nil {
			return err
		}
		if ps := m.GetIsPacket(); m.GetType() != aprsis2.IS2Message_IS_PACKET || len(ps) != 1 ||
			ps[0].GetType() != aprsis2.ISPacket_IS_PACKET || string(ps[0].GetIsPacketData()) != testPacket {
			return fmt.Errorf("unexpected packet message: %v", m)
		}
		m, err = f.read()
		if err != nil {
			return err
		}
		if ps := m.GetIsPacket(); len(ps) != 1 || ps[0].GetType() != aprsis2.ISPacket_IS_PACKET ||
			ps[0].GetRxSnrDb() != 5.25 || string(ps[0].GetIsPacketData()) != testPacket+"m" {
			return fmt.Errorf("unexpected packet message with metadata: %v", m)
		}
		return nil
	})

	var messages []aprsis2.IS2Message_Type
	o := testOptions
	o.IS2 = true
	o.OnIS2Message = func(m *aprsis2.IS2Message) { messages = append(messages, m.GetType()) }
	c, err := Connect(addr, o)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	if !c.IS2() {
		t.Error("IS2() is false on an IS2 connection")
	}
	if s := c.ServerSignature(); s.GetUsername() != "T2TEST" || s.GetAppName() != "fake" {
		t.Errorf("ServerSignature = %v", s)
	}
	if r := c.LoginReply(); r.GetVerified() != aprsis2.VerificationStatus_WEAK {
		t.Errorf("LoginReply = %v", r)
	}

	// ReadPacket and ReadLine both return packets
	if got, err := c.ReadPacket(2 * time.Second); err != nil || got != testPacket+"1" {
		t.Errorf("ReadPacket = %q, %v", got, err)
	}
	if got, err := c.ReadLine(2 * time.Second); err != nil || got != testPacket+"2" {
		t.Errorf("ReadLine = %q, %v", got, err)
	}
	p, err := c.ReadISPacket(2 * time.Second)
	if err != nil {
		t.Fatalf("ReadISPacket failed: %v", err)
	}
	if string(p.GetIsPacketData()) != testPacket+"3" || p.GetRxRssiDbm() != -97.5 ||
		len(p.GetReceived()) != 1 || p.GetReceived()[0].GetServerId() != "T2TEST" {
		t.Errorf("ReadISPacket = %v", p)
	}

	if err := c.SendLine(testPacket); err != nil {
		t.Errorf("SendLine: %v", err)
	}
	if err := c.SendLine("# a keepalive comment"); err != nil {
		t.Errorf("SendLine of a comment: %v", err)
	}
	// the type of the packet is filled in
	err = c.SendISPacket(&aprsis2.ISPacket{
		IsPacketData:    []byte(testPacket + "m"),
		OptionalRxSnrDb: &aprsis2.ISPacket_RxSnrDb{RxSnrDb: 5.25},
	})
	if err != nil {
		t.Errorf("SendISPacket: %v", err)
	}
	wait()

	want := []aprsis2.IS2Message_Type{
		aprsis2.IS2Message_SERVER_SIGNATURE, aprsis2.IS2Message_KEEPALIVE_PING,
		aprsis2.IS2Message_LOGIN_REPLY, aprsis2.IS2Message_KEEPALIVE_PING,
	}
	if fmt.Sprint(messages) != fmt.Sprint(want) {
		t.Errorf("OnIS2Message got %v, want %v", messages, want)
	}

	// the server has closed the connection
	if _, err := c.ReadPacket(2 * time.Second); !errors.Is(err, io.EOF) {
		t.Errorf("ReadPacket after close = %v, want EOF", err)
	}
}

func TestIS2Upgrade(t *testing.T) {
	addr, wait := fakeIS2Server(t, func(f *fakeIS2) error {
		fmt.Fprintf(f.conn, "# aprsc 2.2.0 test banner\r\n")
		if l, _ := f.r.ReadString('\n'); l != "#is2 hello v=1\r\n" {
			return fmt.Errorf("expected the hello, got %q", l)
		}
		// a comment before the answer is skipped; unknown keys of the
		// answer are ignored
		fmt.Fprintf(f.conn, "# aprsc keepalive\r\n#is2 hello future=x v=1\r\n")
		// the login line is discarded by the server
		if l, _ := f.r.ReadString('\n'); !strings.HasPrefix(l, "user N0CALL pass 13023 vers gotest 1.0") {
			return fmt.Errorf("expected the login line, got %q", l)
		}

		if _, err := f.login(aprsis2.IS2LoginReply_OK, aprsis2.VerificationStatus_WEAK); err != nil {
			return err
		}

		// the filter of the login is set with a parameter request; a
		// packet coming before the reply must not be lost
		m, err := f.read()
		if err != nil {
			return err
		}
		par := m.GetParameter()
		if par.GetType() != aprsis2.IS2Parameter_PARAMETER_SET || par.GetFilterString() != "r/60/25/100" {
			return fmt.Errorf("expected a filter request, got %v", m)
		}
		f.sendPackets(&aprsis2.ISPacket{Type: aprsis2.ISPacket_IS_PACKET, IsPacketData: []byte(testPacket)})
		f.send(&aprsis2.IS2Message{
			Type: aprsis2.IS2Message_PARAMETER,
			Parameter: &aprsis2.IS2Parameter{
				Type: aprsis2.IS2Parameter_PARAMETER_APPLIED, RequestId: par.GetRequestId(),
				FilterString: par.GetFilterString(),
			},
		})

		// SetFilter, refused, and the plaintext filter command
		m, err = f.read()
		if err != nil {
			return err
		}
		if m.GetParameter().GetFilterString() != "t/m" {
			return fmt.Errorf("expected a second filter request, got %v", m)
		}
		f.send(&aprsis2.IS2Message{
			Type: aprsis2.IS2Message_PARAMETER,
			Parameter: &aprsis2.IS2Parameter{
				Type: aprsis2.IS2Parameter_PARAMETER_FAILED, RequestId: m.GetParameter().GetRequestId(),
			},
		})
		m, err = f.read()
		if err != nil {
			return err
		}
		if m.GetParameter().GetFilterString() != "p/OH" {
			return fmt.Errorf("expected a third filter request, got %v", m)
		}
		return nil
	})

	o := testOptions
	o.Filter = "r/60/25/100"
	o.NoLogin = true
	c, err := Connect(addr, o)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	lr, err := c.Login()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	want := LogResp{Callsign: "N0CALL", Verified: true, Server: "T2TEST", IS2: true}
	if *lr != want {
		t.Errorf("LogResp = %+v, want %+v", *lr, want)
	}
	if !c.IS2() {
		t.Error("IS2() is false after an upgrade")
	}
	if c.Banner() != "# aprsc 2.2.0 test banner" {
		t.Errorf("Banner = %q", c.Banner())
	}

	if err := c.SetFilter("t/m"); !errors.Is(err, ErrFilterRejected) {
		t.Errorf("SetFilter = %v, want ErrFilterRejected", err)
	}
	if err := c.SendLine("#filter p/OH"); err != nil {
		t.Errorf("SendLine of a filter command: %v", err)
	}
	wait()

	if got, err := c.ReadPacket(2 * time.Second); err != nil || got != testPacket {
		t.Errorf("ReadPacket = %q, %v", got, err)
	}
}

// A server which does not know of IS2 ignores the hello, and the connection
// stays on plaintext.
func TestIS2UpgradeNotOffered(t *testing.T) {
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		r.ReadString('\n')
		r.ReadString('\n')
		fmt.Fprintf(conn, "# logresp N0CALL verified, server T2TEST\r\n%s\r\n", testPacket)
		// a packet with metadata arrives as two lines
		l1, _ := r.ReadString('\n')
		l2, _ := r.ReadString('\n')
		fmt.Fprintf(conn, "%s%s", l1, l2)
	})

	c, err := Connect(addr, testOptions)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	if c.IS2() || c.ServerSignature() != nil || c.LoginReply() != nil {
		t.Error("connection claims to be IS2")
	}
	p, err := c.ReadISPacket(2 * time.Second)
	if err != nil || string(p.GetIsPacketData()) != testPacket || p.GetType() != aprsis2.ISPacket_IS_PACKET {
		t.Errorf("ReadISPacket = %v, %v", p, err)
	}

	err = c.SendISPacket(&aprsis2.ISPacket{
		IsPacketData:        []byte(testPacket),
		OptionalRxRssiDbm:   &aprsis2.ISPacket_RxRssiDbm{RxRssiDbm: -97.5},
		OptionalRxSnrDb:     &aprsis2.ISPacket_RxSnrDb{RxSnrDb: 6},
		RxModem:             aprsis2.ISPacket_MODEM_LORA_APRS,
		OptionalRxBitrate:   &aprsis2.ISPacket_RxBitrate{RxBitrate: 300},
		OptionalRxFrequency: &aprsis2.ISPacket_RxFrequency{RxFrequency: 433775000},
	})
	if err != nil {
		t.Fatalf("SendISPacket: %v", err)
	}
	meta := "#is2-meta rx_rssi_dbm=-97.5 rx_snr_db=6 rx_modem=2 rx_bitrate=300 rx_frequency=433775000"
	if l, err := c.ReadLine(2 * time.Second); err != nil || l != meta {
		t.Errorf("metadata line = %q, %v, want %q", l, err, meta)
	}
	if l, err := c.ReadLine(2 * time.Second); err != nil || l != testPacket {
		t.Errorf("packet line = %q, %v", l, err)
	}
}

func TestIS2UpgradeDisabled(t *testing.T) {
	got := make(chan string, 1)
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		l, _ := r.ReadString('\n')
		got <- l
		fmt.Fprintf(conn, "# logresp N0CALL verified, server T2TEST\r\n")
	})

	o := testOptions
	o.NoIS2Upgrade = true
	c, err := Connect(addr, o)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	if l := <-got; l != "user N0CALL pass 13023 vers gotest 1.0\r\n" {
		t.Errorf("first line sent was %q", l)
	}
}

func TestIS2UpgradeBadVersion(t *testing.T) {
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		r.ReadString('\n')
		fmt.Fprintf(conn, "#is2 hello v=2\r\n")
		r.ReadString('\n')
	})

	if _, err := Connect(addr, testOptions); err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Errorf("Connect = %v, want an unsupported version error", err)
	}
}

func TestIS2LoginRejected(t *testing.T) {
	addr, wait := fakeIS2Server(t, func(f *fakeIS2) error {
		_, err := f.login(aprsis2.IS2LoginReply_FAIL, aprsis2.VerificationStatus_NONE)
		return err
	})

	o := testOptions
	o.IS2 = true
	_, err := Connect(addr, o)
	var le *IS2LoginError
	if !errors.As(err, &le) {
		t.Fatalf("Connect = %v, want an IS2LoginError", err)
	}
	if le.Reply.GetResultCode() != aprsis2.IS2LoginReply_INVALID_USERNAME {
		t.Errorf("result code = %v", le.Reply.GetResultCode())
	}
	wait()
}

func TestIS2LoginFilterRejected(t *testing.T) {
	addr, wait := fakeIS2Server(t, func(f *fakeIS2) error {
		if _, err := f.login(aprsis2.IS2LoginReply_OK, aprsis2.VerificationStatus_NONE); err != nil {
			return err
		}
		m, err := f.read()
		if err != nil {
			return err
		}
		f.send(&aprsis2.IS2Message{
			Type: aprsis2.IS2Message_PARAMETER,
			Parameter: &aprsis2.IS2Parameter{
				Type: aprsis2.IS2Parameter_PARAMETER_FAILED, RequestId: m.GetParameter().GetRequestId(),
			},
		})
		return nil
	})

	o := testOptions
	o.IS2 = true
	o.Filter = "t/m"
	o.NoLogin = true
	c, err := Connect(addr, o)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	// not a login failure, as on plaintext
	lr, err := c.Login()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if !lr.FilterRejected || lr.Verified || !lr.IS2 {
		t.Errorf("LogResp = %+v", *lr)
	}
	wait()
}

// A frame arriving in pieces is not lost when a read times out in the middle
// of it, and garbage is a framing error.
func TestIS2Framing(t *testing.T) {
	step := make(chan struct{})
	addr, wait := fakeIS2Server(t, func(f *fakeIS2) error {
		if _, err := f.login(aprsis2.IS2LoginReply_OK, aprsis2.VerificationStatus_WEAK); err != nil {
			return err
		}
		b, _ := proto.Marshal(&aprsis2.IS2Message{
			Type:     aprsis2.IS2Message_IS_PACKET,
			IsPacket: []*aprsis2.ISPacket{{Type: aprsis2.ISPacket_IS_PACKET, IsPacketData: []byte(testPacket)}},
		})
		frame := append(append([]byte{is2STX, 0, 0, byte(len(b))}, b...), is2ETX)
		f.conn.Write(frame[:10])
		<-step
		f.conn.Write(frame[10:])
		f.conn.Write([]byte("# not a frame\r\n"))
		return nil
	})

	o := testOptions
	o.IS2 = true
	c, err := Connect(addr, o)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	var ne net.Error
	if _, err := c.ReadPacket(100 * time.Millisecond); !errors.As(err, &ne) || !ne.Timeout() {
		t.Errorf("ReadPacket of a partial frame = %v, want a timeout", err)
	}
	close(step)
	if got, err := c.ReadPacket(2 * time.Second); err != nil || got != testPacket {
		t.Errorf("ReadPacket = %q, %v", got, err)
	}
	if _, err := c.ReadPacket(2 * time.Second); !errors.Is(err, ErrIS2Framing) {
		t.Errorf("ReadPacket of garbage = %v, want ErrIS2Framing", err)
	}
	wait()
}

func TestSendIS2MessagePlaintext(t *testing.T) {
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) { r.ReadString('\n') })

	o := testOptions
	o.NoLogin = true
	c, err := Connect(addr, o)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	if err := c.SendIS2Message(&aprsis2.IS2Message{}); err == nil {
		t.Error("SendIS2Message worked on a plaintext connection")
	}
}
