package fap

// Error code constants for parse failures.
const (
	ErrPacketNoBody     = "packet_no_body"
	ErrPacketShort      = "packet_short"
	ErrSrcCallNoGT      = "srccall_nogt"
	ErrSrcCallEmpty     = "srccall_empty"
	ErrSrcCallBadChars  = "srccall_badchars"
	ErrDstCallEmpty     = "dstcall_empty"
	ErrDstCallNoAX25    = "dstcall_noax25"
	ErrDigiEmpty        = "digi_empty"
	ErrDigiCallBadChars = "digicall_badchars"
	ErrNoBody           = "no_body"
	ErrTypeNotSupported = "type_not_supported"
	ErrExpUnsupported   = "exp_unsupp"

	// Position errors
	ErrPosAmbiguity  = "pos_ambiguity"
	ErrPosShort      = "pos_short"
	ErrPosInvalid    = "pos_invalid"
	ErrPosLatInvalid = "pos_lat_invalid"
	ErrPosLonInvalid = "pos_lon_invalid"
	ErrLocInvalid    = "loc_inv"
	ErrLocLarge      = "loc_large"

	// Symbol errors
	ErrSymInvTable = "sym_inv_table"

	// Compressed position errors
	ErrCompShort   = "comp_short"
	ErrCompInvalid = "comp_invalid"

	// Mic-E errors
	ErrMiceShort        = "mice_short"
	ErrMiceInvDstCall   = "mice_inv_dstcall"
	ErrMiceInvInfoField = "mice_inv_infofield"

	// Object/item errors
	ErrObjShort    = "obj_short"
	ErrObjInvalid  = "obj_inv"
	ErrItemShort   = "item_short"
	ErrItemInvalid = "item_invalid"

	// Message errors
	ErrMsgShort   = "msg_short"
	ErrMsgInvalid = "msg_invalid"

	// NMEA errors
	ErrNMEAShort   = "nmea_short"
	ErrNMEAInvalid = "nmea_invalid"

	// Timestamp errors
	ErrTimestampInvalid = "timestamp_inv"

	// Weather errors
	ErrWxInvalid = "wx_invalid"

	// Telemetry errors
	ErrTlmInvalid = "tlm_inv"
)
