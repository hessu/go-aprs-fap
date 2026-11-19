package fap

import "strings"

// dstSymbol maps the symbol codes of the destination callsign symbols,
// GPSxy and SPCxy, to the symbol table and code. From the Perl module.
var dstSymbol = map[string][2]byte{
	"A0": {'\\', '0'},
	"A1": {'\\', '1'},
	"A2": {'\\', '2'},
	"A3": {'\\', '3'},
	"A4": {'\\', '4'},
	"A5": {'\\', '5'},
	"A6": {'\\', '6'},
	"A7": {'\\', '7'},
	"A8": {'\\', '8'},
	"A9": {'\\', '9'},
	"AA": {'\\', 'A'},
	"AB": {'\\', 'B'},
	"AC": {'\\', 'C'},
	"AD": {'\\', 'D'},
	"AE": {'\\', 'E'},
	"AF": {'\\', 'F'},
	"AG": {'\\', 'G'},
	"AH": {'\\', 'H'},
	"AI": {'\\', 'I'},
	"AJ": {'\\', 'J'},
	"AK": {'\\', 'K'},
	"AL": {'\\', 'L'},
	"AM": {'\\', 'M'},
	"AN": {'\\', 'N'},
	"AO": {'\\', 'O'},
	"AP": {'\\', 'P'},
	"AQ": {'\\', 'Q'},
	"AR": {'\\', 'R'},
	"AS": {'\\', 'S'},
	"AT": {'\\', 'T'},
	"AU": {'\\', 'U'},
	"AV": {'\\', 'V'},
	"AW": {'\\', 'W'},
	"AX": {'\\', 'X'},
	"AY": {'\\', 'Y'},
	"AZ": {'\\', 'Z'},
	"BB": {'/', '!'},
	"BC": {'/', '"'},
	"BD": {'/', '#'},
	"BE": {'/', '$'},
	"BF": {'/', '%'},
	"BG": {'/', '&'},
	"BH": {'/', '\''},
	"BI": {'/', '('},
	"BJ": {'/', ')'},
	"BK": {'/', '*'},
	"BL": {'/', '+'},
	"BM": {'/', ','},
	"BN": {'/', '-'},
	"BO": {'/', '.'},
	"BP": {'/', '/'},
	"DS": {'\\', '['},
	"DT": {'\\', '\\'},
	"DU": {'\\', ']'},
	"DV": {'\\', '^'},
	"DW": {'\\', '_'},
	"DX": {'\\', '`'},
	"HS": {'/', '['},
	"HT": {'/', '\\'},
	"HU": {'/', ']'},
	"HV": {'/', '^'},
	"HW": {'/', '_'},
	"HX": {'/', '`'},
	"J1": {'/', '{'},
	"J2": {'/', '|'},
	"J3": {'/', '}'},
	"J4": {'/', '~'},
	"LA": {'/', 'a'},
	"LB": {'/', 'b'},
	"LC": {'/', 'c'},
	"LD": {'/', 'd'},
	"LE": {'/', 'e'},
	"LF": {'/', 'f'},
	"LG": {'/', 'g'},
	"LH": {'/', 'h'},
	"LI": {'/', 'i'},
	"LJ": {'/', 'j'},
	"LK": {'/', 'k'},
	"LL": {'/', 'l'},
	"LM": {'/', 'm'},
	"LN": {'/', 'n'},
	"LO": {'/', 'o'},
	"LP": {'/', 'p'},
	"LQ": {'/', 'q'},
	"LR": {'/', 'r'},
	"LS": {'/', 's'},
	"LT": {'/', 't'},
	"LU": {'/', 'u'},
	"LV": {'/', 'v'},
	"LW": {'/', 'w'},
	"LX": {'/', 'x'},
	"LY": {'/', 'y'},
	"LZ": {'/', 'z'},
	"MR": {'/', ':'},
	"MS": {'/', ';'},
	"MT": {'/', '<'},
	"MU": {'/', '='},
	"MV": {'/', '>'},
	"MW": {'/', '?'},
	"MX": {'/', '@'},
	"NR": {'\\', ':'},
	"NS": {'\\', ';'},
	"NT": {'\\', '<'},
	"NU": {'\\', '='},
	"NV": {'\\', '>'},
	"NW": {'\\', '?'},
	"NX": {'\\', '@'},
	"OB": {'\\', '!'},
	"OC": {'\\', '"'},
	"OD": {'\\', '#'},
	"OE": {'\\', '$'},
	"OF": {'\\', '%'},
	"OG": {'\\', '&'},
	"OH": {'\\', '\''},
	"OI": {'\\', '('},
	"OJ": {'\\', ')'},
	"OK": {'\\', '*'},
	"OL": {'\\', '+'},
	"OM": {'\\', ','},
	"ON": {'\\', '-'},
	"OO": {'\\', '.'},
	"OP": {'\\', '/'},
	"P0": {'/', '0'},
	"P1": {'/', '1'},
	"P2": {'/', '2'},
	"P3": {'/', '3'},
	"P4": {'/', '4'},
	"P5": {'/', '5'},
	"P6": {'/', '6'},
	"P7": {'/', '7'},
	"P8": {'/', '8'},
	"P9": {'/', '9'},
	"PA": {'/', 'A'},
	"PB": {'/', 'B'},
	"PC": {'/', 'C'},
	"PD": {'/', 'D'},
	"PE": {'/', 'E'},
	"PF": {'/', 'F'},
	"PG": {'/', 'G'},
	"PH": {'/', 'H'},
	"PI": {'/', 'I'},
	"PJ": {'/', 'J'},
	"PK": {'/', 'K'},
	"PL": {'/', 'L'},
	"PM": {'/', 'M'},
	"PN": {'/', 'N'},
	"PO": {'/', 'O'},
	"PP": {'/', 'P'},
	"PQ": {'/', 'Q'},
	"PR": {'/', 'R'},
	"PS": {'/', 'S'},
	"PT": {'/', 'T'},
	"PU": {'/', 'U'},
	"PV": {'/', 'V'},
	"PW": {'/', 'W'},
	"PX": {'/', 'X'},
	"PY": {'/', 'Y'},
	"PZ": {'/', 'Z'},
	"Q1": {'\\', '{'},
	"Q2": {'\\', '|'},
	"Q3": {'\\', '}'},
	"Q4": {'\\', '~'},
	"SA": {'\\', 'a'},
	"SB": {'\\', 'b'},
	"SC": {'\\', 'c'},
	"SD": {'\\', 'd'},
	"SE": {'\\', 'e'},
	"SF": {'\\', 'f'},
	"SG": {'\\', 'g'},
	"SH": {'\\', 'h'},
	"SI": {'\\', 'i'},
	"SJ": {'\\', 'j'},
	"SK": {'\\', 'k'},
	"SL": {'\\', 'l'},
	"SM": {'\\', 'm'},
	"SN": {'\\', 'n'},
	"SO": {'\\', 'o'},
	"SP": {'\\', 'p'},
	"SQ": {'\\', 'q'},
	"SR": {'\\', 'r'},
	"SS": {'\\', 's'},
	"ST": {'\\', 't'},
	"SU": {'\\', 'u'},
	"SV": {'\\', 'v'},
	"SW": {'\\', 'w'},
	"SX": {'\\', 'x'},
	"SY": {'\\', 'y'},
	"SZ": {'\\', 'z'},
}

// symbolFromDstCall returns the symbol encoded in a destination callsign
// GPSxyz or SPCxyz, which is used by NMEA packets which have no symbol of
// their own. A port of the Perl module's _get_symbol_fromdst:
//   - GPSCnn and GPSEnn: symbol code nn+32 in the primary (C) or the
//     secondary (E) table, nn 01-94
//   - GPSxyO: secondary table code xy with the overlay O, when x is one
//     of O, A, N, D, S and Q, and O is A-Z or 0-9
//   - GPSxy: the symbol xy of the primary or the secondary table
func symbolFromDstCall(dst string) (table, code byte, ok bool) {
	if !strings.HasPrefix(dst, "GPS") && !strings.HasPrefix(dst, "SPC") {
		return 0, 0, false
	}
	rest := dst[3:]
	n := 0
	for n < len(rest) && n < 3 && isUpperAlphaNum(rest[n]) {
		n++
	}
	if n < 2 {
		return 0, 0, false
	}
	rest = rest[:n]

	if n == 2 {
		if s, found := dstSymbol[rest]; found {
			return s[0], s[1], true
		}
		return 0, 0, false
	}

	switch typ := rest[0]; typ {
	case 'C', 'E':
		if rest[1] < '0' || rest[1] > '9' || rest[2] < '0' || rest[2] > '9' {
			return 0, 0, false
		}
		num := int(rest[1]-'0')*10 + int(rest[2]-'0')
		if num < 1 || num > 94 {
			return 0, 0, false
		}
		table = '/'
		if typ == 'E' {
			table = '\\'
		}
		return table, byte(num + 32), true
	case 'O', 'A', 'N', 'D', 'S', 'Q':
		if s, found := dstSymbol[rest[:2]]; found {
			return rest[2], s[1], true
		}
	}
	return 0, 0, false
}
