package id3

import (
	"fmt"
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
			{1 << 6, uint32(FrameFlagHasGroupID)},
			{1 << 3, uint32(FrameFlagCompressed)},
			{1 << 2, uint32(FrameFlagEncrypted)},
			{1 << 1, uint32(FrameFlagUnsynchronized)},
			{1 << 0, uint32(FrameFlagHasDataLength)},
		},
		bounds: boundsMap{
			"Encoding":         {0, 3, ErrInvalidEncoding},
			"GroupID":          {0x80, 0xf0, ErrInvalidGroupID},
			"LyricContentType": {0, 8, ErrInvalidLyricContentType},
			"PictureType":      {0, 20, ErrInvalidPictureType},
			"TimeStampFormat":  {1, 2, ErrInvalidTimeStampFormat},
		},
		frameTypes: newFrameTypeMap(map[FrameType]string{
			FrameTypeAttachedPicture:              "APIC",
			FrameTypeAudioEncryption:              "AENC",
			FrameTypeAudioSeekPointIndex:          "ASPI",
			FrameTypeComment:                      "COMM",
			FrameTypeEncryptionMethodRegistration: "ENCR",
			FrameTypeGroupID:                      "GRID",
			FrameTypePlayCount:                    "PCNT",
			FrameTypePopularimeter:                "POPM",
			FrameTypePrivate:                      "PRIV",
			FrameTypeLyricsSync:                   "SYLT",
			FrameTypeSyncTempoCodes:               "SYTC",
			FrameTypeTextAlbumName:                "TALB",
			FrameTypeTextBPM:                      "TBPM",
			FrameTypeTextCompilationItunes:        "TCMP",
			FrameTypeTextComposer:                 "TCOM",
			FrameTypeTextGenre:                    "TCON",
			FrameTypeTextCopyright:                "TCOP",
			FrameTypeTextEncodingTime:             "TDEN",
			FrameTypeTextPlaylistDelay:            "TDLY",
			FrameTypeTextOriginalReleaseTime:      "TDOR",
			FrameTypeTextRecordingTime:            "TDRC",
			FrameTypeTextReleaseTime:              "TDRL",
			FrameTypeTextTaggingTime:              "TDTG",
			FrameTypeTextEncodedBy:                "TENC",
			FrameTypeTextLyricist:                 "TEXT",
			FrameTypeTextFileType:                 "TFLT",
			FrameTypeTextInvolvedPeople:           "TIPL",
			FrameTypeTextGroupDescription:         "TIT1",
			FrameTypeTextSongTitle:                "TIT2",
			FrameTypeTextSongSubtitle:             "TIT3",
			FrameTypeTextMusicalKey:               "TKEY",
			FrameTypeTextLanguage:                 "TLAN",
			FrameTypeTextLengthInMs:               "TLEN",
			FrameTypeTextMusicians:                "TMCL",
			FrameTypeTextMediaType:                "TMED",
			FrameTypeTextMood:                     "TMOO",
			FrameTypeTextOriginalAlbum:            "TOAL",
			FrameTypeTextOriginalFileName:         "TOFN",
			FrameTypeTextOriginalLyricist:         "TOLY",
			FrameTypeTextOriginalPerformer:        "TOPE",
			FrameTypeTextOwner:                    "TOWN",
			FrameTypeTextArtist:                   "TPE1",
			FrameTypeTextAlbumArtist:              "TPE2",
			FrameTypeTextConductor:                "TPE3",
			FrameTypeTextRemixer:                  "TPE4",
			FrameTypeTextPartOfSet:                "TPOS",
			FrameTypeTextProducedNotice:           "TPRO",
			FrameTypeTextPublisher:                "TPUB",
			FrameTypeTextTrackNumber:              "TRCK",
			FrameTypeTextRadioStation:             "TRSN",
			FrameTypeTextRadioStationOwner:        "TRSO",
			FrameTypeTextAlbumSortOrderItunes:     "TSO2",
			FrameTypeTextAlbumSortOrder:           "TSOA",
			FrameTypeTextComposerSortOrderItunes:  "TSOC",
			FrameTypeTextPerformerSortOrder:       "TSOP",
			FrameTypeTextTitleSortOrder:           "TSOT",
			FrameTypeTextISRC:                     "TSRC",
			FrameTypeTextEncodingSoftware:         "TSSE",
			FrameTypeTextSetSubtitle:              "TSST",
			FrameTypeTextCustom:                   "TXXX",
			FrameTypeUniqueFileID:                 "UFID",
			FrameTypeTermsOfUse:                   "USER",
			FrameTypeLyricsUnsync:                 "USLT",
			FrameTypeURLCommercial:                "WCOM",
			FrameTypeURLCopyright:                 "WCOP",
			FrameTypeURLAudioFile:                 "WOAF",
			FrameTypeURLArtist:                    "WOAR",
			FrameTypeURLAudioSource:               "WOAS",
			FrameTypeURLRadioStation:              "WORS",
			FrameTypeURLPayment:                   "WPAY",
			FrameTypeURLPublisher:                 "WPUB",
			FrameTypeURLCustom:                    "WXXX",
			FrameTypeUnknown:                      "ZZZZ",
		}),
	}
}

// A property holds the reflection data necessary to update a property's
// value. Usually the property is a struct field.
type property struct {
	typ   reflect.Type
	value reflect.Value
	name  string
}

// The state structure keeps track of persistent state required while
// decoding a single frame.
type state struct {
	frameID     string     // current frame ID
	frameType   FrameType  // current frame type
	structStack valueStack // stack of active struct values
	fieldCount  int        // current frame's field count
	fieldIndex  int        // current frame field index
}

func (c *codec24) DecodeHeader(t *Tag, r io.Reader) (int, error) {
	hdr := make([]byte, 10)
	n, err := io.ReadFull(r, hdr)
	if err != nil {
		return n, err
	}

	// Allow the codec to interpret the flags field.
	flags := uint32(hdr[5])
	t.Flags = TagFlags(c.headerFlags.Decode(flags))

	// Process the tag size.
	size, err := decodeSyncSafeUint32(hdr[6:10])
	t.Size = int(size)
	return n, err
}

func (c *codec24) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	// Read the first 6 bytes of the extended header so we can see how big
	// the additional extended data is.
	rr := newReader()
	if rr.LoadFrom(r, 6); rr.err != nil {
		return rr.n, rr.err
	}

	// Read the size of the extended data.
	size, err := decodeSyncSafeUint32(rr.ConsumeBytes(4))
	if err != nil {
		return rr.n, err
	}

	// The number of extended flag bytes must be 1.
	if rr.ConsumeByte() != 1 {
		return rr.n, ErrInvalidHeader
	}

	// Read the extended flags field.
	exFlags := rr.ConsumeByte()
	if rr.err != nil {
		return rr.n, rr.err
	}

	// Read the rest of the extended header into the buffer.
	if rr.LoadFrom(r, int(size)-6); rr.err != nil {
		return rr.n, rr.err
	}

	if (exFlags & (1 << 6)) != 0 {
		t.Flags |= TagFlagIsUpdate
		if rr.ConsumeByte() != 0 || rr.err != nil {
			return rr.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 5)) != 0 {
		t.Flags |= TagFlagHasCRC
		data := rr.ConsumeBytes(6)
		if rr.err != nil || data[0] != 5 {
			return rr.n, ErrInvalidHeader
		}
		t.CRC, err = decodeSyncSafeUint32(data[1:])
		if err != nil {
			return rr.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 4)) != 0 {
		t.Flags |= TagFlagHasRestrictions
		data := rr.ConsumeBytes(2)
		if rr.err != nil || data[0] != 1 {
			return rr.n, ErrInvalidHeader
		}
		t.Restrictions = data[1]
	}

	return rr.n, rr.err
}

func (c *codec24) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header data to see if it's
	// padding.
	rr := newReader()
	if rr.LoadFrom(r, 4); rr.err != nil {
		return rr.n, rr.err
	}
	hd := rr.ConsumeBytes(4)
	if hd[0] == 0 && hd[1] == 0 && hd[2] == 0 && hd[3] == 0 {
		return rr.n, errPaddingEncountered
	}

	// Read the remaining 6 bytes of the header data.
	if rr.LoadFrom(r, 6); rr.err != nil {
		return rr.n, rr.err
	}
	hd = append(hd, rr.ConsumeAll()...)

	// Decode the frame's payload size.
	size, err := decodeSyncSafeUint32(hd[4:8])
	if err != nil {
		return rr.n, err
	}
	if size < 1 {
		return rr.n, ErrInvalidFrameHeader
	}

	// Decode the frame flags.
	flags := c.frameFlags.Decode(uint32(hd[8])<<8 | uint32(hd[9]))

	// Create the frame header structure.
	header := FrameHeader{
		FrameID: string(hd[0:4]),
		Size:    int(size),
		Flags:   FrameFlags(flags),
	}

	// Read the rest of the frame into the input buffer.
	if rr.LoadFrom(r, header.Size); rr.err != nil {
		return rr.n, rr.err
	}

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (header.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		in := rr.ConsumeAll()
		out := removeUnsyncCodes(in)
		rr.ReplaceBuffer(out)
	}

	// Scan extra header data indicated by the flags.
	if header.Flags != 0 {
		c.scanExtraHeaderData(rr, &header)
		if rr.err != nil {
			return rr.n, rr.err
		}
	}

	// Initialize the frame payload scan state.
	state := state{
		frameID: header.FrameID,
	}

	// Use reflection to interpret the payload's contents.
	typ := c.frameTypes.LookupReflectType(header.FrameID)
	p := property{
		typ:   typ,
		value: reflect.New(typ),
		name:  "",
	}
	c.scanStruct(rr, p, &state)

	// Return the interpreted frame and header.
	if rr.err == nil {
		*f = p.value.Interface().(Frame)

		// The frame's first field is always the header. Copy into it.
		ht := reflect.ValueOf(*f).Elem()
		ht.Field(0).Set(reflect.ValueOf(header))
	}

	return rr.n, rr.err
}

func (c *codec24) scanExtraHeaderData(rr *reader, h *FrameHeader) {
	// If the frame is compressed, it must include a data length indicator.
	if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
		rr.err = ErrInvalidFrameFlags
		return
	}

	if (h.Flags & FrameFlagHasGroupID) != 0 {
		gid := rr.ConsumeByte()
		if rr.err != nil || gid < 0x80 || gid > 0xf0 {
			rr.err = ErrInvalidGroupID
			return
		}
		h.GroupID = gid
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		em := rr.ConsumeByte()
		if rr.err != nil || em < 0x80 || em > 0xf0 {
			rr.err = ErrInvalidEncryptMethod
			return
		}
		h.EncryptMethod = em
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		b := rr.ConsumeBytes(4)
		if rr.err != nil {
			rr.err = ErrInvalidFrameHeader
		}
		h.DataLength, rr.err = decodeSyncSafeUint32(b)
	}
}

var counter = 0

func (c *codec24) scanStruct(rr *reader, p property, state *state) {
	if p.typ.Name() == "FrameHeader" {
		return
	}

	state.structStack.push(p.value.Elem())
	if state.structStack.depth() == 1 {
		state.fieldCount = p.typ.NumField()
	}

	for ii, n := 0, p.typ.NumField(); ii < n; ii++ {
		if state.structStack.depth() == 1 {
			state.fieldIndex = ii
		}

		field := p.typ.Field(ii)

		fp := property{
			typ:   field.Type,
			value: p.value.Elem().Field(ii),
			name:  field.Name,
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			c.scanUint8(rr, fp, state)

		case reflect.Uint16:
			c.scanUint16(rr, fp, state)

		case reflect.Uint32:
			c.scanUint32(rr, fp, state)

		case reflect.Uint64:
			c.scanUint64(rr, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.scanByteSlice(rr, fp, state)
			case reflect.Uint32:
				c.scanUint32Slice(rr, fp, state)
			case reflect.String:
				c.scanStringSlice(rr, fp, state)
			case reflect.Struct:
				c.scanStructSlice(rr, fp, state)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.scanString(rr, fp, state)

		case reflect.Struct:
			c.scanStruct(rr, fp, state)

		default:
			panic(errUnknownFieldType)
		}
	}

	state.structStack.pop()
}

func (c *codec24) scanUint8(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	if p.typ.Name() == "FrameType" {
		state.frameType = c.frameTypes.LookupFrameType(state.frameID)
		p.value.SetUint(uint64(state.frameType))
		return
	}

	bounds, hasBounds := c.bounds[p.name]

	value := rr.ConsumeByte()
	if rr.err != nil {
		return
	}

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		rr.err = bounds.err
		return
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint16(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	var value uint16
	switch p.name {
	case "BPM":
		value = uint16(rr.ConsumeByte())
		if value == 0xff {
			value += uint16(rr.ConsumeByte())
		}
	default:
		b := rr.ConsumeBytes(2)
		value = uint16(b[0])<<8 | uint16(b[1])
	}

	if rr.err != nil {
		return
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint32(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	b := rr.ConsumeBytes(4)

	var value uint64
	for _, bb := range b {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanUint64(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	var b []byte
	switch p.name {
	case "Counter":
		b = rr.ConsumeAll()
	default:
		panic(errUnknownFieldType)
	}

	var value uint64
	for _, bb := range b {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanByteSlice(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	b := rr.ConsumeAll()
	p.value.Set(reflect.ValueOf(b))
}

func (c *codec24) scanUint32Slice(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	if p.name != "IndexOffsets" {
		panic(errUnknownFieldType)
	}

	sf := state.structStack.first()
	length := uint32(sf.FieldByName("IndexedDataLength").Uint())
	bits := uint32(sf.FieldByName("BitsPerIndex").Uint())

	var offsets []uint32

	ff := rr.ConsumeAll()
	switch bits {
	case 8:
		offsets = make([]uint32, 0, len(ff))
		for _, f := range ff {
			frac := uint32(f)
			offset := (frac*length + (1 << 7)) >> 8
			if offset > length {
				offset = length
			}
			offsets = append(offsets, offset)
		}

	case 16:
		offsets = make([]uint32, 0, len(ff)/2)
		for ii := 0; ii < len(ff); ii += 2 {
			frac := uint32(ff[ii])<<8 | uint32(ff[ii+1])
			offset := (frac*length + (1 << 15)) >> 16
			if offset > length {
				offset = length
			}
			offsets = append(offsets, offset)
		}

	default:
		rr.err = ErrInvalidBits
		return
	}

	p.value.Set(reflect.ValueOf(offsets))
}

func (c *codec24) scanStringSlice(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	sf := state.structStack.first()
	enc := Encoding(sf.FieldByName("Encoding").Uint())
	ss := rr.ConsumeStrings(enc)
	if rr.err != nil {
		return
	}
	p.value.Set(reflect.ValueOf(ss))
}

func (c *codec24) scanStructSlice(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	elems := make([]reflect.Value, 0)
	for i := 0; rr.Len() > 0; i++ {
		etyp := p.typ.Elem()
		ep := property{
			typ:   etyp,
			value: reflect.New(etyp),
			name:  fmt.Sprintf("%s[%d]", p.name, i),
		}

		c.scanStruct(rr, ep, state)
		if rr.err != nil {
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

func (c *codec24) scanString(rr *reader, p property, state *state) {
	if rr.err != nil {
		return
	}

	switch p.name {
	case "FrameID":
		p.value.SetString(string(state.frameID))
		return
	case "Language":
		str := rr.ConsumeFixedLengthString(3, EncodingISO88591)
		p.value.SetString(str)
		return
	}

	var enc Encoding
	switch p.typ.Name() {
	case "WesternString":
		enc = EncodingISO88591
	default:
		sf := state.structStack.first()
		enc = Encoding(sf.FieldByName("Encoding").Uint())
	}

	str := rr.ConsumeNextString(enc)

	if rr.err != nil {
		return
	}

	p.value.SetString(str)
}

func (c *codec24) EncodeExtendedHeader(t *Tag, w io.Writer) (int, error) {
	ww := newWriter()

	// Placeholder for size and flags
	ww.StoreBytes([]byte{0, 0, 0, 0, 1, 0})

	var flags uint8
	if (t.Flags & TagFlagIsUpdate) != 0 {
		ww.StoreByte(0)
		flags |= (1 << 6)
	}

	if (t.Flags & TagFlagHasCRC) != 0 {
		ww.StoreByte(5)
		b := make([]byte, 5)
		encodeSyncSafeUint32(b, t.CRC)
		ww.StoreBytes(b)
		flags |= (1 << 5)
	}

	if (t.Flags & TagFlagHasRestrictions) != 0 {
		ww.StoreByte(1)
		ww.StoreByte(t.Restrictions)
		flags |= (1 << 4)
	}

	if flags != 0 {
		b := ww.Bytes()
		encodeSyncSafeUint32(b[0:4], uint32(ww.Len()))
		b[5] = flags

		return w.Write(b)
	}
	return 0, nil
}

func (c *codec24) EncodeHeader(t *Tag, w io.Writer) (int, error) {
	flags := uint8(c.headerFlags.Encode(uint32(t.Flags)))
	hdr := []byte{'I', 'D', '3', byte(t.Version), 0, flags, 0, 0, 0, 0}
	err := encodeSyncSafeUint32(hdr[6:10], uint32(t.Size))
	if err != nil {
		return 0, err
	}

	return w.Write(hdr)
}

func (c *codec24) EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error) {
	ww := newWriter()

	p := property{
		typ:   reflect.TypeOf(f).Elem(),
		value: reflect.ValueOf(f).Elem(),
		name:  "",
	}
	state := state{}

	c.outputStruct(ww, p, &state)
	if ww.err != nil {
		return ww.n, ww.err
	}

	h := HeaderOf(f)
	h.FrameID = c.frameTypes.LookupFrameID(state.frameType)
	h.Size = ww.Len()

	// TODO: Perform frame-only unsync

	exBuf := newWriter()
	if h.Flags != 0 {
		c.outputExtraHeaderData(exBuf, h)
		h.Size += exBuf.Len()
	}

	hdr := make([]byte, 10)
	encodeSyncSafeUint32(hdr[4:8], uint32(h.Size))
	copy(hdr[0:4], []byte(h.FrameID))
	flags := c.frameFlags.Encode(uint32(h.Flags))
	hdr[8] = byte(flags >> 8)
	hdr[9] = byte(flags)

	n, err := w.Write(hdr)
	if err != nil {
		return n, err
	}

	exBuf.SaveTo(w)
	n += exBuf.n

	ww.SaveTo(w)
	n += ww.n
	return n, ww.err
}

func (c *codec24) outputExtraHeaderData(ww *writer, h *FrameHeader) {
	if (h.Flags & FrameFlagCompressed) != 0 {
		h.Flags |= FrameFlagHasDataLength
	}

	if (h.Flags & FrameFlagHasGroupID) != 0 {
		if h.GroupID < 0x80 || h.GroupID > 0xf0 {
			ww.err = ErrInvalidGroupID
		}
		ww.StoreByte(h.GroupID)
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		if h.EncryptMethod < 0x80 || h.EncryptMethod > 0xf0 {
			ww.err = ErrInvalidEncryptMethod
		}
		ww.StoreByte(h.EncryptMethod)
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		b := make([]byte, 4)
		ww.err = encodeSyncSafeUint32(b, uint32(h.Size))
		ww.StoreBytes(b)
	}
}

func (c *codec24) outputStruct(ww *writer, p property, state *state) {
	if p.typ.Name() == "FrameHeader" {
		return
	}

	state.structStack.push(p.value)
	if state.structStack.depth() == 1 {
		state.fieldCount = p.typ.NumField()
	}

	for i, n := 0, p.typ.NumField(); i < n; i++ {
		if state.structStack.depth() == 1 {
			state.fieldIndex = i
		}

		field := p.typ.Field(i)

		fp := property{
			typ:   field.Type,
			value: p.value.Field(i),
			name:  field.Name,
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			c.outputUint8(ww, fp, state)

		case reflect.Uint16:
			c.outputUint16(ww, fp, state)

		case reflect.Uint32:
			c.outputUint32(ww, fp, state)

		case reflect.Uint64:
			c.outputUint64(ww, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.outputByteSlice(ww, fp, state)
			case reflect.Uint32:
				c.outputUint32Slice(ww, fp, state)
			case reflect.String:
				c.outputStringSlice(ww, fp, state)
			case reflect.Struct:
				c.outputStructSlice(ww, fp, state)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.outputString(ww, fp, state)

		case reflect.Struct:
			c.outputStruct(ww, fp, state)

		default:
			panic(errUnknownFieldType)
		}
	}

	state.structStack.pop()
}

func (c *codec24) outputUint8(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	value := uint8(p.value.Uint())

	if p.typ.Name() == "FrameType" {
		state.frameType = FrameType(value)
		return
	}

	bounds, hasBounds := c.bounds[p.name]

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		ww.err = bounds.err
		return
	}

	ww.StoreByte(value)
	if ww.err != nil {
		return
	}
}

func (c *codec24) outputUint16(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	v := uint16(p.value.Uint())

	switch p.name {
	case "BPM":
		if v > 2*0xff {
			ww.err = ErrInvalidBPM
			return
		}
		if v < 0xff {
			ww.StoreByte(uint8(v))
		} else {
			ww.StoreByte(0xff)
			ww.StoreByte(uint8(v - 0xff))
		}
	default:
		b := []byte{byte(v >> 8), byte(v)}
		ww.StoreBytes(b)
	}
}

func (c *codec24) outputUint32(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	v := uint32(p.value.Uint())
	b := []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}

	ww.StoreBytes(b)
}

func (c *codec24) outputUint64(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	v := p.value.Uint()

	switch p.name {
	case "Counter":
		b := make([]byte, 0, 4)
		for v != 0 {
			b = append(b, byte(v&0xff))
			v = v >> 8
		}
		for len(b) < 4 {
			b = append(b, 0)
		}
		for i := len(b) - 1; i >= 0; i-- {
			ww.StoreByte(b[i])
		}
	default:
		panic(errUnknownFieldType)
	}
}

func (c *codec24) outputUint32Slice(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	if p.name != "IndexOffsets" {
		panic(errUnknownFieldType)
	}

	sf := state.structStack.first()
	length := uint32(sf.FieldByName("IndexedDataLength").Uint())
	bits := uint32(sf.FieldByName("BitsPerIndex").Uint())

	n := p.value.Len()
	slice := p.value.Slice(0, n)

	switch bits {
	case 8:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 8) / length
			if frac >= (1 << 8) {
				frac = (1 << 8) - 1
			}
			ww.StoreByte(byte(frac))
		}

	case 16:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 16) / length
			if frac >= (1 << 16) {
				frac = (1 << 16) - 1
			}
			b := []byte{byte(frac >> 8), byte(frac)}
			ww.StoreBytes(b)
		}

	default:
		ww.err = ErrInvalidBits
	}
}

func (c *codec24) outputByteSlice(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	var b []byte
	reflect.ValueOf(&b).Elem().Set(p.value)
	ww.StoreBytes(b)
}

func (c *codec24) outputStringSlice(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	sf := state.structStack.first()
	enc := Encoding(sf.FieldByName("Encoding").Uint())

	var ss []string
	reflect.ValueOf(&ss).Elem().Set(p.value)
	ww.StoreStrings(ss, enc)
}

func (c *codec24) outputStructSlice(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	n := p.value.Len()
	slice := p.value.Slice(0, n)

	for i := 0; i < n; i++ {
		elem := slice.Index(i)

		ep := property{
			typ:   elem.Type(),
			value: elem,
			name:  fmt.Sprintf("%s[%d]", p.name, i),
		}

		c.outputStruct(ww, ep, state)
		if ww.err != nil {
			return
		}
	}
}

func (c *codec24) outputString(ww *writer, p property, state *state) {
	if ww.err != nil {
		return
	}

	v := p.value.String()

	switch p.name {
	case "FrameID":
		state.frameID = v
		return
	case "Language":
		ww.StoreFixedLengthString(v, 3, EncodingISO88591)
		return
	}

	var enc Encoding
	switch p.typ.Name() {
	case "WesternString":
		enc = EncodingISO88591
	default:
		sf := state.structStack.first()
		enc = Encoding(sf.FieldByName("Encoding").Uint())
	}

	// Always terminate strings unless they are the last struct field
	// of the root level struct.
	term := state.structStack.depth() > 1 || (state.fieldIndex != state.fieldCount-1)
	ww.StoreString(v, enc, term)
}
