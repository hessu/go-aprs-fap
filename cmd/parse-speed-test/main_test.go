package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParsePackets(t *testing.T) {
	input := strings.Join([]string{
		"1234567890 N0CALL-1>APRS:;LEADER   *092345z4903.50N/07201.75W>088/036",
		"1234567890 N0CALL-1>APRS:;LEADER  *000045z4903.50N/07201.75W>088/036",
		"1234567890 INVALID",
		"1234567890 # comment line",
	}, "\n") + "\n"

	var stdout, stderr bytes.Buffer
	code := run(nil, strings.NewReader(input), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Parsed 3 packets") {
		t.Errorf("expected 3 packets parsed (comment skipped):\n%s", output)
	}
	if !strings.Contains(output, "OK: 1") {
		t.Errorf("expected 1 OK packet:\n%s", output)
	}
	if !strings.Contains(output, "Failed: 2") {
		t.Errorf("expected 2 failed packet:\n%s", output)
	}
	if !strings.Contains(output, "Error summary") {
		t.Errorf("expected error summary:\n%s", output)
	}
}

func TestNoSpaceInLine(t *testing.T) {
	input := "nospace\n"

	var stdout, stderr bytes.Buffer
	code := run(nil, strings.NewReader(input), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Failed: 1") {
		t.Errorf("expected 1 failed packet:\n%s", output)
	}
	if !strings.Contains(output, "no space in line") {
		t.Errorf("expected 'no space in line' error:\n%s", output)
	}
}

func TestFilterError(t *testing.T) {
	input := "1234567890 INVALID\n"

	var stdout, stderr bytes.Buffer
	code := run([]string{"-e", "packet_no_body"}, strings.NewReader(input), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "INVALID") {
		t.Errorf("expected filtered packet in output:\n%s", output)
	}
}
