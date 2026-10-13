package fap

import (
	"testing"
)

// Base-91 comment telemetry in uncompressed and compressed positions.
// Example packets from the specification, aprs-base91-comment-telemetry.txt
// section 3.11, and from perl-aprs-fap/t/21decode-uncomp-moving.t and
// t/22decode-compressed.t.

type commentTlmCase struct {
	name    string
	packet  string
	comment string
	seq     int
	vals    []*float64 // 5 channels, nil for undefined
	bits    string
}

func checkCommentTelemetry(t *testing.T, tc commentTlmCase) {
	p, err := Parse(tc.packet)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if p.Type != PacketTypeLocation {
		t.Errorf("type = %v, want location", p.Type)
	}
	if p.Comment != tc.comment {
		t.Errorf("comment = %q, want %q", p.Comment, tc.comment)
	}

	tlm := p.TelemetryData
	if tlm == nil {
		t.Fatal("no telemetry data")
	}
	if tlm.Seq != tc.seq {
		t.Errorf("seq = %d, want %d", tlm.Seq, tc.seq)
	}
	if len(tlm.Vals) != 5 {
		t.Fatalf("vals length = %d, want 5", len(tlm.Vals))
	}
	for i := range 5 {
		want := tc.vals[i]
		got := tlm.Vals[i]
		switch {
		case want == nil && got != nil:
			t.Errorf("vals[%d] = %v, want nil", i, *got)
		case want != nil && got == nil:
			t.Errorf("vals[%d] = nil, want %v", i, *want)
		case want != nil && got != nil && *want != *got:
			t.Errorf("vals[%d] = %v, want %v", i, *got, *want)
		}
	}
	if tlm.Bits != tc.bits {
		t.Errorf("bits = %q, want %q", tlm.Bits, tc.bits)
	}
}

func TestCommentTelemetry(t *testing.T) {
	cases := []commentTlmCase{
		{
			// spec 3.11.3: uncompressed position, PHG, comment text and 4 channels
			name:    "uncompressed PHG 4ch",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03NN02739.63E#PHG26303/Siilinjarvi|\"p%T'.agff|",
			comment: "Siilinjarvi",
			seq:     170,
			vals:    []*float64{new(415.0), new(559.0), new(5894.0), new(6348.0), nil},
		},
		{
			// spec 3.11.2: compressed position, comment text and 3 channels
			name:    "compressed 3ch",
			packet:  "N0CALL>APRS,qAR,IGATE:!/0%3RTh<6>dS_http://aprs.fi/|\"p%T'.ag|",
			comment: "http://aprs.fi/",
			seq:     170,
			vals:    []*float64{new(415.0), new(559.0), new(5894.0), nil, nil},
		},
		{
			// uncompressed position with course/speed, 3 channels, from a tracker
			name:    "uncompressed course speed 3ch",
			packet:  "N0CALL-14>APRS,qAR,IGATE:!6304.03N/02739.63E>036/010test |(7Cq?3<\"|",
			comment: "test",
			seq:     659,
			vals:    []*float64{new(3174.0), new(2748.0), new(2458.0), nil, nil},
		},
		{
			// t/21decode-uncomp-moving.t: altitude, 5 channels and bits, all zero
			name:    "uncompressed altitude 5ch bits",
			packet:  "OH7FDN>APZMDR,OH7AA-1*,WIDE2-1,qAR,OH7AA:!6253.52N/02739.47E>036/010/A=000465 |!!!!!!!!!!!!!!|",
			comment: "",
			seq:     0,
			vals:    []*float64{new(0.0), new(0.0), new(0.0), new(0.0), new(0.0)},
			bits:    "00000000",
		},
		{
			// t/22decode-compressed.t: comment text, 5 channels and bits, all zero
			name:    "compressed 5ch bits",
			packet:  "OH2LCQ-10>APZMDR,WIDE3-2,qAo,OH2MQK-1:!//zPHTfVv>!V_ Tero, Green Volvo 960, GGL-880|!!!!!!!!!!!!!!|",
			comment: "Tero, Green Volvo 960, GGL-880",
			seq:     0,
			vals:    []*float64{new(0.0), new(0.0), new(0.0), new(0.0), new(0.0)},
			bits:    "00000000",
		},
		{
			// spec 3.10: the bits '!"' decode to 1, B1 set, LSB first
			name:    "uncompressed bits LSB first",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>|ss1122334455!\"|",
			comment: "",
			seq:     7544,
			vals:    []*float64{new(1472.0), new(1564.0), new(1656.0), new(1748.0), new(1840.0)},
			bits:    "10000000",
		},
		{
			// a '|' in the free-form comment before the telemetry
			name:    "pipe in comment text",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>a|b |!!!!|",
			comment: "a|b",
			seq:     0,
			vals:    []*float64{new(0.0), nil, nil, nil, nil},
		},
		{
			// telemetry followed by a DAO extension, as the spec orders them
			name:    "uncompressed telemetry then DAO",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>comment|\"p%T|!W12!",
			comment: "comment",
			seq:     170,
			vals:    []*float64{new(415.0), nil, nil, nil, nil},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkCommentTelemetry(t, tc)
		})
	}
}

func TestCommentTelemetryInvalid(t *testing.T) {
	cases := []struct {
		name    string
		packet  string
		comment string
	}{
		{
			// 16 characters between the pipes: more than 5 channels and bits
			name:    "too long",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>c |!!!!!!!!!!!!!!!!|",
			comment: "c |!!!!!!!!!!!!!!!!|",
		},
		{
			// only a sequence number, no channel
			name:    "too short",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>c |!!|",
			comment: "c |!!|",
		},
		{
			name:    "odd length",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>c |!!!!!|",
			comment: "c |!!!!!|",
		},
		{
			// '}' and space are outside the base-91 range
			name:    "invalid character",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>c |!!!}|",
			comment: "c |!!!}|",
		},
		{
			name:    "single pipe",
			packet:  "N0CALL>APRS,qAR,IGATE:!6304.03N/02739.63E>c |!!!!",
			comment: "c |!!!!",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if p.TelemetryData != nil {
				t.Errorf("telemetry decoded from invalid extension: %+v", p.TelemetryData)
			}
			if p.Comment != tc.comment {
				t.Errorf("comment = %q, want %q", p.Comment, tc.comment)
			}
		})
	}
}
