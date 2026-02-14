package fap

// Error code constants for parse failures.
const (
	ErrPacketNoBody    = "packet_no_body"
	ErrPacketShort     = "packet_short"
	ErrSrcCallNoGT     = "srccall_nogt"
	ErrSrcCallEmpty    = "srccall_empty"
	ErrDstCallEmpty    = "dstcall_empty"
	ErrDigiEmpty       = "digi_empty"
	ErrNoBody          = "no_body"
	ErrTypeNotSupported = "type_not_supported"

	// Position errors
	ErrPosAmbiguity  = "pos_ambiguity"
	ErrPosShort      = "pos_short"
	ErrPosInvalid    = "pos_invalid"
	ErrPosLatInvalid = "pos_lat_invalid"
	ErrPosLonInvalid = "pos_lon_invalid"

	// Compressed position errors
	ErrCompShort   = "comp_short"
	ErrCompInvalid = "comp_invalid"

	// Mic-E errors
	ErrMiceShort          = "mice_short"
	ErrMiceInvDstCall     = "mice_inv_dstcall"
	ErrMiceInvInfoField   = "mice_inv_infofield"
	ErrMiceInvSymTable    = "sym_inv_table"

	// Object/item errors
	ErrObjShort   = "obj_short"
	ErrObjInvalid = "obj_invalid"
	ErrItemShort  = "item_short"
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
	ErrTlmInvalid = "tlm_invalid"
)
