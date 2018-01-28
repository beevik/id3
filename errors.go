package id3

import (
	"errors"
)

// Possible errors returned by this package.
var (
	ErrInvalidTag           = errors.New("invalid id3 tag")
	ErrInvalidVersion       = errors.New("invalid id3 version")
	ErrInvalidHeaderFlags   = errors.New("invalid header flags")
	ErrBadSync              = errors.New("invalid sync code")
	ErrBadEncoding          = errors.New("invalid encoding type")
	ErrBadText              = errors.New("invalid text string encountered")
	ErrIncompleteFrame      = errors.New("frame truncated prematurely")
	ErrUnknownFrameType     = errors.New("unknown frame type")
	ErrInvalidEncoding      = errors.New("invalid text encoding")
	ErrInvalidFrame         = errors.New("invalid frame structure")
	ErrInvalidFrameHeader   = errors.New("invalid frame header")
	ErrInvalidFrameFlags    = errors.New("invalid frame flags")
	ErrInvalidEncodedString = errors.New("invalid encoded string")
	ErrUnknownFieldType     = errors.New("unknown field type")
	ErrUnimplemented        = errors.New("code path unimplemented")

	errPaddingEncountered = errors.New("padding encountered")
	errInvalidPayloadDef  = errors.New("invalid frame payload definition")
	errInsufficientBuffer = errors.New("insufficient buffer")
)
