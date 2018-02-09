package id3

import "reflect"

// A FrameHolder holds the header and payload of an ID3 frame.
// DEPRECATED.
type FrameHolder struct {
	header FrameHeader
	Frame  Frame
}

// Size returns the encoded size of the frame, not including the header.
func (f *FrameHolder) Size() int {
	return f.header.Size
}

// ID returns the 4-character ID string currently assigned to the frame.
func (f *FrameHolder) ID() string {
	return f.header.ID
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
	FrameTypeTextPartOfSet                         // TPOS
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
	FrameTypeTextTitleSortOrder      // TSOT (v2.4 only)

	// Text frames: v2.3-only frames (ID3v2.3 spec)
	FrameTypeTextDate           // TDAT (subsumed by TDRC in v2.4)
	FrameTypeTextTime           // TIME (subsumed by TDRC in v2.4)
	FrameTypeTextRecordingDates // TRDA (subsumed by TDRC in v2.4)
	FrameTypeTextSize           // TSIZ

	// Text frames: custom text
	FrameTypeTextCustom // TXXX

	// URL frames
	FrameTypeURLCommercial   // WCOM
	FrameTypeURLCopyright    // WCOP
	FrameTypeURLAudioFile    // WOAF
	FrameTypeURLArtist       // WOAR
	FrameTypeURLAudioSource  // WOAS
	FrameTypeURLRadioStation // WORS
	FrameTypeURLPayment      // WPAY
	FrameTypeURLPublisher    // WPUB
	FrameTypeURLCustom       // WXXX

	// Other frames
	FrameTypeComment         // COMM
	FrameTypeAttachedPicture // APIC
	FrameTypeUniqueFileID    // UFID
	FrameTypeTermsOfUse      // USER
	FrameTypeLyricsUnsync    // USLT
	FrameTypeLyricsSync      // SYLT
	FrameTypeSyncTempoCodes  // SYTC
	FrameTypeGroupID         // GRID
	FrameTypePrivate         // PRIV
	FrameTypePlayCount       // PCNT
	FrameTypePopularimeter   // POPM

	// Non-standard values
	FrameTypeUnknown
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
// is described futher in group identifier frames.
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
	Type    FrameType
	FrameID string
	Data    []byte
}

// FrameText may contain the payload of any type of text frame
// except for a custom text frame.  In v2.4, each text frame
// may contain one or more text strings.  In all other versions, only one
// text string may appear.
type FrameText struct {
	Type     FrameType
	Encoding Encoding
	Text     []string
}

// NewFrameText creates a new text frame payload with a single text string.
func NewFrameText(typ FrameType, text string) *FrameText {
	return &FrameText{
		Type:     typ,
		Encoding: EncodingUTF8,
		Text:     []string{text},
	}
}

// FrameTextCustom contains a custom text payload.
type FrameTextCustom struct {
	Type        FrameType
	Encoding    Encoding
	Description string
	Text        string
}

// NewFrameTextCustom creates a new custom text frame payload.
func NewFrameTextCustom(description, text string) *FrameTextCustom {
	return &FrameTextCustom{
		Type:        FrameTypeTextCustom,
		Encoding:    EncodingUTF8,
		Description: description,
		Text:        text,
	}
}

// FrameComment contains a full-text comment field.
type FrameComment struct {
	Type        FrameType
	Encoding    Encoding
	Language    string `id3:"lang"`
	Description string
	Text        string
}

// NewFrameComment creates a new full-text comment frame.
func NewFrameComment(language, description, text string) *FrameComment {
	return &FrameComment{
		Type:        FrameTypeComment,
		Encoding:    EncodingUTF8,
		Language:    language,
		Description: description,
		Text:        text,
	}
}

// FrameURL may contain the payload of any type of URL frame except
// for the user-defined WXXX URL frame.
type FrameURL struct {
	Type FrameType
	URL  string `id3:"iso88519"`
}

// NewFrameURL creates a URL frame of the requested type.
func NewFrameURL(typ FrameType, url string) *FrameURL {
	return &FrameURL{
		Type: typ,
		URL:  url,
	}
}

// FrameURLCustom contains a custom URL payload.
type FrameURLCustom struct {
	Type        FrameType
	Encoding    Encoding
	Description string
	URL         string `id3:"iso88519"`
}

// NewFrameURLCustom creates a custom URL frame.
func NewFrameURLCustom(description, url string) *FrameURLCustom {
	return &FrameURLCustom{
		Type:        FrameTypeURLCustom,
		Encoding:    EncodingUTF8,
		Description: description,
		URL:         url,
	}
}

// FrameAttachedPicture contains the payload of an image frame.
type FrameAttachedPicture struct {
	Type        FrameType
	Encoding    Encoding
	MimeType    string `id3:"iso88519"`
	PictureType PictureType
	Description string
	Data        []byte
}

// NewFrameAttachedPicture creates a new attached-picture frame.
func NewFrameAttachedPicture(mimeType, description string, typ PictureType, data []byte) *FrameAttachedPicture {
	return &FrameAttachedPicture{
		Type:        FrameTypeAttachedPicture,
		Encoding:    EncodingUTF8,
		MimeType:    mimeType,
		PictureType: typ,
		Description: description,
		Data:        data,
	}
}

// FrameUniqueFileID contains a unique file identifier for the MP3.
type FrameUniqueFileID struct {
	Type       FrameType
	Owner      string `id3:"iso88519"`
	Identifier string `id3:"iso88519"`
}

// NewFrameUniqueFileID creates a new Unique FileID frame.
func NewFrameUniqueFileID(owner, id string) *FrameUniqueFileID {
	return &FrameUniqueFileID{
		Type:       FrameTypeUniqueFileID,
		Owner:      owner,
		Identifier: id,
	}
}

// FrameTermsOfUse contains the terms of use description for the MP3.
type FrameTermsOfUse struct {
	Type     FrameType
	Encoding Encoding
	Language string `id3:"lang"`
	Text     string
}

// NewFrameTermsOfUse creates a new terms-of-use frame.
func NewFrameTermsOfUse(language, text string) *FrameTermsOfUse {
	return &FrameTermsOfUse{
		Type:     FrameTypeTermsOfUse,
		Encoding: EncodingUTF8,
		Language: language,
		Text:     text,
	}
}

// FrameLyricsUnsync contains unsynchronized lyrics and text transcription
// data.
type FrameLyricsUnsync struct {
	Type       FrameType
	Encoding   Encoding
	Language   string `id3:"lang"`
	Descriptor string
	Text       string
}

// NewFrameLyricsUnsync creates a new unsynchronized lyrics frame.
func NewFrameLyricsUnsync(language, descriptor, lyrics string) *FrameLyricsUnsync {
	return &FrameLyricsUnsync{
		Type:       FrameTypeLyricsUnsync,
		Encoding:   EncodingUTF8,
		Language:   language,
		Descriptor: descriptor,
		Text:       lyrics,
	}
}

// LyricsSync describes a single syllable or event within a synchronized
// lyric or text frame (SYLT).
type LyricsSync struct {
	Text      string
	TimeStamp uint32
}

// FrameLyricsSync contains synchronized lyrics or text information.
type FrameLyricsSync struct {
	Type        FrameType
	Encoding    Encoding
	Language    string `id3:"lang"`
	Format      TimeStampFormat
	ContentType LyricContentType
	Descriptor  string
	Sync        []LyricsSync
}

// NewFrameLyricsSync creates a new synchronized lyrics frame.
func NewFrameLyricsSync(language, descriptor string,
	format TimeStampFormat, typ LyricContentType) *FrameLyricsSync {
	return &FrameLyricsSync{
		Type:        FrameTypeLyricsSync,
		Encoding:    EncodingUTF8,
		Language:    language,
		Format:      format,
		ContentType: typ,
		Sync:        []LyricsSync{},
	}
}

// AddSync inserts a time-stamped syllable into a synchronized lyric
// frame. It inserts the syllable in sorted order by time stamp.
func (f *FrameLyricsSync) AddSync(sync LyricsSync) {
	var i int
	for i = range f.Sync {
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

// TempoSync describes a tempo change.
type TempoSync struct {
	BPM       uint16 `id3:"tempo"`
	TimeStamp uint32
}

// FrameSyncTempoCodes contains synchronized tempo codes.
type FrameSyncTempoCodes struct {
	Type            FrameType
	TimeStampFormat TimeStampFormat
	Sync            []TempoSync
}

// NewFrameSyncTempoCodes creates a new synchronized tempo codes frame.
func NewFrameSyncTempoCodes(format TimeStampFormat) *FrameSyncTempoCodes {
	return &FrameSyncTempoCodes{
		Type:            FrameTypeSyncTempoCodes,
		TimeStampFormat: format,
		Sync:            []TempoSync{},
	}
}

// AddSync inserts a time-stamped syllable into a synchronized lyric
// frame. It inserts the syllable in sorted order by time stamp.
func (f *FrameSyncTempoCodes) AddSync(sync TempoSync) {
	var i int
	for i = range f.Sync {
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

// FrameGroupID contains information describing the grouping of
// otherwise unrelated frames. If a frame contains an optional group
// identifier, there will be a corresponding GRID frame with data
// describing the group.
type FrameGroupID struct {
	Type    FrameType
	Owner   string `id3:"iso88519"`
	GroupID GroupSymbol
	Data    []byte
}

// NewFrameGroupID creates a new group identifier frame.
func NewFrameGroupID(owner string, groupID GroupSymbol, data []byte) *FrameGroupID {
	return &FrameGroupID{
		Type:    FrameTypeGroupID,
		Owner:   owner,
		GroupID: groupID,
		Data:    data,
	}
}

// FramePrivate contains private information specific to a software
// producer.
type FramePrivate struct {
	Type  FrameType
	Owner string `id3:"iso88519"`
	Data  []byte
}

// NewFramePrivate creates a new private information frame.
func NewFramePrivate(owner string, data []byte) *FramePrivate {
	return &FramePrivate{
		Type:  FrameTypePrivate,
		Owner: owner,
		Data:  data,
	}
}

// FramePlayCount tracks the number of times the MP3 file has been played.
type FramePlayCount struct {
	Type  FrameType
	Count uint64 `id3:"counter"`
}

// NewFramePlayCount creates a new play count frame.
func NewFramePlayCount(count uint64) *FramePlayCount {
	return &FramePlayCount{
		Type:  FrameTypePlayCount,
		Count: count,
	}
}

// FramePopularimeter tracks the "popularimeter" value for an MP3 file.
type FramePopularimeter struct {
	Type   FrameType
	Email  string `id3:"iso88519"`
	Rating uint8
	Count  uint64 `id3:"counter"`
}

// NewFramePopularimeter creates a new "popularimeter" frame.
func NewFramePopularimeter(email string, rating uint8, count uint64) *FramePopularimeter {
	return &FramePopularimeter{
		Type:   FrameTypePopularimeter,
		Email:  email,
		Rating: rating,
		Count:  count,
	}
}

// frameTypes holds all possible frame payload types supported by ID3.
var frameData = []struct {
	frameType   FrameType
	reflectType reflect.Type
}{
	{FrameTypeUnknown, reflect.TypeOf(FrameUnknown{})},
	{FrameTypeTextAlbumArtist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextGroupDescription, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSongTitle, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSongSubtitle, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumName, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalAlbum, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTrackNumber, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPartOfSet, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSetSubtitle, reflect.TypeOf(FrameText{})},
	{FrameTypeTextISRC, reflect.TypeOf(FrameText{})},
	{FrameTypeTextArtist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumArtist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextConductor, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRemixer, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalPerformer, reflect.TypeOf(FrameText{})},
	{FrameTypeTextLyricist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalLyricist, reflect.TypeOf(FrameText{})},
	{FrameTypeTextComposer, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMusicians, reflect.TypeOf(FrameText{})},
	{FrameTypeTextInvolvedPeople, reflect.TypeOf(FrameText{})},
	{FrameTypeTextEncodedBy, reflect.TypeOf(FrameText{})},
	{FrameTypeTextBPM, reflect.TypeOf(FrameText{})},
	{FrameTypeTextLengthInMs, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMusicalKey, reflect.TypeOf(FrameText{})},
	{FrameTypeTextLanguage, reflect.TypeOf(FrameText{})},
	{FrameTypeTextGenre, reflect.TypeOf(FrameText{})},
	{FrameTypeTextFileType, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMediaType, reflect.TypeOf(FrameText{})},
	{FrameTypeTextMood, reflect.TypeOf(FrameText{})},
	{FrameTypeTextCopyright, reflect.TypeOf(FrameText{})},
	{FrameTypeTextProducedNotice, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPublisher, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOwner, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRadioStation, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRadioStationOwner, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalFileName, reflect.TypeOf(FrameText{})},
	{FrameTypeTextPlaylistDelay, reflect.TypeOf(FrameText{})},
	{FrameTypeTextEncodingTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextOriginalReleaseTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRecordingTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextReleaseTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTaggingTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextEncodingSoftware, reflect.TypeOf(FrameText{})},
	{FrameTypeTextAlbumSortOrder, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTitleSortOrder, reflect.TypeOf(FrameText{})},
	{FrameTypeTextDate, reflect.TypeOf(FrameText{})},
	{FrameTypeTextTime, reflect.TypeOf(FrameText{})},
	{FrameTypeTextRecordingDates, reflect.TypeOf(FrameText{})},
	{FrameTypeTextSize, reflect.TypeOf(FrameText{})},
	{FrameTypeTextCustom, reflect.TypeOf(FrameText{})},
	{FrameTypeURLCommercial, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLCopyright, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLAudioFile, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLArtist, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLAudioSource, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLRadioStation, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLPayment, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLPublisher, reflect.TypeOf(FrameURL{})},
	{FrameTypeURLCustom, reflect.TypeOf(FrameURL{})},
	{FrameTypeComment, reflect.TypeOf(FrameComment{})},
	{FrameTypeAttachedPicture, reflect.TypeOf(FrameAttachedPicture{})},
	{FrameTypeUniqueFileID, reflect.TypeOf(FrameUniqueFileID{})},
	{FrameTypeTermsOfUse, reflect.TypeOf(FrameTermsOfUse{})},
	{FrameTypeLyricsUnsync, reflect.TypeOf(FrameLyricsUnsync{})},
	{FrameTypeLyricsSync, reflect.TypeOf(FrameLyricsSync{})},
	{FrameTypeSyncTempoCodes, reflect.TypeOf(FrameSyncTempoCodes{})},
	{FrameTypeGroupID, reflect.TypeOf(FrameGroupID{})},
	{FrameTypePrivate, reflect.TypeOf(FramePrivate{})},
	{FrameTypePlayCount, reflect.TypeOf(FramePlayCount{})},
	{FrameTypePopularimeter, reflect.TypeOf(FramePopularimeter{})},
}

type frameTypeMap struct {
	FrameTypeToFrameID   map[FrameType]string
	FrameIDToReflectType map[string]reflect.Type
	FrameIDToFrameType   map[string]FrameType
}

func newFrameTypeMap(frameTypeToFrameID map[FrameType]string) *frameTypeMap {
	m := &frameTypeMap{}
	m.FrameTypeToFrameID = frameTypeToFrameID

	m.FrameIDToReflectType = make(map[string]reflect.Type)
	for _, d := range frameData {
		id := m.FrameTypeToFrameID[d.frameType]
		m.FrameIDToReflectType[id] = d.reflectType
	}

	m.FrameIDToFrameType = make(map[string]FrameType)
	for k, v := range m.FrameTypeToFrameID {
		m.FrameIDToFrameType[v] = k
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
	t, ok := m.FrameIDToReflectType[id]
	if !ok {
		t = reflect.TypeOf(FrameUnknown{})
	}
	return t
}

func (m *frameTypeMap) LookupFrameType(id string) FrameType {
	t, ok := m.FrameIDToFrameType[id]
	if !ok {
		t = FrameTypeUnknown
	}
	return t
}
