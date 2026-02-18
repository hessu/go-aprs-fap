package fap

// ParseError represents a parse failure with a machine-readable error code
// and a human-readable detail message. Sentinel errors have an empty Msg;
// per-instance errors returned by the parser carry a specific Msg.
//
// Use errors.Is(err, fap.ErrPosShort) to check for a specific error code.
type ParseError struct {
	Code string // Machine-readable error code (e.g. "pos_short")
	Msg  string // Human-readable detail
}

func (e *ParseError) Error() string {
	return "fap: " + e.Code + ": " + e.Msg
}

// Is reports whether target matches this error's Code, enabling errors.Is()
// to match any *ParseError with the same Code regardless of Msg.
func (e *ParseError) Is(target error) bool {
	t, ok := target.(*ParseError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Sentinel parse errors. Use errors.Is(err, fap.ErrXxx) to check.
var (
	ErrPacketNoBody     = &ParseError{Code: "packet_no_body"}
	ErrPacketShort      = &ParseError{Code: "packet_short"}
	ErrSrcCallNoGT      = &ParseError{Code: "srccall_nogt"}
	ErrSrcCallEmpty     = &ParseError{Code: "srccall_empty"}
	ErrSrcCallBadChars  = &ParseError{Code: "srccall_badchars"}
	ErrDstCallEmpty     = &ParseError{Code: "dstcall_empty"}
	ErrDstCallNoAX25    = &ParseError{Code: "dstcall_noax25"}
	ErrDstPathTooMany   = &ParseError{Code: "dstpath_toomany"}
	ErrSrcCallNoAX25    = &ParseError{Code: "srccall_noax25"}
	ErrDigiEmpty        = &ParseError{Code: "digi_empty"}
	ErrDigiCallBadChars = &ParseError{Code: "digicall_badchars"}
	ErrDigiCallNoAX25   = &ParseError{Code: "digicall_noax25"}
	ErrNoBody           = &ParseError{Code: "no_body"}
	ErrTypeNotSupported = &ParseError{Code: "type_not_supported"}
	ErrExpUnsupported   = &ParseError{Code: "exp_unsupp"}

	// Position errors
	ErrPosAmbiguity  = &ParseError{Code: "loc_amb_inv"}
	ErrPosShort      = &ParseError{Code: "pos_short"}
	ErrPosInvalid    = &ParseError{Code: "pos_invalid"}
	ErrPosLatInvalid = &ParseError{Code: "pos_lat_invalid"}
	ErrPosLonInvalid = &ParseError{Code: "pos_lon_invalid"}
	ErrLocInvalid    = &ParseError{Code: "loc_inv"}
	ErrLocLarge      = &ParseError{Code: "loc_large"}

	// Symbol errors
	ErrSymInvTable = &ParseError{Code: "sym_inv_table"}

	// Compressed position errors
	ErrCompShort   = &ParseError{Code: "comp_short"}
	ErrCompInvalid = &ParseError{Code: "comp_invalid"}

	// Mic-E errors
	ErrMiceShort        = &ParseError{Code: "mice_short"}
	ErrMiceInvDstCall   = &ParseError{Code: "mice_inv_dstcall"}
	ErrMiceInvInfoField = &ParseError{Code: "mice_inv_infofield"}

	// Object/item errors
	ErrObjShort    = &ParseError{Code: "obj_short"}
	ErrObjInvalid  = &ParseError{Code: "obj_inv"}
	ErrItemShort   = &ParseError{Code: "item_short"}
	ErrItemInvalid = &ParseError{Code: "item_invalid"}

	// Message errors
	ErrMsgShort      = &ParseError{Code: "msg_short"}
	ErrMsgInvalid    = &ParseError{Code: "msg_invalid"}
	ErrMsgNoDst      = &ParseError{Code: "msg_no_dst"}
	ErrMsgDstTooLong = &ParseError{Code: "msg_dst_long"}
	ErrMsgIDInvalid  = &ParseError{Code: "msg_id_inv"}
	ErrMsgReplyAck   = &ParseError{Code: "msg_replyack"}
	ErrMsgAckRej     = &ParseError{Code: "msg_ack_rej"}
	ErrMsgCRLF       = &ParseError{Code: "msg_cr"}

	// NMEA errors
	ErrNMEAShort   = &ParseError{Code: "nmea_short"}
	ErrNMEAInvalid = &ParseError{Code: "nmea_invalid"}
	ErrGPRMCNoFix  = &ParseError{Code: "gprmc_nofix"}

	// Timestamp errors
	ErrTimestampInvalid = &ParseError{Code: "timestamp_inv"}

	// Weather errors
	ErrWxInvalid = &ParseError{Code: "wx_invalid"}

	// Telemetry errors
	ErrTlmInvalid = &ParseError{Code: "tlm_inv"}
)
