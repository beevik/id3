package id3

import "reflect"

// A Frame holds an entire ID3 frame including its header and payload.
type Frame struct {
	Header  FrameHeader
	Payload FramePayload
}

// A FrameHeader holds the data described by a frame header.
type FrameHeader struct {
	ID            string      // Frame ID string
	Size          int         // Frame size not including 10-byte header
	Flags         FrameFlags  // Flags
	GroupID       GroupSymbol // Optional group identifier
	EncryptMethod uint8       // Optional encryption method identifier
	DataLength    uint32      // Optional data length (if FrameFlagHasDataLength is set)
}

// A FramePayload describes the data held within a frame's payload.
type FramePayload interface {
}

// FrameFlags describe flags that may appear within a FrameHeader. Not all
// flags are supported by all versions of the ID3 codec.
type FrameFlags uint32

// All possible FrameFlags.
const (
	FrameFlagDiscardOnTagAlteration  FrameFlags = 1 << iota // Discard frame if tag is altered
	FrameFlagDiscardOnFileAlteration                        // Discard frame if file is altered
	FrameFlagReadOnly                                       // Frame is read-only
	FrameFlagHasGroupInfo                                   // Frame has group info
	FrameFlagCompressed                                     // Frame is compressed
	FrameFlagEncrypted                                      // Frame is encrypted
	FrameFlagUnsynchronized                                 // Frame is unsynchronized
	FrameFlagHasDataLength                                  // Frame has a data length indicator
)

// PictureType describes the type of picture stored within an APIC frame.
type PictureType uint8

// All possible values of the Type field within an APIC frame payload.
const (
	PictureTypeOther PictureType = iota
	PictureTypeIcon
	PictureTypeIconOther
	PictureTypeCoverFront
	PictureTypeCoverBack
	PictureTypeLeaflet
	PictureTypeMedia
	PictureTypeArtistLead
	PictureTypeArtist
	PictureTypeConductor
	PictureTypeBand
	PictureTypeComposer
	PictureTypeLyricist
	PictureTypeRecordingLocation
	PictureTypeDuringRecording
	PictureTypeDuringPerformance
	PictureTypeVideoCapture
	PictureTypeFish
	PictureTypeIlllustration
	PictureTypeBandLogotype
	PictureTypePublisherLogotype
)

// A GroupSymbol is a value between 0x80 and 0xF0 that uniquely identifies
// a grouped set of frames. The data associated with each GroupSymbol value
// is described futher in GRID frames.
type GroupSymbol byte

// frameTypes holds all possible frame payload types supported by ID3.
var frameTypes = []reflect.Type{
	reflect.TypeOf(FramePayloadUnknown{}),
	reflect.TypeOf(FramePayloadText{}),
	reflect.TypeOf(FramePayloadTXXX{}),
	reflect.TypeOf(FramePayloadCOMM{}),
	reflect.TypeOf(FramePayloadURL{}),
	reflect.TypeOf(FramePayloadWXXX{}),
	reflect.TypeOf(FramePayloadAPIC{}),
	reflect.TypeOf(FramePayloadUFID{}),
	reflect.TypeOf(FramePayloadUSER{}),
	reflect.TypeOf(FramePayloadUSLT{}),
	reflect.TypeOf(FramePayloadGRID{}),
}

type frameID uint16

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

// FramePayloadCOMM contains a full-text comment that doesn't fit in any
// of the other frames.
type FramePayloadCOMM struct {
	frameID     frameID `v22:"COM" v23:"COMM" v24:"COMM"`
	Encoding    Encoding
	Language    string `id3:"lang"`
	Description string
	Text        string
}

// FramePayloadURL may contain the payload of any type of URL frame except
// for the user-defined WXXX URL frame.
type FramePayloadURL struct {
	frameID frameID `v22:"W__" v23:"W___" v24:"W___"`
	URL     string  `id3:"iso88519"`
}

// FramePayloadWXXX contains a custom URL payload.
type FramePayloadWXXX struct {
	frameID     frameID `v22:"WXX" v23:"WXXX" v24:"WXXXX"`
	Encoding    Encoding
	Description string
	URL         string `id3:"iso88519"`
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
