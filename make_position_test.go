package fap

import (
	"testing"
)

func TestMakePositionBasicNE(t *testing.T) {
	result, err := MakePosition(63.06716666666667, 27.6605, nil, nil, nil, "/#", nil)
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!6304.03N/02739.63E#"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionBasicSW(t *testing.T) {
	result, err := MakePosition(-23.64266666666667, -46.797, nil, nil, nil, "/#", nil)
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!2338.56S/04647.82W#"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionMinuteRounding(t *testing.T) {
	result, err := MakePosition(22.9999999, -177.9999999, nil, nil, nil, "/#", nil)
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!2259.99N/17759.99W#"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionWithSpeedCourseAlt(t *testing.T) {
	speed := 83.34
	course := 353.0
	alt := 95.7072
	result, err := MakePosition(52.364, 14.1045, &speed, &course, &alt, "/>", nil)
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!5221.84N/01406.27E>353/045/A=000314"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionWithAltOnly(t *testing.T) {
	alt := 95.7072
	result, err := MakePosition(52.364, 14.1045, nil, nil, &alt, "/>", nil)
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!5221.84N/01406.27E>/A=000314"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionAmbiguity1(t *testing.T) {
	result, err := MakePosition(52.364, 14.1045, nil, nil, nil, "/>", &MakePositionOpts{Ambiguity: 1})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!5221.8 N/01406.2 E>"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionAmbiguity2(t *testing.T) {
	result, err := MakePosition(52.364, 14.1045, nil, nil, nil, "/>", &MakePositionOpts{Ambiguity: 2})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!5221.  N/01406.  E>"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionAmbiguity3(t *testing.T) {
	result, err := MakePosition(52.364, 14.1045, nil, nil, nil, "/>", &MakePositionOpts{Ambiguity: 3})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!522 .  N/0140 .  E>"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionAmbiguity4(t *testing.T) {
	result, err := MakePosition(52.364, 14.1045, nil, nil, nil, "/>", &MakePositionOpts{Ambiguity: 4})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!52  .  N/014  .  E>"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionDAO_US(t *testing.T) {
	result, err := MakePosition(39.15380036630037, -84.62208058608059, nil, nil, nil, "/>", &MakePositionOpts{DAO: true})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!3909.22N/08437.32W>!wjM!"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionDAORounding(t *testing.T) {
	result, err := MakePosition(39.9999999, -84.9999999, nil, nil, nil, "/>", &MakePositionOpts{DAO: true})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!3959.99N/08459.99W>!w{{!"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}

func TestMakePositionDAOWithSpeedCourseAltComment(t *testing.T) {
	speed := 62.968
	course := 321.0
	alt := 192.9384
	result, err := MakePosition(48.37314835164835, 15.71477838827839, &speed, &course, &alt, "/>", &MakePositionOpts{DAO: true, Comment: "Comment blah"})
	if err != nil {
		t.Fatalf("MakePosition failed: %v", err)
	}
	expected := "!4822.38N/01542.88E>321/034/A=000633Comment blah!wr^!"
	if result != expected {
		t.Errorf("got %q, expected %q", result, expected)
	}
}
