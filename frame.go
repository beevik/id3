package id3

import "reflect"

// A Frame represents an ID3 tag frame's header and payload.
type Frame struct {
	Header  FrameHeader
	Payload FramePayload
}

// A FrameHeader holds data common to all ID3 frames.
type FrameHeader struct {
	ID            string      // Frame ID string
	Size          int         // Frame size not including 10-byte header
	Flags         uint8       // See FrameFlag*
	GroupID       GroupSymbol // Optional group identifier
	EncryptMethod uint8       // Optional encryption method identifier
	DataLength    uint32      // Optional data length (if FrameFlagHasDataLength is set)
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

// A GroupSymbol is a value between 0x80 and 0xF0 that uniquely identifies
// a grouped set of frames.
type GroupSymbol byte

// FramePayload represents the payload of an ID3 tag frame.
type FramePayload interface {
}

// frameTypes holds all possible frame payload types supported by ID3.
var frameTypes = []reflect.Type{
	reflect.TypeOf(FramePayloadUnknown{}),
	reflect.TypeOf(FramePayloadText{}),
	reflect.TypeOf(FramePayloadTXXX{}),
	reflect.TypeOf(FramePayloadAPIC{}),
	reflect.TypeOf(FramePayloadUFID{}),
	reflect.TypeOf(FramePayloadUSER{}),
	reflect.TypeOf(FramePayloadUSLT{}),
	reflect.TypeOf(FramePayloadGRID{}),
}

type frameID uint8

// FramePayloadUnknown contains the payload of any frame whose ID is
// unknown to this package.
type FramePayloadUnknown struct {
	frameID frameID `v23:"????" v24:"????"`
	Data    []byte
}

// FramePayloadText may contain the payload of any type of text frame
// except for a user-defined TXXX text frame.  In v2.4, each text frame
// may contain one or more text strings.  In all other versions, only one
// text string may appear.
type FramePayloadText struct {
	frameID  frameID `v22:"T__" v23:"T___" v24:"T___"`
	Encoding Encoding
	Text     []string
}

// FramePayloadTXXX contains a custom text payload.
type FramePayloadTXXX struct {
	frameID     frameID `v22:"TXX" v23:"TXXX" v24:"TXXX"`
	Encoding    Encoding
	Description string
	Text        string
}

// FramePayloadAPIC contains the payload of an image frame.
type FramePayloadAPIC struct {
	frameID     frameID `v22:"PIC" v23:"APIC" v24:"APIC"`
	Encoding    Encoding
	MimeType    string `id3:"iso88519"`
	Type        PictureType
	Description string
	Data        []byte
}

// FramePayloadUFID contains a unique file identifier for the MP3.
type FramePayloadUFID struct {
	frameID    frameID `v22:"UFI" v23:"UFID" v24:"UFID"`
	Owner      string  `id3:"iso88519"`
	Identifier string  `id3:"iso88519"`
}

// FramePayloadUSER contains the terms of use description for the MP3.
type FramePayloadUSER struct {
	frameID  frameID `v23:"USER" v24:"USER"`
	Encoding Encoding
	Language string `id3:"lang"`
	Text     string
}

// FramePayloadUSLT contains unsynchronized lyrics and text transcription
// data.
type FramePayloadUSLT struct {
	frameID    frameID `v22:"ULT" v23:"USLT" v24:"USLT"`
	Encoding   Encoding
	Language   string `id3:"lang"`
	Descriptor string
	Text       string
}

// FramePayloadGRID contains information describing the grouping of
// otherwise unrelated frames. If a frame contains an optional group
// identifier, there will be a corresponding GRID frame with data
// describing the group.
type FramePayloadGRID struct {
	frameID frameID `v23:"GRID" v24:"GRID"`
	Owner   string  `id3:"iso88519"`
	GroupID GroupSymbol
	Data    []byte
}
