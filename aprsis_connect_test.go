package fap

import (
	"bufio"
	"net"
	"testing"
	"time"
)

// fakeServer accepts one connection and runs fn on it.
func fakeServer(t *testing.T, fn func(conn net.Conn, r *bufio.Reader)) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		fn(conn, bufio.NewReader(conn))
	}()

	return ln.Addr().String()
}

func TestParseLogResp(t *testing.T) {
	tests := []struct {
		line string
		want LogResp
	}{
		{"# logresp N0CALL verified, server T2TEST",
			LogResp{Callsign: "N0CALL", Verified: true, Server: "T2TEST"}},
		{"# logresp N0CALL unverified, server T2TEST",
			LogResp{Callsign: "N0CALL", Verified: false, Server: "T2TEST"}},
		{"# logresp N0CALL unverified",
			LogResp{Callsign: "N0CALL"}},
		{"# logresp",
			LogResp{}},
		{"# logresp N0CALL verified, server",
			LogResp{Callsign: "N0CALL", Verified: true}},
	}

	for _, tt := range tests {
		got := parseLogResp(tt.line)
		tt.want.Raw = tt.line
		if *got != tt.want {
			t.Errorf("parseLogResp(%q) = %+v, want %+v", tt.line, *got, tt.want)
		}
	}
}

func TestConnectNoLoginCustomLine(t *testing.T) {
	gotLine := make(chan string, 1)
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		conn.Write([]byte("# aprsc 2.1.20 test banner\r\n"))
		line, _ := r.ReadString('\n')
		gotLine <- line
		conn.Write([]byte("# logresp N0CALL unverified, server T2TEST\r\n"))
	})

	var comments []string
	c, err := Connect(addr, ConnOptions{
		NoLogin:   true,
		LoginLine: "user N0CALL pass -1 vers x 1 garbage",
		OnComment: func(l string) { comments = append(comments, l) },
	})
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	lr, err := c.Login()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if lr.Verified || lr.Server != "T2TEST" {
		t.Errorf("unexpected LogResp %+v", lr)
	}

	if l := <-gotLine; l != "user N0CALL pass -1 vers x 1 garbage\r\n" {
		t.Errorf("server saw login %q", l)
	}
	if c.Banner() != "# aprsc 2.1.20 test banner" {
		t.Errorf("Banner = %q", c.Banner())
	}
	if len(comments) != 2 {
		t.Errorf("OnComment called %d times, want 2: %q", len(comments), comments)
	}
}

func TestLoginRejectedConnectionClosed(t *testing.T) {
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		r.ReadString('\n')
		conn.Write([]byte("# logresp N0CALL unverified, server T2TEST\r\n"))
	})

	c, err := Connect(addr, ConnOptions{Callsign: "N0CALL", Passcode: "-1", AppName: "t", AppVer: "1"})
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()
}

func TestLoginTimeout(t *testing.T) {
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		r.ReadString('\n')
		time.Sleep(time.Second) // never answers in time
	})

	start := time.Now()
	_, err := Connect(addr, ConnOptions{
		Callsign: "N0CALL", Passcode: "-1", AppName: "t", AppVer: "1",
		LoginTimeout: 100 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected a login error")
	}
	if time.Since(start) > 800*time.Millisecond {
		t.Errorf("Login did not honour LoginTimeout, took %v", time.Since(start))
	}
}

func TestSendRawAndCloseWrite(t *testing.T) {
	got := make(chan string, 1)
	addr := fakeServer(t, func(conn net.Conn, r *bufio.Reader) {
		buf := make([]byte, 0, 64)
		tmp := make([]byte, 64)
		for {
			n, err := conn.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if err != nil {
				break
			}
		}
		got <- string(buf)
		conn.Write([]byte("bye\r\n"))
	})

	c, err := Connect(addr, ConnOptions{NoLogin: true})
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	if err := c.SendRaw([]byte("partial")); err != nil {
		t.Fatalf("SendRaw: %v", err)
	}
	if err := c.CloseWrite(); err != nil {
		t.Fatalf("CloseWrite: %v", err)
	}
	if s := <-got; s != "partial" {
		t.Errorf("server received %q", s)
	}

	// the read side still works after the half-close
	if l, err := c.ReadLine(2 * time.Second); err != nil || l != "bye" {
		t.Errorf("ReadLine = %q, %v", l, err)
	}
}

func TestConnectDialer(t *testing.T) {
	called := false
	_, err := Connect("ignored:1", ConnOptions{
		NoLogin: true,
		Dialer: func(network, addr string) (net.Conn, error) {
			called = true
			a, b := net.Pipe()
			b.Close()
			return a, nil
		},
	})
	if err != nil || !called {
		t.Errorf("custom Dialer not used: called=%v err=%v", called, err)
	}
}
