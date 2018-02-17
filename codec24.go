package id3

import (
	"fmt"
	"hash/crc32"
	"io"
	"reflect"
)

//
// codec24
//

type codec24 struct {
	headerFlags   flagMap
	headerExFlags flagMap
	frameFlags    flagMap
	bounds        boundsMap
	frameTypes    *frameTypeMap
}

func newCodec24() *codec24 {
	return &codec24{
		headerFlags: flagMap{
			{1 << 7, uint32(TagFlagUnsync)},
			{1 << 6, uint32(TagFlagExtended)},
			{1 << 5, uint32(TagFlagExperimental)},
			{1 << 4, uint32(TagFlagFooter)},
		},
		headerExFlags: flagMap{
			{1 << 6, uint32(TagFlagIsUpdate)},
			{1 << 5, uint32(TagFlagHasCRC)},
			{1 << 4, uint32(TagFlagHasRestrictions)},
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

// Decode decodes an ID3 v2.4 tag, starting from the fifth byte of the
// tag.  The result is placed into the Tag t.
func (c *codec24) Decode(t *Tag, r *reader) (int, error) {
	// Read the remaining six bytes of the tag header.
	if r.Load(6); r.err != nil {
		return r.n, r.err
	}
	hdr := r.ConsumeBytes(6)
	if hdr[0] != 0 {
		return r.n, ErrInvalidTag
	}

	// Process tag header flags.
	flags := uint32(hdr[1])
	t.Flags = TagFlags(c.headerFlags.Decode(flags))

	// Process tag size.
	size, err := decodeSyncSafeUint32(hdr[2:6])
	if err != nil {
		return r.n, err
	}
	t.Size = int(size)

	// Load the rest of the tag into the reader's buffer.
	if r.Load(t.Size); r.err != nil {
		return r.n, r.err
	}

	// Remove unsync codes.
	if (t.Flags & TagFlagUnsync) != 0 {
		newBuf := removeUnsyncCodes(r.ConsumeAll())
		r.ReplaceBuffer(newBuf)
	}

	// Decode the extended header.
	if (t.Flags & TagFlagExtended) != 0 {
		exSize, err := decodeSyncSafeUint32(r.ConsumeBytes(4))
		if err != nil {
			return r.n, err
		}

		if exFlagsSize := r.ConsumeByte(); exFlagsSize != 1 {
			return r.n, ErrInvalidHeader
		}

		// Decode the extended header flags.
		exFlags := r.ConsumeByte()
		t.Flags = TagFlags(uint32(t.Flags) | c.headerExFlags.Decode(uint32(exFlags)))

		// Consume the rest of the extended header data.
		exBytesConsumed := 6

		if (t.Flags & TagFlagIsUpdate) != 0 {
			r.ConsumeByte()
			exBytesConsumed++
		}

		if (t.Flags & TagFlagHasCRC) != 0 {
			data := r.ConsumeBytes(6)
			if data[0] != 5 {
				return r.n, ErrInvalidHeader
			}
			t.CRC, err = decodeSyncSafeUint32(data[1:6])
			if err != nil {
				return r.n, ErrInvalidHeader
			}
			exBytesConsumed += 6
		}

		if (t.Flags & TagFlagHasRestrictions) != 0 {
			if r.ConsumeByte() != 1 {
				return r.n, ErrInvalidHeader
			}
			t.Restrictions = r.ConsumeByte()
			exBytesConsumed += 2
		}

		// Consume and ignore any remaining bytes in the extended header.
		if exBytesConsumed < int(exSize) {
			r.ConsumeBytes(int(exSize) - exBytesConsumed)
		}

		if r.err != nil {
			return r.n, r.err
		}
	}

	// Validate the CRC.
	if (t.Flags & TagFlagHasCRC) != 0 {
		crc := crc32.ChecksumIEEE(r.Bytes())
		if crc != t.CRC {
			return r.n, ErrFailedCRC
		}
	}

	// Decode the tag's frames until tag data is exhausted or padding is
	// encountered.
	for r.Len() > 0 {
		var f Frame
		_, err = c.decodeFrame(t, &f, r)

		if err == errPaddingEncountered {
			t.Padding = r.Len() + 4
			r.ConsumeAll()
			break
		}

		if err != nil {
			return r.n, err
		}

		t.Frames = append(t.Frames, f)
	}

	return r.n, nil
}

func (c *codec24) decodeFrame(t *Tag, f *Frame, r *reader) (int, error) {
	// Read the first four bytes of the frame header data to see if it's
	// padding.
	id := r.ConsumeBytes(4)
	if r.err != nil {
		return r.n, r.err
	}
	if id[0] == 0 && id[1] == 0 && id[2] == 0 && id[3] == 0 {
		return r.n, errPaddingEncountered
	}

	// Read the remaining 6 bytes of the header data into a buffer.
	hd := r.ConsumeBytes(6)
	if r.err != nil {
		return r.n, r.err
	}

	// Decode the frame's payload size.
	size, err := decodeSyncSafeUint32(hd[0:4])
	if err != nil {
		return r.n, err
	}
	if size < 1 {
		return r.n, ErrInvalidFrameHeader
	}

	// Decode the frame flags.
	flags := c.frameFlags.Decode(uint32(hd[4])<<8 | uint32(hd[5]))

	// Create the frame header structure.
	h := FrameHeader{
		FrameID: string(id),
		Size:    int(size),
		Flags:   FrameFlags(flags),
	}

	// Read the rest of the frame into a new reader.
	r = r.ConsumeIntoNewReader(h.Size)

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (h.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		b := removeUnsyncCodes(r.ConsumeAll())
		r.ReplaceBuffer(b)
	}

	// Scan extra header data.
	if h.Flags != 0 {

		// If the frame is compressed, it must include a data length indicator.
		if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
			return r.n, ErrInvalidFrameFlags
		}

		if (h.Flags & FrameFlagHasGroupID) != 0 {
			gid := r.ConsumeByte()
			if r.err != nil {
				return r.n, r.err
			}
			if gid < 0x80 || gid > 0xf0 {
				return r.n, ErrInvalidGroupID
			}
			h.GroupID = gid
		}

		if (h.Flags & FrameFlagEncrypted) != 0 {
			em := r.ConsumeByte()
			if r.err != nil {
				return r.n, r.err
			}
			if em < 0x80 || em > 0xf0 {
				return r.n, ErrInvalidEncryptMethod
			}
			h.EncryptMethod = em
		}

		if (h.Flags & FrameFlagHasDataLength) != 0 {
			b := r.ConsumeBytes(4)
			if r.err != nil {
				return r.n, ErrInvalidFrameHeader
			}
			h.DataLength, err = decodeSyncSafeUint32(b)
			if err != nil {
				return r.n, err
			}
		}
	}

	// Initialize the frame payload scan state.
	state := state{
		frameID: h.FrameID,
	}

	// Use reflection to interpret the payload's contents.
	typ := c.frameTypes.LookupReflectType(h.FrameID)
	p := property{
		typ:   typ,
		value: reflect.New(typ),
		name:  "",
	}
	c.scanStruct(r, p, &state)

	if r.err != nil {
		return r.n, r.err
	}

	// Use reflection to access the decoded frame payload.
	*f = p.value.Interface().(Frame)

	// The frame's first field is always the header. Use reflection to copy to
	// it.
	ht := reflect.ValueOf(*f).Elem()
	ht.Field(0).Set(reflect.ValueOf(h))

	return r.n, nil
}

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

func (c *codec24) Encode(t *Tag, w *writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec24) EncodeExtendedHeader(t *Tag, w io.Writer) (int, error) {
	ww := newWriter(w)

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
	ww := newWriter(w)

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

	exBuf := newWriter(w)
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

	exBuf.Save()
	n += exBuf.n

	ww.Save()
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
