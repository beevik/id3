package id3

type FrameData interface {
}

type Frame struct {
	Header FrameHeader
	Data   FrameData
}

// A FrameHeader holds data common to all ID3 frames.
type FrameHeader struct {
	ID            string // Frame ID string
	Size          uint32 // Frame size not including 10-byte header
	Flags         uint8  // See FrameFlag*
	GroupID       uint8  // Optional group identifier
	EncryptMethod uint8  // Optional encryption method identifier
	DataLength    uint32 // Optional data length (if FrameFlagHasDataLength is set)
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
	decode(h *FrameHeader, buf []byte) (FrameData, error)
	encode(h *FrameHeader, d FrameData) ([]byte, error)
}

//
// FrameText
//

type FrameDataText struct {
	Text EncodedText
}

//
// FrameDataAPIC
//

type FrameDataAPIC struct {
	Encoding    Encoding
	MimeType    string
	Type        PictureType
	Description string
	Data        []byte
}
