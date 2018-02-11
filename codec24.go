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
}

// The state structure keeps track of persistent state required while
// decoding a single frame.
type state struct {
	frameID    FrameID
	frameType  FrameType
	encoding   Encoding // used by text frames
	bits       uint8    // used by ASPI frame
	fieldIndex int
	fieldCount int
}

func (c *codec24) HeaderFlags() flagMap {
	return c.headerFlags
}

func (c *codec24) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	// Read the first 6 bytes of the extended header so we can see how big
	// the additional extended data is.
	var i ibuf
	if i.Read(r, 6); i.err != nil {
		return i.n, i.err
	}

	// Read the size of the extended data.
	size, err := decodeSyncSafeUint32(i.ConsumeBytes(4))
	if err != nil {
		return i.n, err
	}

	// The number of extended flag bytes must be 1.
	if i.ConsumeByte() != 1 {
		return i.n, ErrInvalidHeader
	}

	// Read the extended flags field.
	exFlags := i.ConsumeByte()
	if i.err != nil {
		return i.n, i.err
	}

	// Read the rest of the extended header into the buffer.
	if i.Read(r, int(size)-6); i.err != nil {
		return i.n, i.err
	}

	if (exFlags & (1 << 6)) != 0 {
		t.Flags |= TagFlagIsUpdate
		if i.ConsumeByte() != 0 || i.err != nil {
			return i.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 5)) != 0 {
		t.Flags |= TagFlagHasCRC
		data := i.ConsumeBytes(6)
		if i.err != nil || data[0] != 5 {
			return i.n, ErrInvalidHeader
		}
		t.CRC, err = decodeSyncSafeUint32(data[1:])
		if err != nil {
			return i.n, ErrInvalidHeader
		}
	}

	if (exFlags & (1 << 4)) != 0 {
		t.Flags |= TagFlagHasRestrictions
		data := i.ConsumeBytes(2)
		if i.err != nil || data[0] != 1 {
			return i.n, ErrInvalidHeader
		}
		t.Restrictions = uint16(data[0])<<8 | uint16(data[1])
	}

	return i.n, i.err
}

func (c *codec24) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header data to see if it'i
	// padding.
	var i ibuf
	if i.Read(r, 4); i.err != nil {
		return i.n, i.err
	}
	hd := i.ConsumeAll()
	if hd[0] == 0 && hd[1] == 0 && hd[2] == 0 && hd[3] == 0 {
		return i.n, errPaddingEncountered
	}

	// Read the remaining 6 bytes of the header data.
	if i.Read(r, 6); i.err != nil {
		return i.n, i.err
	}
	hd = append(hd, i.ConsumeAll()...)

	// Decode the frame's payload size.
	size, err := decodeSyncSafeUint32(hd[4:8])
	if err != nil {
		return i.n, err
	}
	if size < 1 {
		return i.n, ErrInvalidFrameHeader
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
	if i.Read(r, header.Size); i.err != nil {
		return i.n, i.err
	}

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (header.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		i.Replace(removeUnsyncCodes(i.buf))
	}

	// Scan extra header data indicated by the flags.
	if header.Flags != 0 {
		c.scanExtraHeaderData(&i, &header)
		if i.err != nil {
			return i.n, i.err
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
	c.scanStruct(&i, p, &state, 0)

	// Return the interpreted frame and header.
	if i.err == nil {
		*f = p.value.Interface().(Frame)

		// The frame's first field is always the header. Copy into it.
		ht := reflect.ValueOf(*f).Elem()
		ht.Field(0).Set(reflect.ValueOf(header))
	}

	return i.n, i.err
}

func (c *codec24) scanExtraHeaderData(i *ibuf, h *FrameHeader) {
	// If the frame is compressed, it must include a data length indicator.
	if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
		i.err = ErrInvalidFrameFlags
		return
	}

	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		gid := i.ConsumeByte()
		if i.err != nil || gid < 0x80 || gid > 0xf0 {
			i.err = ErrInvalidFrameHeader
			return
		}
		h.GroupID = GroupSymbol(gid)
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		em := i.ConsumeByte()
		if i.err != nil || em < 0x80 || em > 0xf0 {
			i.err = ErrInvalidFrameHeader
			return
		}
		h.EncryptMethod = em
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		b := i.ConsumeBytes(4)
		if i.err != nil {
			i.err = ErrInvalidFrameHeader
		}
		h.DataLength, i.err = decodeSyncSafeUint32(b)
	}
}

func (c *codec24) scanStruct(s *ibuf, p property, state *state, depth int) {
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
			case reflect.Float32:
				c.scanFloat32Slice(s, fp, state)
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

func (c *codec24) scanUint8(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	if p.typ.Name() == "FrameType" {
		state.frameType = c.frameTypes.LookupFrameType(state.frameID)
		p.value.SetUint(uint64(state.frameType))
		return
	}

	b, hasBounds := c.bounds[p.typ.Name()]

	value := i.ConsumeByte()
	if i.err != nil {
		return
	}

	if hasBounds && (value < uint8(b.min) || value > uint8(b.max)) {
		i.err = ErrInvalidFrame
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

func (c *codec24) scanUint16(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	var value uint16
	switch p.typ.Name() {
	case "Tempo":
		value = uint16(i.ConsumeByte())
		if value == 0xff {
			value += uint16(i.ConsumeByte())
		}
	default:
		b := i.ConsumeBytes(2)
		value = uint16(b[0])<<8 | uint16(b[1])
	}

	if i.err != nil {
		return
	}

	p.value.SetUint(uint64(value))
}

func (c *codec24) scanUint32(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	buf := i.ConsumeBytes(4)

	var value uint64
	for _, b := range buf {
		value = (value << 8) | uint64(b)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanUint64(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	var buf []byte
	switch p.typ.Name() {
	case "Counter":
		buf = i.ConsumeAll()
	default:
		panic(errUnknownFieldType)
	}

	if i.err != nil {
		i.err = ErrInvalidFrame
		return
	}

	var value uint64
	for _, b := range buf {
		value = (value << 8) | uint64(b)
	}

	p.value.SetUint(value)
}

func (c *codec24) scanByteSlice(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	b := i.ConsumeAll()
	p.value.Set(reflect.ValueOf(b))
}

func (c *codec24) scanFloat32Slice(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	if p.typ.Elem().Name() != "Fraction" {
		panic(errUnknownFieldType)
	}

	if state.bits != 8 && state.bits != 16 {
		i.err = ErrInvalidFrame
		return
	}

	var indexes []Fraction

	ff := i.ConsumeAll()
	switch state.bits {
	case 8:
		indexes = make([]Fraction, len(ff))
		for _, b := range ff {
			v := uint32(b)
			indexes = append(indexes, Fraction(v)/Fraction(1<<8))
		}
	case 16:
		indexes = make([]Fraction, len(ff)/2)
		for ii := 0; ii < len(ff); ii += 2 {
			v := uint32(ff[ii])<<8 | uint32(ff[ii])
			indexes = append(indexes, Fraction(v)/Fraction(1<<16))
		}
	}

	p.value.Set(reflect.ValueOf(indexes))
}

func (c *codec24) scanStringSlice(i *ibuf, p property, state *state) {
	if i.err != nil {
		return
	}

	ss := i.ConsumeStrings(state.encoding)
	if i.err != nil {
		return
	}
	p.value.Set(reflect.ValueOf(ss))
}

func (c *codec24) scanStructSlice(i *ibuf, p property, state *state, depth int) {
	if i.err != nil {
		return
	}

	elems := make([]reflect.Value, 0)
	for i.Len() > 0 {
		etyp := p.typ.Elem()
		ep := property{
			typ:   etyp,
			value: reflect.New(etyp),
		}

		c.scanStruct(i, ep, state, depth+1)
		if i.err != nil {
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

func (c *codec24) scanString(i *ibuf, p property, state *state) {
	if i.err != nil {
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
		str = i.ConsumeFixedLengthString(3, EncodingISO88591)
	default:
		str = i.ConsumeNextString(enc)
	}

	if i.err != nil {
		return
	}

	p.value.SetString(str)
}

func (c *codec24) EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error) {
	o := newOutput()

	p := property{
		typ:   reflect.TypeOf(f).Elem(),
		value: reflect.ValueOf(f).Elem(),
	}
	state := state{
		encoding: EncodingISO88591,
	}

	c.outputStruct(o, p, &state, 0)
	if o.err != nil {
		return o.n, o.err
	}

	h := HeaderOf(f)
	h.FrameID = c.frameTypes.LookupFrameID(state.frameType)
	h.Size = o.Len()

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

	nn, err := w.Write(o.Bytes())
	n += nn
	return n, err
}

func (c *codec24) outputStruct(o *obuf, p property, state *state, depth int) {
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
			value: p.value.Field(i),
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			c.outputUint8(o, fp, state)

		case reflect.Uint16:
			c.outputUint16(o, fp, state)

		case reflect.Uint32:
			c.outputUint32(o, fp, state)

		case reflect.Uint64:
			c.outputUint64(o, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.outputByteSlice(o, fp, state)
			case reflect.Float32:
				c.outputFloat32Slice(o, fp, state)
			case reflect.String:
				c.outputStringSlice(o, fp, state)
			case reflect.Struct:
				c.outputStructSlice(o, fp, state, depth+1)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			c.outputString(o, fp, state, depth)

		case reflect.Struct:
			c.outputStruct(o, fp, state, depth+1)

		default:
			panic(errUnknownFieldType)
		}
	}
}

func (c *codec24) outputUint8(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	value := uint8(p.value.Uint())

	if p.typ.Name() == "FrameType" {
		state.frameType = FrameType(value)
		state.frameID = c.frameTypes.LookupFrameID(state.frameType)
		return
	}

	b, hasBounds := c.bounds[p.typ.Name()]

	if hasBounds && (value < uint8(b.min) || value > uint8(b.max)) {
		o.err = ErrInvalidFrame
		return
	}

	o.WriteByte(value)
	if o.err != nil {
		return
	}

	switch p.typ.Name() {
	case "Encoding":
		state.encoding = Encoding(value)
	case "Bits":
		state.bits = value
	}
}

func (c *codec24) outputUint16(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	v := uint16(p.value.Uint())

	switch p.typ.Name() {
	case "Tempo":
		if v > 2*0xff {
			o.err = ErrInvalidFrame
			return
		}
		if v < 0xff {
			o.WriteByte(uint8(v))
		} else {
			o.WriteByte(0xff)
			o.WriteByte(uint8(v - 0xff))
		}
	default:
		b := []byte{byte(v >> 8), byte(v)}
		o.WriteBytes(b)
	}
}

func (c *codec24) outputUint32(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	v := uint32(p.value.Uint())
	b := []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
	o.WriteBytes(b)
}

func (c *codec24) outputUint64(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	v := p.value.Uint()

	switch p.typ.Name() {
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
			o.WriteByte(b[i])
		}
	default:
		panic(errUnknownFieldType)
	}
}

func (c *codec24) outputFloat32Slice(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	if p.typ.Elem().Name() != "Fraction" {
		panic(errUnknownFieldType)
	}

	n := p.value.Len()
	sl := p.value.Slice(0, n)

	switch state.bits {
	case 8:
		for i := 0; i < n; i++ {
			o.WriteByte(byte(sl.Index(i).Float() * float64(1<<8)))
		}

	case 16:
		for i := 0; i < n; i++ {
			v := uint32(sl.Index(i).Float() * float64(1<<16))
			b := []byte{uint8(v >> 8), byte(uint8(v))}
			o.WriteBytes(b)
		}

	default:
		o.err = ErrInvalidFrame
	}
}

func (c *codec24) outputByteSlice(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	var b []byte
	reflect.ValueOf(&b).Elem().Set(p.value)
	o.WriteBytes(b)
}

func (c *codec24) outputStringSlice(o *obuf, p property, state *state) {
	if o.err != nil {
		return
	}

	var ss []string
	reflect.ValueOf(&ss).Elem().Set(p.value)
	o.err = o.WriteStrings(ss, state.encoding)
}

func (c *codec24) outputStructSlice(o *obuf, p property, state *state, depth int) {
	if o.err != nil {
		return
	}

	n := p.value.Len()
	sl := p.value.Slice(0, n)

	for i := 0; i < n; i++ {
		elem := sl.Index(i)

		ep := property{
			typ:   elem.Type(),
			value: elem,
		}

		c.outputStruct(o, ep, state, depth+1)
		if o.err != nil {
			return
		}
	}
}

func (c *codec24) outputString(o *obuf, p property, state *state, depth int) {
	if o.err != nil {
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
		o.WriteFixedLengthString(v, 3, enc)
	default:
		// Always terminate strings unless they are the last struct field
		// of the root level struct.
		term := depth > 0 || (state.fieldIndex != state.fieldCount-1)
		o.WriteString(v, enc, term)
	}
}
