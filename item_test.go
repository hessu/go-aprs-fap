package fap

import (
	"errors"
	"fmt"
	"testing"
)

// Item tests based on APRS101 spec format:
// )ITEMNAME!position  (alive)
// )ITEMNAME_position  (killed)
// Item name is 3-9 characters, terminated by ! or _

func TestItemAlive(t *testing.T) {
	// Alive item with uncompressed position
	packet := "OH2KKU-1>APRS:)AID #2!4903.50N/07201.75WA"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse item: %v", err)
	}
	if p.Type != PacketTypeItem {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeItem)
	}
	if p.ItemName != "AID #2" {
		t.Errorf("itemname = %q, want %q", p.ItemName, "AID #2")
	}
	if p.Alive == nil || !*p.Alive {
		t.Error("expected alive = true")
	}

	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "49.0583" {
		t.Errorf("latitude = %s, want 49.0583", got)
	}
	if p.Longitude == nil {
		t.Fatal("longitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Longitude); got != "-72.0292" {
		t.Errorf("longitude = %s, want -72.0292", got)
	}

	if p.SymbolTable != '/' {
		t.Errorf("symboltable = %c, want /", p.SymbolTable)
	}
	if p.SymbolCode != 'A' {
		t.Errorf("symbolcode = %c, want A", p.SymbolCode)
	}
}

func TestItemKilled(t *testing.T) {
	// Killed item (underscore terminator)
	packet := "OH2KKU-1>APRS:)AID #2_4903.50N/07201.75WA"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse killed item: %v", err)
	}
	if p.Type != PacketTypeItem {
		t.Errorf("type = %q, want %q", p.Type, PacketTypeItem)
	}
	if p.ItemName != "AID #2" {
		t.Errorf("itemname = %q, want %q", p.ItemName, "AID #2")
	}
	if p.Alive == nil || *p.Alive {
		t.Error("expected alive = false")
	}
}

func TestItemShortName(t *testing.T) {
	// Item with minimum 3-char name
	packet := "OH2KKU-1>APRS:)X1Y!4903.50N/07201.75WA"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse short-name item: %v", err)
	}
	if p.ItemName != "X1Y" {
		t.Errorf("itemname = %q, want %q", p.ItemName, "X1Y")
	}
	if p.Alive == nil || !*p.Alive {
		t.Error("expected alive = true")
	}
}

func TestItemMaxLengthName(t *testing.T) {
	// Item with maximum 9-character name, terminator at index 9
	packet := "N0CALL-15>APRS,TCPIP*,qAC,T2TEST:)MyRadio99!4327.00N/00119.00WlMyRadio99"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse 9-char name item: %v", err)
	}
	if p.ItemName != "MyRadio99" {
		t.Errorf("itemname = %q, want %q", p.ItemName, "MyRadio99")
	}
	if p.Alive == nil || !*p.Alive {
		t.Error("expected alive = true")
	}
	if p.Latitude == nil {
		t.Fatal("latitude is nil")
	}
	if got := fmt.Sprintf("%.4f", *p.Latitude); got != "43.4500" {
		t.Errorf("latitude = %s, want 43.4500", got)
	}
}

func TestItemTooShort(t *testing.T) {
	// Packet too short for an item
	packet := "OH2KKU-1>APRS:)short"

	_, err := Parse(packet)
	if err == nil {
		t.Fatal("expected error for too-short item")
	}
	if !errors.Is(err, ErrItemShort) {
		t.Errorf("error = %v, want %v", err, ErrItemShort)
	}
}

func TestItemWithCourseSpeed(t *testing.T) {
	// Item with course/speed extension
	packet := "OH2KKU-1>APRS:)MOBILE!4903.50N/07201.75W>088/036"

	p, err := Parse(packet)
	if err != nil {
		t.Fatalf("failed to parse item with course/speed: %v", err)
	}
	if p.ItemName != "MOBILE" {
		t.Errorf("itemname = %q, want %q", p.ItemName, "MOBILE")
	}
	if p.Course == nil || *p.Course != 88 {
		t.Errorf("course = %v, want 88", p.Course)
	}
	if p.Speed == nil {
		t.Fatal("speed is nil")
	}
	// 36 knots * 1.852 = 66.672 km/h
	if got := fmt.Sprintf("%.3f", *p.Speed); got != "66.672" {
		t.Errorf("speed = %s, want 66.672", got)
	}
}
