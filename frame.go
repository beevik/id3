package id3

import "io"

// A Frame is a piece of an ID3 tag that contains information about the
// MP3 file.
type Frame interface {
	// ID returns the 3- or 4-character string representing the type of
	// frame.
	ID() string

	// ReadFrom reads the contents of a frame from an IO stream.
	ReadFrom(r io.Reader) (n int64, err error)

	// WriteTo writes the contents of a frame to an IO stream.
	WriteTo(w io.Writer) (n int64, err error)
}

// A FrameHeader holds data common to all ID3 frames.
type FrameHeader struct {
	TextID string
	Size   uint8
	Flags  uint8
}

func (h *FrameHeader) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}

func (h *FrameHeader) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

// Possible values of flags stored per frame.
const (
	FrameFlagDiscardOnTagAlteration  uint8 = 1 << 0
	FrameFlagDiscardOnFileAlteration       = 1 << 1
	FrameFlagReadOnly                      = 1 << 2
	FrameFlagHasGroupInfo                  = 1 << 3
	FrameFlagCompressed                    = 1 << 4
	FrameFlagEncrypted                     = 1 << 5
	FrameFlagUnsynchronized                = 1 << 6
	FrameFlagHasDataLength                 = 1 << 7
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
