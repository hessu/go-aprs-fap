package fap

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseErrorError(t *testing.T) {
	err := &ParseError{Code: "test_code", Msg: "test message"}
	want := "fap: test_code: test message"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestParseErrorIs(t *testing.T) {
	tests := []struct {
		name   string
		err    *ParseError
		target error
		want   bool
	}{
		{
			name:   "same code different msg",
			err:    &ParseError{Code: "obj_short", Msg: "object packet too short"},
			target: ErrObjShort,
			want:   true,
		},
		{
			name:   "different code",
			err:    &ParseError{Code: "obj_short", Msg: "object packet too short"},
			target: ErrObjInvalid,
			want:   false,
		},
		{
			name:   "non-ParseError target",
			err:    &ParseError{Code: "obj_short", Msg: "detail"},
			target: fmt.Errorf("some other error"),
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := errors.Is(tc.err, tc.target); got != tc.want {
				t.Errorf("errors.Is() = %v, want %v", got, tc.want)
			}
		})
	}
}
