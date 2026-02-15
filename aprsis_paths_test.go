package fap

import (
	"testing"
)

func TestAPRSISPathIPv6AfterQConstruct(t *testing.T) {
	input := "IQ3VQ>APD225,TCPIP*,qAI,IQ3VQ,THIRD,92E5A2B6,T2HUB1,200106F8020204020000000000000002,T2FINLAND:!4526.66NI01104.68E#PHG21306/- Lnx APRS Srv - sez. ARI VR EST"
	p, err := Parse(input)
	if err != nil {
		t.Fatalf("failed to parse a packet with an IPv6 address in the path: %v", err)
	}
	if len(p.Digipeaters) != 8 {
		t.Fatalf("digipeaters count = %d, want at least 8", len(p.Digipeaters))
	}
	if p.Digipeaters[6].Call != "200106F8020204020000000000000002" {
		t.Errorf("digi[6].call = %q, want %q", p.Digipeaters[6].Call, "200106F8020204020000000000000002")
	}
}

func TestAPRSISPathIPv6BeforeQConstruct(t *testing.T) {
	input := "IQ3VQ>APD225,200106F8020204020000000000000002,TCPIP*,qAI,IQ3VQ,THIRD,92E5A2B6,T2HUB1,200106F8020204020000000000000002,T2FINLAND:!4526.66NI01104.68E#PHG21306/- Lnx APRS Srv - sez. ARI VR EST"
	_, err := Parse(input)
	if err == nil {
		t.Fatalf("managed to parse a packet with an IPv6 address in the path before qAI")
	}
}
