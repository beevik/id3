package id3

import "reflect"

// A FrameHolder holds the header and payload of an ID3 frame.
type FrameHolder struct {
	header FrameHeader
	Frame  Frame
}

// NewFrameHolder creates a new frame holder, which contains the header
// and payload of an ID3 frame.
func NewFrameHolder(frame Frame) *FrameHolder {
	t := reflect.ValueOf(frame).Elem()
	return &FrameHolder{
		header: FrameHeader{ID: FrameID(t.Field(0).String())},
		Frame:  frame,
	}
}

// Size returns the encoded size of the frame, not including the header.
func (f *FrameHolder) Size() int {
	return f.header.Size
}

// ID returns the 4-character ID string currently assigned to the frame.
func (f *FrameHolder) ID() FrameID {
	return f.header.ID
}

// A FrameID is a 4-character string indicating the type of ID3 frame.
type FrameID string

// A FrameHeader holds the data described by a frame header.
type FrameHeader struct {
	ID            FrameID     // Frame ID string
	Size          int         // Frame size not including 10-byte header
	Flags         FrameFlags  // Flags
	GroupID       GroupSymbol // Optional group identifier
	EncryptMethod uint8       // Optional encryption method identifier
	DataLength    uint32      // Optional data length (if FrameFlagHasDataLength is set)
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

// A TextType value identifies the type of an ID3 text frame.
type TextType uint8

// All standard types of text frames.
const (
	// Identification (ID3v2.4 spec section 4.2.1)
	TextTypeGroupDescription TextType = iota // TIT1
	TextTypeSongTitle                        // TIT2
	TextTypeSongSubtitle                     // TIT3
	TextTypeAlbumName                        // TALB
	TextTypeOriginalAlbum                    // TOAL
	TextTypeTrackNumber                      // TRCK
	TextTypePartOfSet                        // TPOS
	TextTypeSetSubtitle                      // TSST (v2.4 only)
	TextTypeISRC                             // TSRC

	// Involved persons (ID3v2.4 spec section 4.2.2)
	TextTypeArtist            // TPE1
	TextTypeAlbumArtist       // TPE2
	TextTypeConductor         // TPE3
	TextTypeRemixer           // TPE4
	TextTypeOriginalPerformer // TOPE
	TextTypeLyricist          // TEXT
	TextTypeOriginalLyricist  // TOLY
	TextTypeComposer          // TCOM
	TextTypeMusicians         // TMCL (v2.4 only)
	TextTypeInvolvedPeople    // TIPL (v2.4 only)
	TextTypeEncodedBy         // TENC

	// Derived and subjective properties (ID3v2.4 spec section 4.2.3)
	TextTypeBPM        // TBPM
	TextTypeLengthInMs // TLEN
	TextTypeMusicalKey // TKEY
	TextTypeLanguage   // TLAN
	TextTypeGenre      // TCON (see Genre)
	TextTypeFileType   // TFLT (see FileType)
	TextTypeMediaType  // TMED
	TextTypeMood       // TMOO (v2.4 only)

	// Rights and license (ID3v2.4 spec section 4.2.4)
	TextTypeCopyright         // TCOP
	TextTypeProducedNotice    // TPRO (v2.4 only)
	TextTypePublisher         // TPUB
	TextTypeOwner             // TOWN
	TextTypeRadioStation      // TRSN
	TextTypeRadioStationOwner // TRSO

	// Other text frames (ID3v2.4 spec section 4.2.5)
	TextTypeOriginalFileName    // TOFN
	TextTypePlaylistDelay       // TDLY
	TextTypeEncodingTime        // TDEN (v2.4 only)
	TextTypeOriginalReleaseTime // TDOR (v2.4 only)
	TextTypeRecordingTime       // TDRC (v2.4 only)
	TextTypeReleaseTime         // TDRL (v2.4 only)
	TextTypeTaggingTime         // TDTG (v2.4 only)
	TextTypeEncodingSoftware    // TSSE
	TextTypeAlbumSortOrder      // TSOA (v2.4 only)
	TextTypeTitleSortOrder      // TSOT (v2.4 only)

	// v2.3-only frames (ID3v2.3 spec)
	TextTypeDate                // TDAT (TDRC in v2.4)
	TextTypeTime                // TIME (TDRC in v2.4)
	TextTypeOriginalReleaseYear // TORY (TDOR in v2.4)
	TextTypeRecordingDates      // TRDA (TDRC in v2.4)
	TextTypeYear                // TYER (TDRC in v2.4)
	TextTypeSize                // TSIZ

	// Non-standard values
	TextTypeUnknown
)

// TimeStampFormat indicates the type of time stamp used: milliseconds or
// MPEG frame.
type TimeStampFormat byte

// All possible values of the TimeStampFormat type.
const (
	TimeStampFrames TimeStampFormat = 1 + iota
	TimeStampMilliseconds
)

// LyricContentType indicates type type of lyrics stored in a synchronized
// lyric frame.
type LyricContentType byte

// All possible values of the LyricContentType type.
const (
	LyricContentTypeOther LyricContentType = iota
	LyricContentTypeLyrics
	LyricContentTypeTranscription
	LyricContentTypeMovement
	LyricContentTypeEvents
	LyricContentTypeChord
	LyricContentTypeTrivia
	LyricContentTypeWebURL
	LyricContentTypeImageURL
)

// A GroupSymbol is a value between 0x80 and 0xF0 that uniquely identifies
// a grouped set of frames. The data associated with each GroupSymbol value
// is described futher in GRID frames.
type GroupSymbol byte

// A Frame is an interface capable of representing the payload of any of the
// possible frame types (e.g., FrameText, FrameURL, etc.).
//
// Use a type assertion to access the frame's contents. For example:
//
//	for _, h := range tag.FrameHolders {
//		switch f := h.Frame.(type) {
// 			case *id3.FrameText:
// 				fmt.Printf("%v\n", f.Text)
//			case *id3.FrameURL:
//				fmt.Printf("%s\n", f.URL)
//		}
//	}
type Frame interface {
}

// FrameUnknown contains the payload of any frame whose ID is
// unknown to this package.
type FrameUnknown struct {
	ID   FrameID `v22:"?" v23:"?" v24:"?"`
	Data []byte
}

// FrameText may contain the payload of any type of text frame
// except for a user-defined TXXX text frame.  In v2.4, each text frame
// may contain one or more text strings.  In all other versions, only one
// text string may appear.
type FrameText struct {
	ID       FrameID  `v22:"T" v23:"T" v24:"T"`
	Type     TextType `id3:"texttype"`
	Encoding Encoding
	Text     []string
}

// NewFrameText creates a new text frame payload.
func NewFrameText(typ TextType, text string) *FrameText {
	return &FrameText{
		ID:       FrameID(v24TextTypeToFrameID[int(typ)]),
		Type:     typ,
		Encoding: EncodingUTF8,
		Text:     []string{text},
	}
}

// FrameTextCustom contains a custom text payload.
type FrameTextCustom struct {
	ID          FrameID `v22:"TXX" v23:"TXXX" v24:"TXXX"`
	Encoding    Encoding
	Description string
	Text        string
}

// NewFrameTextCustom creates a new custom text frame payload.
func NewFrameTextCustom(description, text string) *FrameTextCustom {
	return &FrameTextCustom{
		ID:          "TXXX",
		Encoding:    EncodingUTF8,
		Description: description,
		Text:        text,
	}
}

// FrameComment contains a full-text comment field.
type FrameComment struct {
	ID          FrameID `v22:"COM" v23:"COMM" v24:"COMM"`
	Encoding    Encoding
	Language    string `id3:"lang"`
	Description string
	Text        string
}

// NewFrameComment creates a new full-text comment frame.
func NewFrameComment(language, description, text string) *FrameComment {
	return &FrameComment{
		ID:          "COMM",
		Encoding:    EncodingUTF8,
		Language:    language,
		Description: description,
		Text:        text,
	}
}

// FrameURL may contain the payload of any type of URL frame except
// for the user-defined WXXX URL frame.
type FrameURL struct {
	ID  FrameID `v22:"W" v23:"W" v24:"W"`
	URL string  `id3:"iso88519"`
}

// FrameURLCustom contains a custom URL payload.
type FrameURLCustom struct {
	ID          FrameID `v22:"WXX" v23:"WXXX" v24:"WXXXX"`
	Encoding    Encoding
	Description string
	URL         string `id3:"iso88519"`
}

// FrameAttachedPicture contains the payload of an image frame.
type FrameAttachedPicture struct {
	ID          FrameID `v22:"PIC" v23:"APIC" v24:"APIC"`
	Encoding    Encoding
	MimeType    string `id3:"iso88519"`
	Type        PictureType
	Description string
	Data        []byte
}

// FrameUniqueFileID contains a unique file identifier for the MP3.
type FrameUniqueFileID struct {
	ID         FrameID `v22:"UFI" v23:"UFID" v24:"UFID"`
	Owner      string  `id3:"iso88519"`
	Identifier string  `id3:"iso88519"`
}

// FrameTermsOfUse contains the terms of use description for the MP3.
type FrameTermsOfUse struct {
	ID       FrameID `v23:"USER" v24:"USER"`
	Encoding Encoding
	Language string `id3:"lang"`
	Text     string
}

// FrameLyricsUnsync contains unsynchronized lyrics and text transcription
// data.
type FrameLyricsUnsync struct {
	ID         FrameID `v22:"ULT" v23:"USLT" v24:"USLT"`
	Encoding   Encoding
	Language   string `id3:"lang"`
	Descriptor string
	Text       string
}

// LyricSync describes a single syllable or event within a synchronized
// lyric or text frame (SYLT).
type LyricSync struct {
	Text      string
	TimeStamp uint32
}

// FrameLyricsSync contains synchronized lyrics or text information.
type FrameLyricsSync struct {
	ID              FrameID `v22:"SLT" v23:"SYLT" v24:"SYLT"`
	Encoding        Encoding
	Language        string `id3:"lang"`
	TimeStampFormat TimeStampFormat
	ContentType     LyricContentType
	Descriptor      string
	Sync            []LyricSync
}

// TempoSync describes a tempo change.
type TempoSync struct {
	BPM       uint16 `id3:"tempo"`
	TimeStamp uint32
}

// FrameSyncTempoCodes contains synchronized tempo codes.
type FrameSyncTempoCodes struct {
	ID              FrameID `v22:"STC" v23:"SYTC" v24:"SYTC"`
	TimeStampFormat TimeStampFormat
	Sync            []TempoSync
}

// FrameGroupID contains information describing the grouping of
// otherwise unrelated frames. If a frame contains an optional group
// identifier, there will be a corresponding GRID frame with data
// describing the group.
type FrameGroupID struct {
	ID      FrameID `v23:"GRID" v24:"GRID"`
	Owner   string  `id3:"iso88519"`
	GroupID GroupSymbol
	Data    []byte
}

// FramePrivate contains private information specific to a software
// producer.
type FramePrivate struct {
	ID    FrameID `v23:"PRIV" v24:"PRIV"`
	Owner string  `id3:"iso88519"`
	Data  []byte
}

// FramePlayCount tracks the number of times the MP3 file has been played.
type FramePlayCount struct {
	ID    FrameID `v22:"CNT" v23:"PCNT" v24:"PCNT"`
	Count uint64  `id3:"counter"`
}

// FramePopularimeter tracks the "popularimeter" value for an MP3 file.
type FramePopularimeter struct {
	ID     FrameID `v22:"POP" v23:"POPM" v24:"POPM"`
	Email  string  `id3:"iso88519"`
	Rating uint8
	Count  uint64 `id3:"counter"`
}

// frameTypes holds all possible frame payload types supported by ID3.
var frameTypes = []reflect.Type{
	reflect.TypeOf(FrameUnknown{}),
	reflect.TypeOf(FrameText{}),
	reflect.TypeOf(FrameTextCustom{}),
	reflect.TypeOf(FrameComment{}),
	reflect.TypeOf(FrameURL{}),
	reflect.TypeOf(FrameURLCustom{}),
	reflect.TypeOf(FrameAttachedPicture{}),
	reflect.TypeOf(FrameUniqueFileID{}),
	reflect.TypeOf(FrameTermsOfUse{}),
	reflect.TypeOf(FrameLyricsUnsync{}),
	reflect.TypeOf(FrameLyricsSync{}),
	reflect.TypeOf(FrameSyncTempoCodes{}),
	reflect.TypeOf(FrameGroupID{}),
	reflect.TypeOf(FramePrivate{}),
	reflect.TypeOf(FramePlayCount{}),
	reflect.TypeOf(FramePopularimeter{}),
}
