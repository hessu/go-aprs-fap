package fap

import (
	"fmt"
	"testing"
)

// Tests ported from perl-aprs-fap/t/25decode-dao.t

func TestDAOUncompressedHumanReadable(t *testing.T) {
	// Uncompressed packet with human-readable DAO, DAO in beginning of comment
	packet := "K0ELR-15>APOT02,WIDE1-1,WIDE2-1,qAo,K0ELR:/102033h4133.03NX09029.49Wv204/000!W33! 12.3V 21C/A=000665"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse an uncompressed packet with WGS84 human-readable DAO: %v", err)
	}

	if p.DaoDatumByte != 'W' {
		t.Errorf("daodatumbyte = %q, want %q", p.DaoDatumByte, byte('W'))
	}
	if p.Comment != "12.3V 21C" {
		t.Errorf("comment = %q, want %q", p.Comment, "12.3V 21C")
	}
	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Latitude); got != "41.55055" {
		t.Errorf("latitude = %s, want 41.55055", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Longitude); got != "-90.49155" {
		t.Errorf("longitude = %s, want -90.49155", got)
	}
	if p.Altitude == nil {
		t.Fatal("altitude is nil")
	}
	if got := fmt.Sprintf("%.0f", *p.Altitude); got != "203" {
		t.Errorf("altitude = %s, want 203", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.3f", *p.PosResolution); got != "1.852" {
		t.Errorf("posresolution = %s, want 1.852", got)
	}
}

func TestDAOCompressedBase91(t *testing.T) {
	// Compressed packet with BASE91 DAO, DAO in end of comment
	packet := "OH7LZB-9>APZMDR,WIDE2-2,qAo,OH2RCH:!/0(yiTc5y>{2O http://aprs.fi/!w11!"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse a compressed packet with WGS84 BASE91 DAO: %v", err)
	}

	if p.DaoDatumByte != 'W' {
		t.Errorf("daodatumbyte = %q, want %q", p.DaoDatumByte, byte('W'))
	}
	if p.Comment != "http://aprs.fi/" {
		t.Errorf("comment = %q, want %q", p.Comment, "http://aprs.fi/")
	}
	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Latitude); got != "60.15273" {
		t.Errorf("latitude = %s, want 60.15273", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Longitude); got != "24.66222" {
		t.Errorf("longitude = %s, want 24.66222", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.PosResolution); got != "0.1852" {
		t.Errorf("posresolution = %s, want 0.1852", got)
	}
}

func TestDAOMicEBase91(t *testing.T) {
	// Mic-E packet with BASE91 DAO, DAO in middle of comment
	packet := "OH2JCQ-9>VP1U88,TRACE2-2,qAR,OH2RDK-5:'5'9\"^Rj/]\"4-}Foo !w66!Bar"
	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse a mic-e packet with WGS84 BASE91 DAO: %v", err)
	}

	if p.DaoDatumByte != 'W' {
		t.Errorf("daodatumbyte = %q, want %q", p.DaoDatumByte, byte('W'))
	}
	if p.Comment != "]Foo Bar" {
		t.Errorf("comment = %q, want %q", p.Comment, "]Foo Bar")
	}
	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Latitude); got != "60.26471" {
		t.Errorf("latitude = %s, want 60.26471", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.5f", *p.Longitude); got != "25.18821" {
		t.Errorf("longitude = %s, want 25.18821", got)
	}
	if p.PosResolution == nil {
		t.Fatal("posresolution is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.PosResolution); got != "0.1852" {
		t.Errorf("posresolution = %s, want 0.1852", got)
	}
}
