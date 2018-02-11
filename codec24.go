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
			{1 << 6, uint32(FrameFlagHasGroupInfo)},
			{1 << 3, uint32(FrameFlagCompressed)},
			{1 << 2, uint32(FrameFlagEncrypted)},
			{1 << 1, uint32(FrameFlagUnsynchronized)},
			{1 << 0, uint32(FrameFlagHasDataLength)},
		},
		bounds: boundsMap{
			"Encoding":         {0, 3},
			"GroupID":          {0x80, 0xf0},
			"LyricContentType": {0, 8},
			"PictureType":      {0, 20},
			"TimeStampFormat":  {1, 2},
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

func (c *codec24) HeaderFlags() flagMap {
	return c.headerFlags
}

func (c *codec24) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	// Read the first 6 bytes of the extended header so we can see how big
	// the additional extended data is.
	buf := newBuffer()
	if buf.Read(r, 6); buf.err != nil {
		return buf.n, buf.err
	}

	// Read the size of the extended data.
	size, err := decodeSyncSafeUint32(buf.ConsumeBytes(4))
	if err != nil {
		return buf.n, err
	}

	// The number of extended flag bytes must be 1.
	if buf.ConsumeByte() != 1 {
		return buf.n, ErrInvalidHeader
	}

	// Read the extended flags field.
	exFlags := buf.ConsumeByte()
	if buf.err != nil {
		return buf.n, buf.err
	}

	// Read the rest of the extended header into the buffer.
	if buf.Read(r, int(size)-6); buf.err != nil {
		return buf.n, buf.err
	}

	if (exFlags & (1 << 6)) != 0 {
		t.Flags |= TagFlagIsUpdate
		if buf.ConsumeByte() != 0 || buf.err != nil {
			return buf.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 5)) != 0 {
		t.Flags |= TagFlagHasCRC
		data := buf.ConsumeBytes(6)
		if buf.err != nil || data[0] != 5 {
			return buf.n, ErrInvalidHeader
		}
		t.CRC, err = decodeSyncSafeUint32(data[1:])
		if err != nil {
			return buf.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 4)) != 0 {
		t.Flags |= TagFlagHasRestrictions
		data := buf.ConsumeBytes(2)
		if buf.err != nil || data[0] != 1 {
			return buf.n, ErrInvalidHeader
		}
		t.Restrictions = uint16(data[0])<<8 | uint16(data[1])
	}

	return buf.n, buf.err
}

func (c *codec24) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header data to see if it's
	// padding.
	buf := newBuffer()
	if buf.Read(r, 4); buf.err != nil {
		return buf.n, buf.err
	}
	hd := buf.ConsumeAll()
	if hd[0] == 0 && hd[1] == 0 && hd[2] == 0 && hd[3] == 0 {
		return buf.n, errPaddingEncountered
	}

	// Read the remaining 6 bytes of the header data.
	if buf.Read(r, 6); buf.err != nil {
		return buf.n, buf.err
	}
	hd = append(hd, buf.ConsumeAll()...)

	// Decode the frame's payload size.
	size, err := decodeSyncSafeUint32(hd[4:8])
	if err != nil {
		return buf.n, err
	}
	if size < 1 {
		return buf.n, ErrInvalidFrameHeader
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
	if buf.Read(r, header.Size); buf.err != nil {
		return buf.n, buf.err
	}

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (header.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		buf.Replace(removeUnsyncCodes(buf.buf))
	}

	// Scan extra header data indicated by the flags.
	if header.Flags != 0 {
		c.scanExtraHeaderData(buf, &header)
		if buf.err != nil {
			return buf.n, buf.err
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
	c.scanStruct(buf, p, &state)

	// Return the interpreted frame and header.
	if buf.err == nil {
		*f = p.value.Interface().(Frame)

		// The frame's first field is always the header. Copy into it.
		ht := reflect.ValueOf(*f).Elem()
		ht.Field(0).Set(reflect.ValueOf(header))
	}

	return buf.n, buf.err
}

func (c *codec24) scanExtraHeaderData(buf *buffer, h *FrameHeader) {
	// If the frame is compressed, it must include a data length indicator.
	if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
		buf.err = ErrInvalidFrameFlags
		return
	}

	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		gid := buf.ConsumeByte()
		if buf.err != nil || gid < 0x80 || gid > 0xf0 {
			buf.err = ErrInvalidFrameHeader
			return
		}
		h.GroupID = gid
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		em := buf.ConsumeByte()
		if buf.err != nil || em < 0x80 || em > 0xf0 {
			buf.err = ErrInvalidFrameHeader
			return
		}
		h.EncryptMethod = em
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		b := buf.ConsumeBytes(4)
		if buf.err != nil {
			buf.err = ErrInvalidFrameHeader
		}
		h.DataLength, buf.err = decodeSyncSafeUint32(b)
	}
}

var counter = 0

func (c *codec24) scanStruct(s *buffer, p property, state *state) {
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

		counter++
		if counter == 58 {
			foo := 0
			_ = foo
		}
		field := p.typ.Field(ii)

		fp := property{
			typ:   field.Type,
			value: p.value.Elem().Field(ii),
			name:  field.Name,
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
				c.scanStructSlice(s, fp, state)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.scanString(s, fp, state)

		case reflect.Struct:
			c.scanStruct(s, fp, state)

		default:
			panic(errUnknownFieldType)
		}
	}

	state.structStack.pop()
}

func (c *codec24) scanUint8(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	if p.typ.Name() == "FrameType" {
		state.frameType = c.frameTypes.LookupFrameType(state.frameID)
		p.value.SetUint(uint64(state.frameType))
		return
	}

	bounds, hasBounds := c.bounds[p.name]

	value := buf.ConsumeByte()
	if buf.err != nil {
		return
	}

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		buf.err = ErrInvalidFrame
		return
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint16(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	var value uint16
	switch p.name {
	case "BPM":
		value = uint16(buf.ConsumeByte())
		if value == 0xff {
			value += uint16(buf.ConsumeByte())
		}
	default:
		b := buf.ConsumeBytes(2)
		value = uint16(b[0])<<8 | uint16(b[1])
	}

	if buf.err != nil {
		return
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint32(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	b := buf.ConsumeBytes(4)

	var value uint64
	for _, bb := range b {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanUint64(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	var b []byte
	switch p.name {
	case "Counter":
		b = buf.ConsumeAll()
	default:
		panic(errUnknownFieldType)
	}

	if buf.err != nil {
		buf.err = ErrInvalidFrame
		return
	}

	var value uint64
	for _, bb := range b {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanByteSlice(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	b := buf.ConsumeAll()
	p.value.Set(reflect.ValueOf(b))
}

func (c *codec24) scanUint32Slice(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	if p.name != "IndexOffsets" {
		panic(errUnknownFieldType)
	}

	sf := state.structStack.first()
	length := uint32(sf.FieldByName("IndexedDataLength").Uint())
	bits := uint32(sf.FieldByName("BitsPerIndex").Uint())

	var offsets []uint32

	ff := buf.ConsumeAll()
	switch bits {
	case 8:
		offsets = make([]uint32, len(ff))
		for _, f := range ff {
			frac := uint32(f)
			offset := (frac*length + (1 << 7)) << 8
			if offset > length {
				offset = length
			}
			offsets = append(offsets, offset)
		}

	case 16:
		offsets = make([]uint32, len(ff)/2)
		for ii := 0; ii < len(ff); ii += 2 {
			frac := uint32(ff[ii])<<8 | uint32(ff[ii+1])
			offset := (frac*length + (1 << 15)) << 16
			if offset > length {
				offset = length
			}
			offsets = append(offsets, offset)
		}

	default:
		buf.err = ErrInvalidFrame
		return
	}

	p.value.Set(reflect.ValueOf(offsets))
}

func (c *codec24) scanStringSlice(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	sf := state.structStack.first()
	enc := Encoding(sf.FieldByName("Encoding").Uint())
	ss := buf.ConsumeStrings(enc)
	if buf.err != nil {
		return
	}
	p.value.Set(reflect.ValueOf(ss))
}

func (c *codec24) scanStructSlice(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	elems := make([]reflect.Value, 0)
	for i := 0; buf.Len() > 0; i++ {
		etyp := p.typ.Elem()
		ep := property{
			typ:   etyp,
			value: reflect.New(etyp),
			name:  fmt.Sprintf("%s[%d]", p.name, i),
		}

		c.scanStruct(buf, ep, state)
		if buf.err != nil {
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

func (c *codec24) scanString(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	if p.name == "FrameID" {
		p.value.SetString(string(state.frameID))
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

	var str string
	switch p.name {
	case "Language":
		str = buf.ConsumeFixedLengthString(3, EncodingISO88591)
	default:
		str = buf.ConsumeNextString(enc)
	}

	if buf.err != nil {
		return
	}

	p.value.SetString(str)
}

func (c *codec24) EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error) {
	buf := newBuffer()

	p := property{
		typ:   reflect.TypeOf(f).Elem(),
		value: reflect.ValueOf(f).Elem(),
		name:  "",
	}
	state := state{}

	c.outputStruct(buf, p, &state)
	if buf.err != nil {
		return buf.n, buf.err
	}

	h := HeaderOf(f)
	h.FrameID = c.frameTypes.LookupFrameID(state.frameType)
	h.Size = buf.Len()

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

	buf.Write(w)
	n += buf.n
	return n, buf.err
}

func (c *codec24) outputStruct(buf *buffer, p property, state *state) {
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
			c.outputUint8(buf, fp, state)

		case reflect.Uint16:
			c.outputUint16(buf, fp, state)

		case reflect.Uint32:
			c.outputUint32(buf, fp, state)

		case reflect.Uint64:
			c.outputUint64(buf, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.outputByteSlice(buf, fp, state)
			case reflect.Uint32:
				c.outputUint32Slice(buf, fp, state)
			case reflect.String:
				c.outputStringSlice(buf, fp, state)
			case reflect.Struct:
				c.outputStructSlice(buf, fp, state)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.outputString(buf, fp, state)

		case reflect.Struct:
			c.outputStruct(buf, fp, state)

		default:
			panic(errUnknownFieldType)
		}
	}

	state.structStack.pop()
}

func (c *codec24) outputUint8(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	value := uint8(p.value.Uint())

	if p.typ.Name() == "FrameType" {
		state.frameType = FrameType(value)
		return
	}

	bounds, hasBounds := c.bounds[p.name]

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		buf.err = ErrInvalidFrame
		return
	}

	buf.AddByte(value)
	if buf.err != nil {
		return
	}
}

func (c *codec24) outputUint16(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	v := uint16(p.value.Uint())

	switch p.name {
	case "BPM":
		if v > 2*0xff {
			buf.err = ErrInvalidFrame
			return
		}
		if v < 0xff {
			buf.AddByte(uint8(v))
		} else {
			buf.AddByte(0xff)
			buf.AddByte(uint8(v - 0xff))
		}
	default:
		b := []byte{byte(v >> 8), byte(v)}
		buf.AddBytes(b)
	}
}

func (c *codec24) outputUint32(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	v := uint32(p.value.Uint())
	b := []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}

	buf.AddBytes(b)
}

func (c *codec24) outputUint64(buf *buffer, p property, state *state) {
	if buf.err != nil {
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
			buf.AddByte(b[i])
		}
	default:
		panic(errUnknownFieldType)
	}
}

func (c *codec24) outputUint32Slice(buf *buffer, p property, state *state) {
	if buf.err != nil {
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
			buf.AddByte(byte(frac))
		}

	case 16:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 16) / length
			if frac >= (1 << 16) {
				frac = (1 << 16) - 1
			}
			b := []byte{byte(frac >> 8), byte(frac)}
			buf.AddBytes(b)
		}

	default:
		buf.err = ErrInvalidFrame
	}
}

func (c *codec24) outputByteSlice(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	var b []byte
	reflect.ValueOf(&b).Elem().Set(p.value)
	buf.AddBytes(b)
}

func (c *codec24) outputStringSlice(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	sf := state.structStack.first()
	enc := Encoding(sf.FieldByName("Encoding").Uint())

	var ss []string
	reflect.ValueOf(&ss).Elem().Set(p.value)
	buf.AddStrings(ss, enc)
}

func (c *codec24) outputStructSlice(buf *buffer, p property, state *state) {
	if buf.err != nil {
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

		c.outputStruct(buf, ep, state)
		if buf.err != nil {
			return
		}
	}
}

func (c *codec24) outputString(buf *buffer, p property, state *state) {
	if buf.err != nil {
		return
	}

	v := p.value.String()

	if p.name == "FrameID" {
		state.frameID = v
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

	switch p.name {
	case "Language":
		buf.AddFixedLengthString(v, 3, enc)
	default:
		// Always terminate strings unless they are the last struct field
		// of the root level struct.
		term := state.structStack.depth() > 1 || (state.fieldIndex != state.fieldCount-1)
		buf.AddString(v, enc, term)
	}
}
