package id3

import (
	"bytes"
)

// A Frame is a piece of an ID3 tag that contains information about the
// MP3 file.
type Frame interface {
	// ID returns the 3- or 4-character string representing the type of
	// frame.
	ID() string
}

// A FrameHeader holds data common to all ID3 frames.
type FrameHeader struct {
	IDvalue    string // 3 or 4 character ID string
	Size       uint32 // frame size not including header
	Flags      uint8  // See FrameFlag*
	GroupID    uint8  // Optional group identifier
	DataLength uint32 // Optional data length (if FrameFlagHasDataLength is set)
}

// Possible values of flags stored per frame.
const (
	FrameFlagDiscardOnTagAlteration  uint8 = 1 << 0 // Discard frame if tag is altered
	FrameFlagDiscardOnFileAlteration       = 1 << 1 // Discard frame if file is altered
	FrameFlagReadOnly                      = 1 << 2 // Frame is read-only
	FrameFlagHasGroupInfo                  = 1 << 3 // Frame has group info
	FrameFlagCompressed                    = 1 << 4 // Frame is compressed
	FrameFlagEncrypted                     = 1 << 5 // Frame is encrypted
	FrameFlagUnsynchronized                = 1 << 6 // Frame is unsynchronized
	FrameFlagHasDataLength                 = 1 << 7 // Frame has a data length indicator
)

// Encoding represents the type of encoding used on a text string with an
// ID3 frame.
type Encoding uint8

// Possible values used to indicate the type of text encoding.
const (
	EncodingISO88591 Encoding = 0
	EncodingUTF16BOM          = 1
	EncodingUTF16             = 2
	EncodingUTF8              = 3
)

// PictureType represents the type of picture stored in an APIC frame.
type PictureType uint8

// Possible values used to indicate an APIC frame's picture type.
const (
	PictureTypeOther             PictureType = 0
	PictureTypeIcon                          = 1
	PictureTypeIconOther                     = 2
	PictureTypeCoverFront                    = 3
	PictureTypeCoverBack                     = 4
	PictureTypeLeaflet                       = 5
	PictureTypeMedia                         = 6
	PictureTypeArtistLead                    = 7
	PictureTypeArtist                        = 8
	PictureTypeConductor                     = 9
	PictureTypeBand                          = 10
	PictureTypeComposer                      = 11
	PictureTypeLyricist                      = 12
	PictureTypeRecordingLocation             = 13
	PictureTypeDuringRecording               = 14
	PictureTypeDuringPerformance             = 15
	PictureTypeVideoCapture                  = 16
	PictureTypeFish                          = 17
	PictureTypeIlllustration                 = 18
	PictureTypeBandLogotype                  = 19
	PictureTypePublisherLogotype             = 20
)

// A codec used to encode/decode a particular type of frame.
type frameCodec interface {
	decode(buf *bytes.Buffer) (Frame, error)
	encode(frame Frame, buf *bytes.Buffer) error
}

//
// FrameText
//

type FrameText struct {
	FrameHeader
	Encoding Encoding
	Text     string
}

func (f *FrameText) ID() string {
	return f.FrameHeader.IDvalue
}

func NewFrameText(id string) *FrameText {
	return &FrameText{
		FrameHeader{id, 1, 0, 0, 0},
		EncodingUTF8,
		"",
	}
}

//
// FrameAPIC
//

type FrameAPIC struct {
	FrameHeader
	Encoding    Encoding
	MimeType    string
	Type        PictureType
	Description string
	Data        []byte
}

func (f *FrameAPIC) ID() string {
	return "APIC"
}
