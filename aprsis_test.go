package fap

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()

	packets := []string{
		"N0CALL-1>APRS:;LEADER   *092345z4903.50N/07201.75W>088/036",
		"N0CALL-1>APRS:)AID #2!4903.50N/07201.75WA",
		"# server comment",
		"N0CALL-2>TQ4W2V,WIDE2-1,qAo,IGATE:`c51!f?>/]\"3x}=",
	}

	serverErr := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverErr <- fmt.Errorf("accept: %w", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			serverErr <- fmt.Errorf("read login: %w", err)
			return
		}
		login := string(buf[:n])
		if !strings.HasPrefix(login, "user N0CALL pass 13023") {
			serverErr <- fmt.Errorf("unexpected login: %q", login)
			return
		}

		fmt.Fprintf(conn, "# logresp N0CALL verified, server T2TEST\r\n")

		for _, pkt := range packets {
			fmt.Fprintf(conn, "%s\r\n", pkt)
		}

		serverErr <- nil
	}()

	c, err := Dial(addr, "N0CALL", "13023", "gotest", "1.0")
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c.Close()

	// ReadPacket should skip the server comment line
	for _, pkt := range packets {
		if strings.HasPrefix(pkt, "#") {
			continue
		}

		got, err := c.ReadPacket(2 * time.Second)
		if err != nil {
			t.Fatalf("ReadPacket failed: %v", err)
		}
		if got != pkt {
			t.Errorf("ReadPacket = %q, want %q", got, pkt)
		}

		p, err := Parse(got)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", got, err)
			continue
		}
		if p.SrcCallsign == "" {
			t.Errorf("Parse(%q): empty source callsign", got)
		}
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("server error: %v", err)
	}
}

func TestDialWithFilter(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()

	serverErr := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverErr <- fmt.Errorf("accept: %w", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			serverErr <- fmt.Errorf("read login: %w", err)
			return
		}
		login := string(buf[:n])
		if !strings.Contains(login, "filter r/60.0/25.0/100") {
			serverErr <- fmt.Errorf("filter not in login: %q", login)
			return
		}

		fmt.Fprintf(conn, "# logresp N0CALL verified, server T2TEST\r\n")
		serverErr <- nil
	}()

	c, err := Dial(addr, "N0CALL", "13023", "gotest", "1.0", "r/60.0/25.0/100")
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c.Close()

	if err := <-serverErr; err != nil {
		t.Fatalf("server error: %v", err)
	}
}

func TestAprsPasscode(t *testing.T) {
	tests := []struct {
		callsign string
		expected int16
	}{
		{"N0CALL", 13023},
		{"N1CALL", 13022},
		// With SSID - should produce same result as without.
		{"N0CALL-9", 13023},
		// Lowercase - should produce same result as uppercase.
		{"n0call", 13023},
	}

	for _, tc := range tests {
		t.Run(tc.callsign, func(t *testing.T) {
			got := AprsPasscode(tc.callsign)
			if got != tc.expected {
				t.Errorf("AprsPasscode(%q) = %d, want %d", tc.callsign, got, tc.expected)
			}
		})
	}
}
