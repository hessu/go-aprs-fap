package fap

// Tests ported from perl-aprs-fap/t/31decode-wx-ultw.t

import (
	"fmt"
	"testing"
)

func TestULTW(t *testing.T) {
	packet := "WC4PEM-14>APN391,WIDE2-1,qAo,K2KZ-3:$ULTW0053002D028D02FA2813000D87BD000103E8015703430010000C"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse an ULTW wx packet: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 64 {
		t.Errorf("wind_direction = %v, want 64", wx.WindDirection)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != "0.3" {
		t.Errorf("wind_speed = %s, want 0.3", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "2.3" {
		t.Errorf("wind_gust = %s, want 2.3", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "18.5" {
		t.Errorf("temp = %s, want 18.5", got)
	}
	if wx.Humidity == nil || *wx.Humidity != 100 {
		t.Errorf("humidity = %v, want 100", wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "1025.9" {
		t.Errorf("pressure = %s, want 1025.9", got)
	}

	if wx.Rain24h != nil {
		t.Errorf("rain_24h = %v, want nil", *wx.Rain24h)
	}
	if wx.Rain1h != nil {
		t.Errorf("rain_1h = %v, want nil", *wx.Rain1h)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "4.1" {
		t.Errorf("rain_midnight = %s, want 4.1", got)
	}

	if wx.Soft != "" {
		t.Errorf("soft = %q, want empty", wx.Soft)
	}
}

func TestULTWBelowZero(t *testing.T) {
	packet := "SR3DGT>APN391,SQ2LYH-14,SR4DOS,WIDE2*,qAo,SR4NWO-1:$ULTW00000000FFEA0000296F000A9663000103E80016025D"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse an ULTW wx packet: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 0 {
		t.Errorf("wind_direction = %v, want 0", wx.WindDirection)
	}
	if wx.WindSpeed != nil {
		t.Errorf("wind_speed = %v, want nil", *wx.WindSpeed)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindGust); got != "0.0" {
		t.Errorf("wind_gust = %s, want 0.0", got)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "-19.0" {
		t.Errorf("temp = %s, want -19.0", got)
	}
	if wx.Humidity == nil || *wx.Humidity != 100 {
		t.Errorf("humidity = %v, want 100", wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "1060.7" {
		t.Errorf("pressure = %s, want 1060.7", got)
	}

	if wx.Rain24h != nil {
		t.Errorf("rain_24h = %v, want nil", *wx.Rain24h)
	}
	if wx.Rain1h != nil {
		t.Errorf("rain_1h = %v, want nil", *wx.Rain1h)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "0.0" {
		t.Errorf("rain_midnight = %s, want 0.0", got)
	}

	if wx.Soft != "" {
		t.Errorf("soft = %q, want empty", wx.Soft)
	}
}

func TestULTWLogging(t *testing.T) {
	packet := "MB7DS>APRS,TCPIP*,qAC,APRSUK2:!!00000066013D000028710166--------0158053201200210"

	p, err := Parse(packet, nil)
	if err != nil {
		t.Fatalf("failed to parse an ULTW logging wx packet: %v", err)
	}

	wx := p.Wx
	if wx == nil {
		t.Fatalf("wx is nil")
	}

	if wx.WindDirection == nil || *wx.WindDirection != 144 {
		t.Errorf("wind_direction = %v, want 144", wx.WindDirection)
	}
	if got := fmt.Sprintf("%.1f", *wx.WindSpeed); got != "14.7" {
		t.Errorf("wind_speed = %s, want 14.7", got)
	}
	if wx.WindGust != nil {
		t.Errorf("wind_gust = %v, want nil", *wx.WindGust)
	}

	if got := fmt.Sprintf("%.1f", *wx.Temp); got != "-0.2" {
		t.Errorf("temp = %s, want -0.2", got)
	}
	if got := fmt.Sprintf("%.1f", *wx.TempIn); got != "2.1" {
		t.Errorf("temp_in = %s, want 2.1", got)
	}
	if wx.Humidity != nil {
		t.Errorf("humidity = %v, want nil", *wx.Humidity)
	}
	if got := fmt.Sprintf("%.1f", *wx.Pressure); got != "1035.3" {
		t.Errorf("pressure = %s, want 1035.3", got)
	}

	if wx.Rain24h != nil {
		t.Errorf("rain_24h = %v, want nil", *wx.Rain24h)
	}
	if wx.Rain1h != nil {
		t.Errorf("rain_1h = %v, want nil", *wx.Rain1h)
	}
	if got := fmt.Sprintf("%.1f", *wx.RainMidnight); got != "73.2" {
		t.Errorf("rain_midnight = %s, want 73.2", got)
	}

	if wx.Soft != "" {
		t.Errorf("soft = %q, want empty", wx.Soft)
	}
}
