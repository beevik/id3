package id3

import (
	"io"
	"reflect"
)

//
// codec24
//

type codec24 struct {
	headerFlags flagMap
	frameFlags  flagMap
	bounds      boundsMap
	frameTypes  *frameTypeMap
}

func newCodec24() *codec24 {
	return &codec24{
		headerFlags: flagMap{
			{1 << 7, uint32(TagFlagUnsync)},
			{1 << 6, uint32(TagFlagExtended)},
			{1 << 5, uint32(TagFlagExperimental)},
			{1 << 4, uint32(TagFlagFooter)},
		},
		frameFlags: flagMap{
			{1 << 14, uint32(FrameFlagDiscardOnTagAlteration)},
			{1 << 13, uint32(FrameFlagDiscardOnFileAlteration)},
			{1 << 12, uint32(FrameFlagReadOnly)},
			{1 << 6, uint32(FrameFlagHasGroupInfo)},
			{1 << 3, uint32(FrameFlagCompressed)},
			{1 << 2, uint32(FrameFlagEncrypted)},
			{1 << 1, uint32(FrameFlagUnsynchronized)},
			{1 << 0, uint32(FrameFlagHasDataLength)},
		},
		bounds: boundsMap{
			"Encoding":         {0, 3},
			"GroupSymbol":      {0x80, 0xf0},
			"PictureType":      {0, 20},
			"TimeStampFormat":  {1, 2},
			"LyricContentType": {0, 8},
		},
		frameTypes: newFrameTypeMap(map[FrameType]string{
			FrameTypeAttachedPicture:             "APIC",
			FrameTypeAudioEncryption:             "AENC",
			FrameTypeAudioSeekPointIndex:         "ASPI",
			FrameTypeComment:                     "COMM",
			FrameTypeGroupID:                     "GRID",
			FrameTypePlayCount:                   "PCNT",
			FrameTypePopularimeter:               "POPM",
			FrameTypePrivate:                     "PRIV",
			FrameTypeLyricsSync:                  "SYLT",
			FrameTypeSyncTempoCodes:              "SYTC",
			FrameTypeTextAlbumName:               "TALB",
			FrameTypeTextBPM:                     "TBPM",
			FrameTypeTextCompilationItunes:       "TCMP",
			FrameTypeTextComposer:                "TCOM",
			FrameTypeTextGenre:                   "TCON",
			FrameTypeTextCopyright:               "TCOP",
			FrameTypeTextEncodingTime:            "TDEN",
			FrameTypeTextPlaylistDelay:           "TDLY",
			FrameTypeTextOriginalReleaseTime:     "TDOR",
			FrameTypeTextRecordingTime:           "TDRC",
			FrameTypeTextReleaseTime:             "TDRL",
			FrameTypeTextTaggingTime:             "TDTG",
			FrameTypeTextEncodedBy:               "TENC",
			FrameTypeTextLyricist:                "TEXT",
			FrameTypeTextFileType:                "TFLT",
			FrameTypeTextInvolvedPeople:          "TIPL",
			FrameTypeTextGroupDescription:        "TIT1",
			FrameTypeTextSongTitle:               "TIT2",
			FrameTypeTextSongSubtitle:            "TIT3",
			FrameTypeTextMusicalKey:              "TKEY",
			FrameTypeTextLanguage:                "TLAN",
			FrameTypeTextLengthInMs:              "TLEN",
			FrameTypeTextMusicians:               "TMCL",
			FrameTypeTextMediaType:               "TMED",
			FrameTypeTextMood:                    "TMOO",
			FrameTypeTextOriginalAlbum:           "TOAL",
			FrameTypeTextOriginalFileName:        "TOFN",
			FrameTypeTextOriginalLyricist:        "TOLY",
			FrameTypeTextOriginalPerformer:       "TOPE",
			FrameTypeTextOwner:                   "TOWN",
			FrameTypeTextArtist:                  "TPE1",
			FrameTypeTextAlbumArtist:             "TPE2",
			FrameTypeTextConductor:               "TPE3",
			FrameTypeTextRemixer:                 "TPE4",
			FrameTypeTextPartOfSet:               "TPOS",
			FrameTypeTextProducedNotice:          "TPRO",
			FrameTypeTextPublisher:               "TPUB",
			FrameTypeTextTrackNumber:             "TRCK",
			FrameTypeTextRadioStation:            "TRSN",
			FrameTypeTextRadioStationOwner:       "TRSO",
			FrameTypeTextAlbumSortOrderItunes:    "TSO2",
			FrameTypeTextAlbumSortOrder:          "TSOA",
			FrameTypeTextComposerSortOrderItunes: "TSOC",
			FrameTypeTextPerformerSortOrder:      "TSOP",
			FrameTypeTextTitleSortOrder:          "TSOT",
			FrameTypeTextISRC:                    "TSRC",
			FrameTypeTextEncodingSoftware:        "TSSE",
			FrameTypeTextSetSubtitle:             "TSST",
			FrameTypeTextCustom:                  "TXXX",
			FrameTypeUniqueFileID:                "UFID",
			FrameTypeTermsOfUse:                  "USER",
			FrameTypeLyricsUnsync:                "USLT",
			FrameTypeURLCommercial:               "WCOM",
			FrameTypeURLCopyright:                "WCOP",
			FrameTypeURLAudioFile:                "WOAF",
			FrameTypeURLArtist:                   "WOAR",
			FrameTypeURLAudioSource:              "WOAS",
			FrameTypeURLRadioStation:             "WORS",
			FrameTypeURLPayment:                  "WPAY",
			FrameTypeURLPublisher:                "WPUB",
			FrameTypeURLCustom:                   "WXXX",
			FrameTypeUnknown:                     "ZZZZ",
		}),
	}
}

// A property holds the reflection data necessary to update a property's
// value. Usually the property is a struct field.
type property struct {
	typ   reflect.Type
	name  string // field name if the property is a struct field
	value reflect.Value
}

// The state structure keeps track of persistent state required while
// decoding a single frame.
type state struct {
	frameID           FrameID
	frameType         FrameType
	encoding          Encoding // used by text frames
	bits              uint8    // used by ASPI frame
	fieldCount        int      // number of fields in the frame payload struct
	fieldIndex        int      // current field in the frame payload struct
	indexedDataLength uint32   // used for ASPI frames
}

func (c *codec24) HeaderFlags() flagMap {
	return c.headerFlags
}

func (c *codec24) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	// Read the first 6 bytes of the extended header so we can see how big
	// the additional extended data is.
	b := newInputBuf()
	if b.Read(r, 6); b.err != nil {
		return b.n, b.err
	}

	// Read the size of the extended data.
	size, err := decodeSyncSafeUint32(b.ConsumeBytes(4))
	if err != nil {
		return b.n, err
	}

	// The number of extended flag bytes must be 1.
	if b.ConsumeByte() != 1 {
		return b.n, ErrInvalidHeader
	}

	// Read the extended flags field.
	exFlags := b.ConsumeByte()
	if b.err != nil {
		return b.n, b.err
	}

	// Read the rest of the extended header into the buffer.
	if b.Read(r, int(size)-6); b.err != nil {
		return b.n, b.err
	}

	if (exFlags & (1 << 6)) != 0 {
		t.Flags |= TagFlagIsUpdate
		if b.ConsumeByte() != 0 || b.err != nil {
			return b.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 5)) != 0 {
		t.Flags |= TagFlagHasCRC
		data := b.ConsumeBytes(6)
		if b.err != nil || data[0] != 5 {
			return b.n, ErrInvalidHeader
		}
		t.CRC, err = decodeSyncSafeUint32(data[1:])
		if err != nil {
			return b.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 4)) != 0 {
		t.Flags |= TagFlagHasRestrictions
		data := b.ConsumeBytes(2)
		if b.err != nil || data[0] != 1 {
			return b.n, ErrInvalidHeader
		}
		t.Restrictions = uint16(data[0])<<8 | uint16(data[1])
	}

	return b.n, b.err
}

func (c *codec24) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header data to see if it's
	// padding.
	b := newInputBuf()
	if b.Read(r, 4); b.err != nil {
		return b.n, b.err
	}
	hd := b.ConsumeAll()
	if hd[0] == 0 && hd[1] == 0 && hd[2] == 0 && hd[3] == 0 {
		return b.n, errPaddingEncountered
	}

	// Read the remaining 6 bytes of the header data.
	if b.Read(r, 6); b.err != nil {
		return b.n, b.err
	}
	hd = append(hd, b.ConsumeAll()...)

	// Decode the frame's payload size.
	size, err := decodeSyncSafeUint32(hd[4:8])
	if err != nil {
		return b.n, err
	}
	if size < 1 {
		return b.n, ErrInvalidFrameHeader
	}

	// Decode the frame flags.
	flags := c.frameFlags.Decode(uint32(hd[8])<<8 | uint32(hd[9]))

	// Create the frame header structure.
	header := FrameHeader{
		FrameID: FrameID(hd[0:4]),
		Size:    int(size),
		Flags:   FrameFlags(flags),
	}

	// Read the rest of the frame into the input buffer.
	if b.Read(r, header.Size); b.err != nil {
		return b.n, b.err
	}

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (header.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		b.Replace(removeUnsyncCodes(b.buf))
	}

	// Scan extra header data indicated by the flags.
	if header.Flags != 0 {
		c.scanExtraHeaderData(b, &header)
		if b.err != nil {
			return b.n, b.err
		}
	}

	// Initialize the frame payload scan state.
	state := state{
		frameID:  header.FrameID,
		encoding: EncodingISO88591,
	}

	// Use reflection to interpret the payload's contents.
	typ := c.frameTypes.LookupReflectType(header.FrameID)
	p := property{
		typ:   typ,
		value: reflect.New(typ),
	}
	c.scanStruct(b, p, &state, 0)

	// Return the interpreted frame and header.
	if b.err == nil {
		*f = p.value.Interface().(Frame)

		// The frame's first field is always the header. Copy into it.
		ht := reflect.ValueOf(*f).Elem()
		ht.Field(0).Set(reflect.ValueOf(header))
	}

	return b.n, b.err
}

func (c *codec24) scanExtraHeaderData(b *inputBuf, h *FrameHeader) {
	// If the frame is compressed, it must include a data length indicator.
	if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
		b.err = ErrInvalidFrameFlags
		return
	}

	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		gid := b.ConsumeByte()
		if b.err != nil || gid < 0x80 || gid > 0xf0 {
			b.err = ErrInvalidFrameHeader
			return
		}
		h.GroupID = GroupSymbol(gid)
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		em := b.ConsumeByte()
		if b.err != nil || em < 0x80 || em > 0xf0 {
			b.err = ErrInvalidFrameHeader
			return
		}
		h.EncryptMethod = em
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		bb := b.ConsumeBytes(4)
		if b.err != nil {
			b.err = ErrInvalidFrameHeader
		}
		h.DataLength, b.err = decodeSyncSafeUint32(bb)
	}
}

func (c *codec24) scanStruct(s *inputBuf, p property, state *state, depth int) {
	if p.typ.Name() == "FrameHeader" {
		return
	}

	if depth == 0 {
		state.fieldCount = p.typ.NumField()
	}

	for ii, n := 0, p.typ.NumField(); ii < n; ii++ {
		if depth == 0 {
			state.fieldIndex = ii
		}

		field := p.typ.Field(ii)

		fp := property{
			typ:   field.Type,
			name:  field.Name,
			value: p.value.Elem().Field(ii),
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			c.scanUint8(s, fp, state)

		case reflect.Uint16:
			c.scanUint16(s, fp, state)

		case reflect.Uint32:
			c.scanUint32(s, fp, state)

		case reflect.Uint64:
			c.scanUint64(s, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.scanByteSlice(s, fp, state)
			case reflect.Uint32:
				c.scanUint32Slice(s, fp, state)
			case reflect.String:
				c.scanStringSlice(s, fp, state)
			case reflect.Struct:
				c.scanStructSlice(s, fp, state, depth+1)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.scanString(s, fp, state)

		case reflect.Struct:
			c.scanStruct(s, fp, state, depth+1)

		default:
			panic(errUnknownFieldType)
		}
	}
}

func (c *codec24) scanUint8(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	if p.typ.Name() == "FrameType" {
		state.frameType = c.frameTypes.LookupFrameType(state.frameID)
		p.value.SetUint(uint64(state.frameType))
		return
	}

	bounds, hasBounds := c.bounds[p.typ.Name()]

	value := b.ConsumeByte()
	if b.err != nil {
		return
	}

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		b.err = ErrInvalidFrame
		return
	}

	switch p.typ.Name() {
	case "Encoding":
		state.encoding = Encoding(value)
	case "Bits":
		state.bits = value
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint16(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	var value uint16
	switch p.typ.Name() {
	case "Tempo":
		value = uint16(b.ConsumeByte())
		if value == 0xff {
			value += uint16(b.ConsumeByte())
		}
	default:
		bb := b.ConsumeBytes(2)
		value = uint16(bb[0])<<8 | uint16(bb[1])
	}

	if b.err != nil {
		return
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint32(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	buf := b.ConsumeBytes(4)

	var value uint64
	for _, bb := range buf {
		value = (value << 8) | uint64(bb)
	}

	if state.frameType == FrameTypeAudioSeekPointIndex && p.name == "IndexedDataLength" {
		state.indexedDataLength = uint32(value)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanUint64(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	var buf []byte
	switch p.typ.Name() {
	case "Counter":
		buf = b.ConsumeAll()
	default:
		panic(errUnknownFieldType)
	}

	if b.err != nil {
		b.err = ErrInvalidFrame
		return
	}

	var value uint64
	for _, bb := range buf {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanByteSlice(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	bb := b.ConsumeAll()
	p.value.Set(reflect.ValueOf(bb))
}

func (c *codec24) scanUint32Slice(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	if p.typ.Elem().Name() != "IndexOffset" {
		panic(errUnknownFieldType)
	}

	var offsets []IndexOffset

	ff := b.ConsumeAll()
	switch state.bits {
	case 8:
		offsets = make([]IndexOffset, len(ff))
		for _, f := range ff {
			frac := uint32(f)
			offset := (frac*state.indexedDataLength + (1 << 7)) << 8
			if offset > state.indexedDataLength {
				offset = state.indexedDataLength
			}
			offsets = append(offsets, IndexOffset(offset))
		}

	case 16:
		offsets = make([]IndexOffset, len(ff)/2)
		for ii := 0; ii < len(ff); ii += 2 {
			frac := uint32(ff[ii])<<8 | uint32(ff[ii+1])
			offset := (frac*state.indexedDataLength + (1 << 15)) << 16
			if offset > state.indexedDataLength {
				offset = state.indexedDataLength
			}
			offsets = append(offsets, IndexOffset(offset))
		}

	default:
		b.err = ErrInvalidFrame
		return
	}

	p.value.Set(reflect.ValueOf(offsets))
}

func (c *codec24) scanStringSlice(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	ss := b.ConsumeStrings(state.encoding)
	if b.err != nil {
		return
	}
	p.value.Set(reflect.ValueOf(ss))
}

func (c *codec24) scanStructSlice(b *inputBuf, p property, state *state, depth int) {
	if b.err != nil {
		return
	}

	elems := make([]reflect.Value, 0)
	for b.Len() > 0 {
		etyp := p.typ.Elem()
		ep := property{
			typ:   etyp,
			value: reflect.New(etyp),
		}

		c.scanStruct(b, ep, state, depth+1)
		if b.err != nil {
			return
		}

		elems = append(elems, ep.value)
	}

	slice := reflect.MakeSlice(p.typ, len(elems), len(elems))
	for ii := range elems {
		slice.Index(ii).Set(elems[ii].Elem())
	}
	p.value.Set(slice)
}

func (c *codec24) scanString(b *inputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	if p.typ.Name() == "FrameID" {
		p.value.SetString(string(state.frameID))
		return
	}

	enc := state.encoding
	if p.typ.Name() == "WesternString" {
		enc = EncodingISO88591
	}

	var str string
	switch p.typ.Name() {
	case "LanguageString":
		str = b.ConsumeFixedLengthString(3, EncodingISO88591)
	default:
		str = b.ConsumeNextString(enc)
	}

	if b.err != nil {
		return
	}

	p.value.SetString(str)
}

func (c *codec24) EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error) {
	b := newOutputBuf()

	p := property{
		typ:   reflect.TypeOf(f).Elem(),
		value: reflect.ValueOf(f).Elem(),
	}
	state := state{
		encoding: EncodingISO88591,
	}

	c.outputStruct(b, p, &state, 0)
	if b.err != nil {
		return b.n, b.err
	}

	h := HeaderOf(f)
	h.FrameID = c.frameTypes.LookupFrameID(state.frameType)
	h.Size = b.Len()

	hdr := make([]byte, 10)

	encodeSyncSafeUint32(hdr[4:8], uint32(h.Size))

	copy(hdr[0:4], []byte(h.FrameID))

	flags := c.frameFlags.Encode(uint32(h.Flags))
	hdr[8] = byte(flags >> 8)
	hdr[9] = byte(flags)

	// TODO: Handle extended header creation

	n, err := w.Write(hdr)
	if err != nil {
		return n, err
	}

	nn, err := w.Write(b.Bytes())
	n += nn
	return n, err
}

func (c *codec24) outputStruct(b *outputBuf, p property, state *state, depth int) {
	if p.typ.Name() == "FrameHeader" {
		return
	}

	if depth == 0 {
		state.fieldCount = p.typ.NumField()
	}

	for i, n := 0, p.typ.NumField(); i < n; i++ {
		if depth == 0 {
			state.fieldIndex = i
		}

		field := p.typ.Field(i)

		fp := property{
			typ:   field.Type,
			name:  field.Name,
			value: p.value.Field(i),
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			c.outputUint8(b, fp, state)

		case reflect.Uint16:
			c.outputUint16(b, fp, state)

		case reflect.Uint32:
			c.outputUint32(b, fp, state)

		case reflect.Uint64:
			c.outputUint64(b, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.outputByteSlice(b, fp, state)
			case reflect.Uint32:
				c.outputUint32Slice(b, fp, state)
			case reflect.String:
				c.outputStringSlice(b, fp, state)
			case reflect.Struct:
				c.outputStructSlice(b, fp, state, depth+1)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.outputString(b, fp, state, depth)

		case reflect.Struct:
			c.outputStruct(b, fp, state, depth+1)

		default:
			panic(errUnknownFieldType)
		}
	}
}

func (c *codec24) outputUint8(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	value := uint8(p.value.Uint())

	if p.typ.Name() == "FrameType" {
		state.frameType = FrameType(value)
		return
	}

	bounds, hasBounds := c.bounds[p.typ.Name()]

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		b.err = ErrInvalidFrame
		return
	}

	b.AddByte(value)
	if b.err != nil {
		return
	}

	switch p.typ.Name() {
	case "Encoding":
		state.encoding = Encoding(value)
	case "Bits":
		state.bits = value
	}
}

func (c *codec24) outputUint16(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	v := uint16(p.value.Uint())

	switch p.typ.Name() {
	case "Tempo":
		if v > 2*0xff {
			b.err = ErrInvalidFrame
			return
		}
		if v < 0xff {
			b.AddByte(uint8(v))
		} else {
			b.AddByte(0xff)
			b.AddByte(uint8(v - 0xff))
		}
	default:
		bb := []byte{byte(v >> 8), byte(v)}
		b.AddBytes(bb)
	}
}

func (c *codec24) outputUint32(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	v := uint32(p.value.Uint())
	bb := []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}

	if state.frameType == FrameTypeAudioSeekPointIndex && p.name == "IndexedDataLength" {
		state.indexedDataLength = v
	}

	b.AddBytes(bb)
}

func (c *codec24) outputUint64(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	v := p.value.Uint()

	switch p.typ.Name() {
	case "Counter":
		bb := make([]byte, 0, 4)
		for v != 0 {
			bb = append(bb, byte(v&0xff))
			v = v >> 8
		}
		for len(bb) < 4 {
			bb = append(bb, 0)
		}
		for i := len(bb) - 1; i >= 0; i-- {
			b.AddByte(bb[i])
		}
	default:
		panic(errUnknownFieldType)
	}
}

func (c *codec24) outputUint32Slice(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	if p.typ.Elem().Name() != "IndexOffset" {
		panic(errUnknownFieldType)
	}

	n := p.value.Len()
	slice := p.value.Slice(0, n)

	switch state.bits {
	case 8:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 8) / state.indexedDataLength
			if frac >= (1 << 8) {
				frac = (1 << 8) - 1
			}
			b.AddByte(byte(frac))
		}

	case 16:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 16) / state.indexedDataLength
			if frac >= (1 << 16) {
				frac = (1 << 16) - 1
			}
			bb := []byte{byte(frac >> 8), byte(frac)}
			b.AddBytes(bb)
		}

	default:
		b.err = ErrInvalidFrame
	}
}

func (c *codec24) outputByteSlice(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	var bb []byte
	reflect.ValueOf(&bb).Elem().Set(p.value)
	b.AddBytes(bb)
}

func (c *codec24) outputStringSlice(b *outputBuf, p property, state *state) {
	if b.err != nil {
		return
	}

	var ss []string
	reflect.ValueOf(&ss).Elem().Set(p.value)
	b.AddStrings(ss, state.encoding)
}

func (c *codec24) outputStructSlice(b *outputBuf, p property, state *state, depth int) {
	if b.err != nil {
		return
	}

	n := p.value.Len()
	slice := p.value.Slice(0, n)

	for i := 0; i < n; i++ {
		elem := slice.Index(i)

		ep := property{
			typ:   elem.Type(),
			value: elem,
		}

		c.outputStruct(b, ep, state, depth+1)
		if b.err != nil {
			return
		}
	}
}

func (c *codec24) outputString(b *outputBuf, p property, state *state, depth int) {
	if b.err != nil {
		return
	}

	v := p.value.String()

	if p.typ.Name() == "FrameID" {
		state.frameID = FrameID(v)
		return
	}

	enc := state.encoding
	if p.typ.Name() == "WesternString" {
		enc = EncodingISO88591
	}

	switch p.typ.Name() {
	case "LanguageString":
		b.AddFixedLengthString(v, 3, enc)
	default:
		// Always terminate strings unless they are the last struct field
		// of the root level struct.
		term := depth > 0 || (state.fieldIndex != state.fieldCount-1)
		b.AddString(v, enc, term)
	}
}
