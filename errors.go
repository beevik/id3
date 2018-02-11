package id3

import (
	"errors"
)

// Possible errors returned by this package.
var (
	ErrFailedCRC               = errors.New("tag failed CRC check")
	ErrIncompleteFrame         = errors.New("frame truncated prematurely")
	ErrInvalidBits             = errors.New("invalid bits value, should be 8 or 16")
	ErrInvalidBPM              = errors.New("invalid BPM value, must be less than 511")
	ErrInvalidEncodedString    = errors.New("invalid encoded string")
	ErrInvalidEncoding         = errors.New("invalid text encoding")
	ErrInvalidEncryptMethod    = errors.New("invalid encrypt method, must be between 0x80 and 0xf0")
	ErrInvalidFixedLenString   = errors.New("invalid fixed length string")
	ErrInvalidFooter           = errors.New("invalid footer")
	ErrInvalidFrame            = errors.New("invalid frame structure")
	ErrInvalidFrameFlags       = errors.New("invalid frame flags")
	ErrInvalidFrameHeader      = errors.New("invalid frame header")
	ErrInvalidGroupID          = errors.New("invalid group id, must be between 0x80 and 0xf0")
	ErrInvalidHeader           = errors.New("invalid tag header")
	ErrInvalidHeaderFlags      = errors.New("invalid header flags")
	ErrInvalidLyricContentType = errors.New("invalid lyric content type")
	ErrInvalidPictureType      = errors.New("invalid picture type")
	ErrInvalidSync             = errors.New("invalid sync code")
	ErrInvalidTag              = errors.New("invalid id3 tag")
	ErrInvalidText             = errors.New("invalid text string encountered")
	ErrInvalidTimeStampFormat  = errors.New("invalid time stamp format")
	ErrInvalidVersion          = errors.New("invalid id3 version")
	ErrUnknownFrameType        = errors.New("unknown frame type")

	errInsufficientBuffer = errors.New("insufficient buffer")
	errInvalidPayloadDef  = errors.New("invalid frame payload definition")
	errPaddingEncountered = errors.New("padding encountered")
	errUnimplemented      = errors.New("code path unimplemented")
	errUnknownFieldType   = errors.New("unknown field type")
)
