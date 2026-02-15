package fap

import (
	"testing"
)

func TestWithAX25(t *testing.T) {
	tests := []struct {
		name       string
		packet     string
		wantErr    bool
		wantCode   string
		wantSrc    string
		wantDst    string
		wantDigis  []string
	}{
		{
			name:      "valid AX.25 packet",
			packet:    "OH7LZB-9>APX200,OH7AA-1*,WIDE2-1:!6028.51N/02505.68E#PHG7220/RELAY,WIDE, OH7LZB testing",
			wantSrc:   "OH7LZB-9",
			wantDst:   "APX200",
			wantDigis: []string{"OH7AA-1", "WIDE2-1"},
		},
		{
			name:      "valid AX.25 no digipeaters",
			packet:    "OH7LZB>APRS:>status",
			wantSrc:   "OH7LZB",
			wantDst:   "APRS",
			wantDigis: nil,
		},
		{
			name:     "invalid AX.25 source callsign",
			packet:   "TOOLONGCALLSIGN>APRS:>status",
			wantErr:  true,
			wantCode: ErrSrcCallNoAX25,
		},
		{
			name:     "source SSID too large for AX.25",
			packet:   "OH7LZB-16>APRS:>status",
			wantErr:  true,
			wantCode: ErrSrcCallNoAX25,
		},
		{
			name:     "invalid AX.25 digipeater callsign",
			packet:   "OH7LZB>APRS,TOOLONGCALLSIGN:>status",
			wantErr:  true,
			wantCode: ErrDigiCallNoAX25,
		},
		{
			name:     "too many path components for AX.25",
			packet:   "OH7LZB>APRS,D1,D2,D3,D4,D5,D6,D7,D8,D9:>status",
			wantErr:  true,
			wantCode: ErrDstPathTooMany,
		},
		{
			name:      "max 8 digipeaters is valid",
			packet:    "OH7LZB>APRS,D1,D2,D3,D4,D5,D6,D7,D8:>status",
			wantSrc:   "OH7LZB",
			wantDst:   "APRS",
			wantDigis: []string{"D1", "D2", "D3", "D4", "D5", "D6", "D7", "D8"},
		},
		{
			name:      "source callsign normalized to uppercase",
			packet:    "oh7lzb>APRS:>status",
			wantSrc:   "OH7LZB",
			wantDst:   "APRS",
			wantDigis: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.packet, WithAX25())
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error with code %q, got nil", tc.wantCode)
				}
				if p.ResultCode != tc.wantCode {
					t.Errorf("ResultCode = %q, want %q", p.ResultCode, tc.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.SrcCallsign != tc.wantSrc {
				t.Errorf("SrcCallsign = %q, want %q", p.SrcCallsign, tc.wantSrc)
			}
			if p.DstCallsign != tc.wantDst {
				t.Errorf("DstCallsign = %q, want %q", p.DstCallsign, tc.wantDst)
			}
			if tc.wantDigis == nil {
				if len(p.Digipeaters) != 0 {
					t.Errorf("expected no digipeaters, got %d", len(p.Digipeaters))
				}
			} else {
				if len(p.Digipeaters) != len(tc.wantDigis) {
					t.Fatalf("got %d digipeaters, want %d", len(p.Digipeaters), len(tc.wantDigis))
				}
				for i, want := range tc.wantDigis {
					if p.Digipeaters[i].Call != want {
						t.Errorf("Digipeaters[%d].Call = %q, want %q", i, p.Digipeaters[i].Call, want)
					}
				}
			}
		})
	}
}
