package id3

import (
	"io"
	"reflect"
)

//
// codec24
//

type codec24 struct {
	payloadTypes typeMap   // table of all frame payload types
	headerFlags  flagMap   // tag header flags
	frameFlags   flagMap   // frame header flags
	bounds       boundsMap // integer field bounds, by type
}

func newCodec24() *codec24 {
	return &codec24{
		payloadTypes: newTypeMap("v24"),
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
			"Encoding":    {0, 3},
			"GroupSymbol": {0x80, 0xf0},
			"PictureType": {0, 20},
		},
	}
}

func (c *codec24) HeaderFlags() flagMap {
	return c.headerFlags
}

func (c *codec24) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	// Read the first 6 bytes of the extended header so we can see how big
	// the addition extended data is.
	var s scanner
	if s.Read(r, 6); s.err != nil {
		return s.n, s.err
	}

	// Read the size of the extended data.
	size, err := decodeSyncSafeUint32(s.ConsumeBytes(4))
	if err != nil {
		return s.n, err
	}

	// The number of extended flag bytes must be 1.
	if s.ConsumeByte() != 1 {
		return s.n, ErrInvalidHeader
	}

	// Read the extended flags field.
	exFlags := s.ConsumeByte()
	if s.err != nil {
		return s.n, s.err
	}

	// Read the rest of the extended header into the buffer.
	if s.Read(r, int(size)-6); s.err != nil {
		return s.n, s.err
	}

	// Scan extended data fields indicated by the flags.
	if (exFlags & (1 << 6)) != 0 {
		t.Flags |= TagFlagIsUpdate
		if s.ConsumeByte() != 0 || s.err != nil {
			return s.n, ErrInvalidHeader
		}
	}
	if (exFlags & (1 << 5)) != 0 {
		t.Flags |= TagFlagHasCRC
		data := s.ConsumeBytes(6)
		if s.err != nil || data[0] != 5 {
			return s.n, ErrInvalidHeader
		}
		t.CRC, err = decodeSyncSafeUint32(data[1:])
		if err != nil {
			return s.n, ErrInvalidHeader
		}
	}
	if (exFlags & (1 << 4)) != 0 {
		t.Flags |= TagFlagHasRestrictions
		data := s.ConsumeBytes(2)
		if s.err != nil || data[0] != 1 {
			return s.n, ErrInvalidHeader
		}
		t.Restrictions = uint16(data[0])<<8 | uint16(data[1])
	}

	return s.n, s.err
}

func (c *codec24) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header to see if it's padding.
	var s scanner
	if s.Read(r, 4); s.err != nil {
		return s.n, s.err
	}

	hdr := s.ConsumeAll()
	if hdr[0] == 0 && hdr[1] == 0 && hdr[2] == 0 && hdr[3] == 0 {
		return s.n, errPaddingEncountered
	}
	f.Header.ID = string(hdr[0:4])

	// Read the rest of the header.
	if s.Read(r, 6); s.err != nil {
		return s.n, s.err
	}
	hdr = append(hdr, s.ConsumeAll()...)

	// Process the frame's size.
	size, err := decodeSyncSafeUint32(hdr[4:8])
	if err != nil {
		return s.n, err
	}
	if size < 1 {
		return s.n, ErrInvalidFrameHeader
	}
	f.Header.Size = int(size)

	// Process the flags.
	flags := uint32(hdr[8])<<8 | uint32(hdr[9])
	f.Header.Flags = FrameFlags(c.frameFlags.Decode(flags))

	// Read the rest of the frame into a buffer.
	if s.Read(r, f.Header.Size); s.err != nil {
		return s.n, s.err
	}

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (f.Header.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		s.Replace(removeUnsyncCodes(s.buf))
	}

	// Scan extra header data indicated by flags.
	if f.Header.Flags != 0 {
		c.scanExtraHeaderData(&s, &f.Header)
		if s.err != nil {
			return s.n, s.err
		}
	}

	// Select a frame payload type based on the ID.
	typ := c.payloadTypes.Lookup24(f.Header.ID)

	// Instantiate a new frame payload using reflection.
	v := reflect.New(typ)

	// Use the reflection type of the payload to process the frame's data.
	enc := EncodingISO88591
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := v.Elem().Field(i)

		switch field.Type.Kind() {
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.scanByteSlice(&s, fieldValue)
			case reflect.String:
				c.scanStringSlice(&s, fieldValue, enc)
			default:
				s.err = ErrUnknownFieldType
			}

		case reflect.String:
			c.scanString(&s, field, fieldValue, enc)

		case reflect.Uint8:
			v := c.scanUint8(&s, field, fieldValue)
			if field.Type.Name() == "Encoding" {
				enc = Encoding(v)
			}

		case reflect.Uint16:
			// skip (this is the frameId)

		case reflect.Uint64:
			c.scanUint64(&s, field, fieldValue)

		default:
			s.err = ErrUnknownFieldType
		}
	}

	if s.err == nil {
		f.Payload = v.Interface().(FramePayload)
	}

	return s.n, s.err
}

func (c *codec24) scanExtraHeaderData(s *scanner, h *FrameHeader) {
	// If the frame is compressed, it must include a data length indicator.
	if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
		s.err = ErrInvalidFrameFlags
		return
	}

	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		gid := s.ConsumeByte()
		if s.err != nil || gid < 0x80 || gid > 0xf0 {
			s.err = ErrInvalidFrameHeader
			return
		}
		h.GroupID = GroupSymbol(gid)
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		em := s.ConsumeByte()
		if s.err != nil || em < 0x80 || em > 0xf0 {
			s.err = ErrInvalidFrameHeader
			return
		}
		h.EncryptMethod = em
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		b := s.ConsumeBytes(4)
		if s.err != nil {
			s.err = ErrInvalidFrameHeader
		}
		h.DataLength, s.err = decodeSyncSafeUint32(b)
	}
}

func (c *codec24) scanString(s *scanner, field reflect.StructField, v reflect.Value, enc Encoding) string {
	var str string
	if s.err != nil {
		return str
	}

	tags := getTags(field, "id3")
	if tags.Lookup("iso88519") {
		enc = EncodingISO88591
	}

	if tags.Lookup("lang") {
		str = s.ConsumeFixedLenString(3, EncodingISO88591)
	} else {
		str = s.ConsumeNextString(enc)
	}
	if s.err != nil {
		return str
	}

	v.SetString(str)
	return str
}

func (c *codec24) scanStringSlice(s *scanner, v reflect.Value, enc Encoding) []string {
	var ss []string
	if s.err != nil {
		return ss
	}

	ss = s.ConsumeStrings(enc)
	if s.err != nil {
		return ss
	}
	v.Set(reflect.ValueOf(ss))
	return ss
}

func (c *codec24) scanByteSlice(s *scanner, v reflect.Value) []byte {
	var b []byte
	if s.err != nil {
		return b
	}

	b = s.ConsumeAll()
	v.Set(reflect.ValueOf(b))
	return b
}

func (c *codec24) scanUint8(s *scanner, field reflect.StructField, v reflect.Value) uint8 {
	var value uint8
	if s.err != nil {
		return value
	}

	b, hasBounds := c.bounds[field.Type.Name()]

	value = s.ConsumeByte()
	if s.err != nil || (hasBounds && (value < uint8(b.min) || value > uint8(b.max))) {
		s.err = ErrInvalidFrame
		return value
	}

	v.SetUint(uint64(value))
	return value
}

func (c *codec24) scanUint64(s *scanner, field reflect.StructField, v reflect.Value) uint64 {
	var value uint64
	if s.err != nil {
		return value
	}

	tags := getTags(field, "id3")

	var buf []byte
	if tags.Lookup("counter") {
		buf = s.ConsumeAll()
	} else {
		buf = s.ConsumeBytes(8)
	}

	if s.err != nil {
		s.err = ErrInvalidFrame
		return value
	}

	for _, b := range buf {
		value = (value << 8) | uint64(b)
	}

	v.SetUint(value)
	return value
}

func (c *codec24) EncodeFrame(t *Tag, f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
