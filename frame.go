package id3

import "reflect"

// A FrameHeader holds the data described by a frame header.
type FrameHeader struct {
	FrameID       string     // Frame ID string
	Size          int        // Frame size not including 10-byte header
	Flags         FrameFlags // Flags
	GroupID       uint8      // Optional group identifier
	EncryptMethod uint8      // Optional encryption method identifier
	DataLength    uint32     // Optional data length (if FrameFlagHasDataLength is set)
}

// SetFlag sets the requested frame flag on or off.
func (h *FrameHeader) SetFlag(flag FrameFlags, value bool) {
	switch {
	case (flag & (FrameFlagEncrypted | FrameFlagHasGroupID)) != 0:
		// ignore. SetGroupID or SetEncryptMethod should be used instead.
	case value == true:
		h.Flags |= flag
	case value == false:
		h.Flags &= ^flag
	}
}

// SetGroupID sets a group identifier on the frame. The id value must be
// between 0x80 and 0xf0, and the tag must also have a group description frame
// for the id. Use an id of 0 to clear the frame's group identifier.
func (h *FrameHeader) SetGroupID(id uint8) error {
	switch {
	case id == 0:
		h.GroupID = 0
		h.Flags &= ^FrameFlagHasGroupID
	case id >= 0x80 && id <= 0xf0:
		h.GroupID = id
		h.Flags |= FrameFlagHasGroupID
	default:
		return ErrInvalidGroupID
	}
	return nil
}

// SetEncryptMethod sets an encryption method on the frame. The method value
// must be between 0x80 and 0xf0, and the tag must also have an encryption
// frame for the method value. Use a method value of 0 to clear the frame's
// encryption method.
func (h *FrameHeader) SetEncryptMethod(method uint8) error {
	switch {
	case method == 0:
		h.EncryptMethod = 0
		h.Flags &= ^FrameFlagEncrypted
	case method >= 0x80 && method <= 0xf0:
		h.EncryptMethod = method
		h.Flags |= FrameFlagEncrypted
	default:
		return ErrInvalidEncryptMethod
	}
	return nil
}

// A WesternString is a string that is always saved into the tag using
// ISO 8559-1 encoding.
type WesternString string

// FrameFlags describe flags that may appear within a FrameHeader. Not all
// flags are supported by all versions of the ID3 codec.
type FrameFlags uint32

// All possible FrameFlags.
const (
	FrameFlagDiscardOnTagAlteration  FrameFlags = 1 << iota // Discard frame if tag is altered
	FrameFlagDiscardOnFileAlteration                        // Discard frame if file is altered
	FrameFlagReadOnly                                       // Frame is read-only
	FrameFlagHasGroupID                                     // Frame has group info
	FrameFlagCompressed                                     // Frame is compressed
	FrameFlagEncrypted                                      // Frame is encrypted
	FrameFlagUnsynchronized                                 // Frame is unsynchronized
	FrameFlagHasDataLength                                  // Frame has a data length indicator
)

// A FrameType value identifies the type of an ID3 frame.
type FrameType uint8

// All standard types of text frames.
const (
	// Text frames: Identification (ID3v2.4 spec section 4.2.1)
	FrameTypeTextGroupDescription FrameType = iota // TIT1
	FrameTypeTextSongTitle                         // TIT2
	FrameTypeTextSongSubtitle                      // TIT3
	FrameTypeTextAlbumName                         // TALB
	FrameTypeTextOriginalAlbum                     // TOAL
	FrameTypeTextTrackNumber                       // TRCK
	FrameTypeTextPartOfSet                         // TPOS (CD number)
	FrameTypeTextSetSubtitle                       // TSST (v2.4 only)
	FrameTypeTextISRC                              // TSRC

	// Text frames: Involved persons (ID3v2.4 spec section 4.2.2)
	FrameTypeTextArtist            // TPE1
	FrameTypeTextAlbumArtist       // TPE2
	FrameTypeTextConductor         // TPE3
	FrameTypeTextRemixer           // TPE4
	FrameTypeTextOriginalPerformer // TOPE
	FrameTypeTextLyricist          // TEXT
	FrameTypeTextOriginalLyricist  // TOLY
	FrameTypeTextComposer          // TCOM
	FrameTypeTextMusicians         // TMCL (v2.4 only)
	FrameTypeTextInvolvedPeople    // TIPL (v2.4 only)
	FrameTypeTextEncodedBy         // TENC

	// Text frames: Derived and subjective properties (ID3v2.4 spec section 4.2.3)
	FrameTypeTextBPM        // TBPM
	FrameTypeTextLengthInMs // TLEN
	FrameTypeTextMusicalKey // TKEY
	FrameTypeTextLanguage   // TLAN
	FrameTypeTextGenre      // TCON (see Genre)
	FrameTypeTextFileType   // TFLT (see FileType)
	FrameTypeTextMediaType  // TMED
	FrameTypeTextMood       // TMOO (v2.4 only)

	// Text frames: Rights and license (ID3v2.4 spec section 4.2.4)
	FrameTypeTextCopyright         // TCOP
	FrameTypeTextProducedNotice    // TPRO (v2.4 only)
	FrameTypeTextPublisher         // TPUB
	FrameTypeTextOwner             // TOWN
	FrameTypeTextRadioStation      // TRSN
	FrameTypeTextRadioStationOwner // TRSO

	// Text frames: Other text frames (ID3v2.4 spec section 4.2.5)
	FrameTypeTextOriginalFileName    // TOFN
	FrameTypeTextPlaylistDelay       // TDLY
	FrameTypeTextEncodingTime        // TDEN (v2.4 only)
	FrameTypeTextOriginalReleaseTime // TDOR (v2.4) or TORY (v2.3)
	FrameTypeTextRecordingTime       // TDRC (v2.4) or TYER (v2.3)
	FrameTypeTextReleaseTime         // TDRL (v2.4 only)
	FrameTypeTextTaggingTime         // TDTG (v2.4 only)
	FrameTypeTextEncodingSoftware    // TSSE
	FrameTypeTextAlbumSortOrder      // TSOA (v2.4 only)
	FrameTypeTextPerformerSortOrder  // TSOP (v2.4 only)
	FrameTypeTextTitleSortOrder      // TSOT (v2.4 only)

	// Text frames: v2.3-only frames (ID3v2.3 spec)
	FrameTypeTextDate           // TDAT (v2.3 only, subsumed by TDRC in v2.4)
	FrameTypeTextTime           // TIME (v2.3 only, subsumed by TDRC in v2.4)
	FrameTypeTextRecordingDates // TRDA (v2.3 only, subsumed by TDRC in v2.4)
	FrameTypeTextSize           // TSIZ (v2.3 only)

	// Text frames: non-standard frames that commonly appear in the wild
	FrameTypeTextCompilationItunes       // TCMP (iTunes)
	FrameTypeTextAlbumSortOrderItunes    // TSO2 (iTunes)
	FrameTypeTextComposerSortOrderItunes // TSOC (iTunes)

	// Text frames: custom text
	FrameTypeTextCustom // TXXX

	// URL frames
	FrameTypeURLArtist       // WOAR
	FrameTypeURLAudioFile    // WOAF
	FrameTypeURLAudioSource  // WOAS
	FrameTypeURLCommercial   // WCOM
	FrameTypeURLCopyright    // WCOP
	FrameTypeURLCustom       // WXXX
	FrameTypeURLPayment      // WPAY
	FrameTypeURLPublisher    // WPUB
	FrameTypeURLRadioStation // WORS

	// Other frames
	FrameTypeAttachedPicture     // APIC
	FrameTypeAudioEncryption     // AENC
	FrameTypeAudioSeekPointIndex // ASPI
	FrameTypeComment             // COMM
	FrameTypeGroupID             // GRID
	FrameTypeLyricsSync          // SYLT
	FrameTypeLyricsUnsync        // USLT
	FrameTypePlayCount           // PCNT
	FrameTypePopularimeter       // POPM
	FrameTypePrivate             // PRIV
	FrameTypeSyncTempoCodes      // SYTC
	FrameTypeTermsOfUse          // USER
	FrameTypeUniqueFileID        // UFID

	// Non-standard values
	FrameTypeUnknown
)

// A Frame is an interface capable of representing the payload of any of the
// possible frame types (e.g., FrameText, FrameURL, etc.).
//
// Use a type assertion to access the frame's contents. For example, given a
// Frame f:
//
// 	if ft, ok := f.(*id3.FrameText); ok {
//		fmt.Printf("%s\n", ft.Text)
//	}
//
// OR:
//
//	switch ff := f.(type) {
// 		case *id3.FrameText:
// 			fmt.Printf("%v\n", ff.Text)
//		case *id3.FrameURL:
//			fmt.Printf("%s\n", ff.URL)
//	}
type Frame interface {
}

// HeaderOf returns a pointer to the frame's header data.
func HeaderOf(f Frame) *FrameHeader {
	var hdr *FrameHeader
	ft := reflect.ValueOf(f).Elem()
	hv := reflect.ValueOf(&hdr).Elem()
	hv.Set(ft.Field(0).Addr())
	return hdr
}

// PictureType describes the type of picture stored within an Attached
// Picture frame.
type PictureType uint8

// All possible picture types.
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

// FrameAttachedPicture contains the payload of an image frame.
type FrameAttachedPicture struct {
	Header      FrameHeader
	FrameType   FrameType
	Encoding    Encoding
	MimeType    WesternString
	PictureType PictureType
	Description string
	Data        []byte
}

// NewFrameAttachedPicture creates a new attached-picture frame.
func NewFrameAttachedPicture(mimeType, description string, typ PictureType, data []byte) *FrameAttachedPicture {
	return &FrameAttachedPicture{
		FrameType:   FrameTypeAttachedPicture,
		Encoding:    EncodingUTF8,
		MimeType:    WesternString(mimeType),
		PictureType: typ,
		Description: description,
		Data:        data,
	}
}

// FrameAudioEncryption indicates if the audio stream is encrypted and, if
// so, provides data used by an encryption algorithm to decode it.
type FrameAudioEncryption struct {
	Header        FrameHeader
	FrameType     FrameType
	Owner         WesternString
	PreviewStart  uint16
	PreviewLength uint16
	Data          []byte
}

// NewFrameAudioEncryption creates a new audio encryption frame.
func NewFrameAudioEncryption(owner string, previewStart, previewLength uint16, data []byte) *FrameAudioEncryption {
	return &FrameAudioEncryption{
		FrameType:     FrameTypeAudioEncryption,
		Owner:         WesternString(owner),
		PreviewStart:  previewStart,
		PreviewLength: previewLength,
		Data:          data,
	}
}

// FrameAudioSeekPointIndex contains audio indexing data useful for locating
// important positions within the encoded audio data.
type FrameAudioSeekPointIndex struct {
	Header            FrameHeader
	FrameType         FrameType
	IndexedDataStart  uint32
	IndexedDataLength uint32
	IndexPoints       uint16
	BitsPerIndex      uint8 // must be 8 or 16
	IndexOffsets      []uint32
}

// NewFrameAudioSeekPointIndex creates a new audio seek point index frame.
func NewFrameAudioSeekPointIndex(indexedDataStart, indexedDataLength uint32) *FrameAudioSeekPointIndex {
	return &FrameAudioSeekPointIndex{
		FrameType:         FrameTypeAudioSeekPointIndex,
		IndexedDataStart:  indexedDataStart,
		IndexedDataLength: indexedDataLength,
		IndexPoints:       0,
		BitsPerIndex:      16,
		IndexOffsets:      []uint32{},
	}
}

// AddIndexOffset inserts a new index offset into the audio seek point index
// frame. The offset is relative to the frame's IndexedDataStart value and
// must be less than the frame's IndexedDataLength field.
func (f *FrameAudioSeekPointIndex) AddIndexOffset(o uint32) {
	var i int
	for i = 0; i < len(f.IndexOffsets); i++ {
		if f.IndexOffsets[i] > o {
			break
		}
	}

	switch {
	case i == len(f.IndexOffsets):
		f.IndexOffsets = append(f.IndexOffsets, o)
	default:
		f.IndexOffsets = append(f.IndexOffsets, 0)
		copy(f.IndexOffsets[i+1:], f.IndexOffsets[i:])
		f.IndexOffsets[i] = o
	}

	f.IndexPoints++
}

// FrameComment contains a full-text comment field.
type FrameComment struct {
	Header      FrameHeader
	FrameType   FrameType
	Encoding    Encoding
	Language    string
	Description string
	Text        string
}

// NewFrameComment creates a new full-text comment frame.
func NewFrameComment(language, description, text string) *FrameComment {
	return &FrameComment{
		FrameType:   FrameTypeComment,
		Encoding:    EncodingUTF8,
		Language:    language,
		Description: description,
		Text:        text,
	}
}

// FrameGroupID contains information describing the grouping of
// otherwise unrelated frames. If a frame contains an optional group
// identifier, there will be a corresponding GRID frame with data
// describing the group.
type FrameGroupID struct {
	Header    FrameHeader
	FrameType FrameType
	Owner     WesternString
	GroupID   uint8
	Data      []byte
}

// NewFrameGroupID creates a new group identifier frame.
func NewFrameGroupID(owner string, groupID uint8, data []byte) *FrameGroupID {
	return &FrameGroupID{
		FrameType: FrameTypeGroupID,
		Owner:     WesternString(owner),
		GroupID:   groupID,
		Data:      data,
	}
}

// FrameLyricsUnsync contains unsynchronized lyrics and text transcription
// data.
type FrameLyricsUnsync struct {
	Header     FrameHeader
	FrameType  FrameType
	Encoding   Encoding
	Language   string
	Descriptor string
	Text       string
}

// NewFrameLyricsUnsync creates a new unsynchronized lyrics frame.
func NewFrameLyricsUnsync(language, descriptor, lyrics string) *FrameLyricsUnsync {
	return &FrameLyricsUnsync{
		FrameType:  FrameTypeLyricsUnsync,
		Encoding:   EncodingUTF8,
		Language:   language,
		Descriptor: descriptor,
		Text:       lyrics,
	}
}

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

// TimeStampFormat indicates the type of time stamp used: milliseconds or
// MPEG frame.
type TimeStampFormat byte

// All possible values of the TimeStampFormat type.
const (
	TimeStampFrames TimeStampFormat = 1 + iota
	TimeStampMilliseconds
)

// LyricsSync describes a single syllable or event within a synchronized
// lyric or text frame (SYLT).
type LyricsSync struct {
	Text      string
	TimeStamp uint32
}

// FrameLyricsSync contains synchronized lyrics or text information.
type FrameLyricsSync struct {
	Header           FrameHeader
	FrameType        FrameType
	Encoding         Encoding
	Language         string
	Format           TimeStampFormat
	LyricContentType LyricContentType
	Descriptor       string
	Sync             []LyricsSync
}

// NewFrameLyricsSync creates a new synchronized lyrics frame.
func NewFrameLyricsSync(language, descriptor string,
	format TimeStampFormat, typ LyricContentType) *FrameLyricsSync {
	return &FrameLyricsSync{
		FrameType:        FrameTypeLyricsSync,
		Encoding:         EncodingUTF8,
		Language:         language,
		Format:           format,
		LyricContentType: typ,
		Descriptor:       descriptor,
		Sync:             []LyricsSync{},
	}
}

// AddSync inserts a time-stamped syllable into a synchronized lyric
// frame. It inserts the syllable in sorted order by time stamp.
func (f *FrameLyricsSync) AddSync(sync LyricsSync) {
	var i int
	for i = 0; i < len(f.Sync); i++ {
		if f.Sync[i].TimeStamp > sync.TimeStamp {
			break
		}
	}
	switch {
	case i == len(f.Sync):
		f.Sync = append(f.Sync, sync)
	default:
		f.Sync = append(f.Sync, LyricsSync{})
		copy(f.Sync[i+1:], f.Sync[i:])
		f.Sync[i] = sync
	}
}

// FramePrivate contains private information specific to a software
// producer.
type FramePrivate struct {
	Header    FrameHeader
	FrameType FrameType
	Owner     WesternString
	Data      []byte
}

// NewFramePrivate creates a new private information frame.
func NewFramePrivate(owner string, data []byte) *FramePrivate {
	return &FramePrivate{
		FrameType: FrameTypePrivate,
		Owner:     WesternString(owner),
		Data:      data,
	}
}

// FramePlayCount tracks the number of times the MP3 file has been played.
type FramePlayCount struct {
	Header    FrameHeader
	FrameType FrameType
	Counter   uint64
}

// NewFramePlayCount creates a new play count frame.
func NewFramePlayCount(counter uint64) *FramePlayCount {
	return &FramePlayCount{
		FrameType: FrameTypePlayCount,
		Counter:   counter,
	}
}

// FramePopularimeter tracks the "popularimeter" value for an MP3 file.
type FramePopularimeter struct {
	Header    FrameHeader
	FrameType FrameType
	Email     WesternString
	Rating    uint8
	Counter   uint64
}

// NewFramePopularimeter creates a new "popularimeter" frame.
func NewFramePopularimeter(email string, rating uint8, counter uint64) *FramePopularimeter {
	return &FramePopularimeter{
		FrameType: FrameTypePopularimeter,
		Email:     WesternString(email),
		Rating:    rating,
		Counter:   counter,
	}
}

// TempoSync describes a tempo change.
type TempoSync struct {
	BPM       uint16
	TimeStamp uint32
}

// FrameSyncTempoCodes contains synchronized tempo codes.
type FrameSyncTempoCodes struct {
	Header          FrameHeader
	FrameType       FrameType
	TimeStampFormat TimeStampFormat
	Sync            []TempoSync
}

// NewFrameSyncTempoCodes creates a new synchronized tempo codes frame.
func NewFrameSyncTempoCodes(format TimeStampFormat) *FrameSyncTempoCodes {
	return &FrameSyncTempoCodes{
		FrameType:       FrameTypeSyncTempoCodes,
		TimeStampFormat: format,
		Sync:            []TempoSync{},
	}
}

// AddSync inserts a time-stamped syllable into a synchronized lyric
// frame. It inserts the syllable in sorted order by time stamp.
func (f *FrameSyncTempoCodes) AddSync(sync TempoSync) {
	var i int
	for i = 0; i < len(f.Sync); i++ {
		if f.Sync[i].TimeStamp > sync.TimeStamp {
			break
		}
	}
	switch {
	case i == len(f.Sync):
		f.Sync = append(f.Sync, sync)
	default:
		f.Sync = append(f.Sync, TempoSync{})
		copy(f.Sync[i+1:], f.Sync[i:])
		f.Sync[i] = sync
	}
}

// FrameTermsOfUse contains the terms of use description for the MP3.
type FrameTermsOfUse struct {
	Header    FrameHeader
	FrameType FrameType
	Encoding  Encoding
	Language  string
	Text      string
}

// NewFrameTermsOfUse creates a new terms-of-use frame.
func NewFrameTermsOfUse(language, text string) *FrameTermsOfUse {
	return &FrameTermsOfUse{
		FrameType: FrameTypeTermsOfUse,
		Encoding:  EncodingUTF8,
		Language:  language,
		Text:      text,
	}
}

// FrameText may contain the payload of any type of text frame
// except for a custom text frame.  In v2.4, each text frame
// may contain one or more text strings.  In all other versions, only one
// text string may appear.
type FrameText struct {
	Header    FrameHeader
	FrameType FrameType
	Encoding  Encoding
	Text      []string
}

// NewFrameText creates a new text frame payload with a single text string.
func NewFrameText(typ FrameType, text string) *FrameText {
	return &FrameText{
		FrameType: typ,
		Encoding:  EncodingUTF8,
		Text:      []string{text},
	}
}

// FrameTextCustom contains a custom text payload.
type FrameTextCustom struct {
	Header      FrameHeader
	FrameType   FrameType
	Encoding    Encoding
	Description string
	Text        string
}

// NewFrameTextCustom creates a new custom text frame payload.
func NewFrameTextCustom(description, text string) *FrameTextCustom {
	return &FrameTextCustom{
		FrameType:   FrameTypeTextCustom,
		Encoding:    EncodingUTF8,
		Description: description,
		Text:        text,
	}
}

// FrameUnknown contains the payload of any frame whose ID is
// unknown to this package.
type FrameUnknown struct {
	Header    FrameHeader
	FrameType FrameType
	FrameID   string
	Data      []byte
}

// NewFrameUnknown creates a new frame of unknown type.
func NewFrameUnknown(id string, data []byte) *FrameUnknown {
	return &FrameUnknown{
		FrameType: FrameTypeUnknown,
		FrameID:   id,
		Data:      data,
	}
}

// FrameURL may contain the payload of any type of URL frame except
// for the user-defined WXXX URL frame.
type FrameURL struct {
	Header    FrameHeader
	FrameType FrameType
	URL       WesternString
}

// NewFrameURL creates a URL frame of the requested type.
func NewFrameURL(typ FrameType, url string) *FrameURL {
	return &FrameURL{
		FrameType: typ,
		URL:       WesternString(url),
	}
}

// FrameURLCustom contains a custom URL payload.
type FrameURLCustom struct {
	Header      FrameHeader
	FrameType   FrameType
	Encoding    Encoding
	Description string
	URL         WesternString
}

// NewFrameURLCustom creates a custom URL frame.
func NewFrameURLCustom(description, url string) *FrameURLCustom {
	return &FrameURLCustom{
		FrameType:   FrameTypeURLCustom,
		Encoding:    EncodingUTF8,
		Description: description,
		URL:         WesternString(url),
	}
}

// FrameUniqueFileID contains a unique file identifier for the MP3.
type FrameUniqueFileID struct {
	Header     FrameHeader
	FrameType  FrameType
	Owner      WesternString
	Identifier WesternString
}

// NewFrameUniqueFileID creates a new Unique FileID frame.
func NewFrameUniqueFileID(owner, id string) *FrameUniqueFileID {
	return &FrameUniqueFileID{
		FrameType:  FrameTypeUniqueFileID,
		Owner:      WesternString(owner),
		Identifier: WesternString(id),
	}
}

//
// Frame list and type map
//

// frameList holds all possible frame payload types supported by ID3.
var frameList = []struct {
	frameType   FrameType
	reflectType reflect.Type
}{
	{FrameTypeAttachedPicture, reflect.TypeOf(FrameAttachedPicture{})},
	{FrameTypeAudioEncryption, reflect.TypeOf(FrameAudioEncryption{})},
	{FrameTypeAudioSeekPointIndex, reflect.TypeOf(FrameAudioSeekPointIndex{})},
	{FrameTypeComment, reflect.TypeOf(FrameComment{})},
	{FrameTypeGroupID, reflect.TypeOf(FrameGroupID{})},
	{FrameTypeLyricsSync, reflect.TypeOf(FrameLyricsSync{})},
	{FrameTypeLyricsUnsync, reflect.TypeOf(FrameLyricsUnsync{})},
	{FrameTypePlayCount, reflect.TypeOf(FramePlayCount{})},
	{FrameTypePopularimeter, reflect.TypeOf(FramePopularimeter{})},
	{FrameTypePrivate, reflect.TypeOf(FramePrivate{})},
	{FrameTypeSyncTempoCodes, reflect.TypeOf(FrameSyncTempoCodes{})},
	{FrameTypeTermsOfUse, reflect.TypeOf(FrameTermsOfUse{})},
	{FrameTypeTextAlbumArtist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumArtist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumName, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumSortOrder, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumSortOrderItunes, reflect.TypeOf(FrameText{})},
	{FrameTypeTextArtist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextBPM, reflect.TypeOf(FrameText{})},
	{FrameTypeTextCompilationItunes, reflect.TypeOf(FrameText{})},
	{FrameTypeTextComposer, reflect.TypeOf(FrameText{})},
	{FrameTypeTextComposerSortOrderItunes, reflect.TypeOf(FrameText{})},
	{FrameTypeTextConductor, reflect.TypeOf(FrameText{})},
	{FrameTypeTextCopyright, reflect.TypeOf(FrameText{})},
	{FrameTypeTextCustom, reflect.TypeOf(FrameTextCustom{})},
	{FrameTypeTextDate, reflect.TypeOf(FrameText{})},
	{FrameTypeTextEncodedBy, reflect.TypeOf(FrameText{})},
	{FrameTypeTextEncodingSoftware, reflect.TypeOf(FrameText{})},
	{FrameTypeTextEncodingTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextFileType, reflect.TypeOf(FrameText{})},
	{FrameTypeTextGenre, reflect.TypeOf(FrameText{})},
	{FrameTypeTextGroupDescription, reflect.TypeOf(FrameText{})},
	{FrameTypeTextInvolvedPeople, reflect.TypeOf(FrameText{})},
	{FrameTypeTextISRC, reflect.TypeOf(FrameText{})},
	{FrameTypeTextLanguage, reflect.TypeOf(FrameText{})},
	{FrameTypeTextLengthInMs, reflect.TypeOf(FrameText{})},
	{FrameTypeTextLyricist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMediaType, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMood, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMusicalKey, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMusicians, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalAlbum, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalFileName, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalLyricist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalPerformer, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalReleaseTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOwner, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPartOfSet, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPerformerSortOrder, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPlaylistDelay, reflect.TypeOf(FrameText{})},
	{FrameTypeTextProducedNotice, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPublisher, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRadioStation, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRadioStationOwner, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRecordingDates, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRecordingTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextReleaseTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRemixer, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSetSubtitle, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSize, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSongSubtitle, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSongTitle, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTaggingTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTitleSortOrder, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTrackNumber, reflect.TypeOf(FrameText{})},
	{FrameTypeUniqueFileID, reflect.TypeOf(FrameUniqueFileID{})},
	{FrameTypeUnknown, reflect.TypeOf(FrameUnknown{})},
	{FrameTypeURLArtist, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLAudioFile, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLAudioSource, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLCommercial, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLCopyright, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLCustom, reflect.TypeOf(FrameURLCustom{})},
	{FrameTypeURLPayment, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLPublisher, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLRadioStation, reflect.TypeOf(FrameURL{})},
}

type frameTypeMap struct {
	FrameTypeToFrameID   map[FrameType]string
	FrameIDToFrameType   map[string]FrameType
	FrameIDToReflectType map[string]reflect.Type
}

func newFrameTypeMap(frameTypeToFrameID map[FrameType]string) *frameTypeMap {
	m := &frameTypeMap{
		FrameTypeToFrameID:   frameTypeToFrameID,
		FrameIDToFrameType:   make(map[string]FrameType),
		FrameIDToReflectType: make(map[string]reflect.Type),
	}

	for k, v := range m.FrameTypeToFrameID {
		m.FrameIDToFrameType[v] = k
	}

	for _, f := range frameList {
		id := m.FrameTypeToFrameID[f.frameType]
		m.FrameIDToReflectType[id] = f.reflectType
	}

	return m
}

func (m *frameTypeMap) LookupFrameID(t FrameType) string {
	id, ok := m.FrameTypeToFrameID[t]
	if !ok {
		id = m.FrameTypeToFrameID[FrameTypeUnknown]
	}
	return id
}

func (m *frameTypeMap) LookupReflectType(id string) reflect.Type {
	t, ok := m.FrameIDToReflectType[string(id)]
	if !ok {
		t = reflect.TypeOf(FrameUnknown{})
	}
	return t
}

func (m *frameTypeMap) LookupFrameType(id string) FrameType {
	t, ok := m.FrameIDToFrameType[string(id)]
	if !ok {
		t = FrameTypeUnknown
	}
	return t
}
