package id3

import (
	"io"
	"reflect"
)

//
// codec24
//

type codec24 struct {
	payloadTypes typeMap // table of all frame payload types
	headerFlags  flagMap // tag header flags
	frameFlags   flagMap // frame header flags
	buf          []byte  // (de)serialization buffer
	n            int     // total bytes (de)serialized
	err          error   // first error encountered when (de)serializing
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
	}
}

func (c *codec24) HeaderFlags() flagMap {
	return c.headerFlags
}

func (c *codec24) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	// Read the first 6 bytes of the extended header so we can see how big
	// the addition extended data is.
	c.buf = make([]byte, 6)
	c.n, c.err = io.ReadFull(r, c.buf)
	if c.err != nil {
		return c.n, c.err
	}

	// Read the size of the extended data.
	var size uint32
	size, c.err = decodeSyncSafeUint32(c.consumeBytes(4))
	if c.err != nil {
		return c.n, c.err
	}

	// The number of extended flag bytes must be 1.
	if c.consumeByte() != 1 {
		return c.n, ErrInvalidHeader
	}

	// Read the extended flags field.
	exFlags := c.consumeByte()
	if c.err != nil {
		return c.n, c.err
	}

	// Read the rest of the extended header into the buffer.
	var n int
	c.buf = make([]byte, size-6)
	n, c.err = io.ReadFull(r, c.buf)
	c.n += n
	if c.err != nil {
		return c.n, c.err
	}

	// Scan extended data fields indicated by the flags.
	if (exFlags & (1 << 6)) != 0 {
		t.Flags |= TagFlagIsUpdate
		if c.consumeByte() != 0 || c.err != nil {
			return c.n, ErrInvalidHeader
		}
	}
	if (exFlags & (1 << 5)) != 0 {
		t.Flags |= TagFlagHasCRC
		data := c.consumeBytes(6)
		if c.err != nil || data[0] != 5 {
			return c.n, ErrInvalidHeader
		}
		t.CRC, c.err = decodeSyncSafeUint32(data[1:])
		if c.err != nil {
			return c.n, ErrInvalidHeader
		}
	}
	if (exFlags & (1 << 4)) != 0 {
		t.Flags |= TagFlagHasRestrictions
		data := c.consumeBytes(2)
		if c.err != nil || data[0] != 1 {
			return c.n, ErrInvalidHeader
		}
		t.Restrictions = uint16(data[0])<<8 | uint16(data[1])
	}
	return c.n, c.err
}

func (c *codec24) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header to see if it's padding.
	hdr := make([]byte, 10)
	nn, err := io.ReadFull(r, hdr[0:4])
	if err != nil {
		return nn, err
	}
	if hdr[0] == 0 && hdr[1] == 0 && hdr[2] == 0 && hdr[3] == 0 {
		return nn, errPaddingEncountered
	}
	f.Header.ID = string(hdr[0:4])

	// Read the rest of the header.
	n, err := io.ReadFull(r, hdr[4:10])
	nn += n
	if err != nil {
		return nn, err
	}

	// Process the frame's size.
	size, err := decodeSyncSafeUint32(hdr[4:8])
	if err != nil {
		return nn, err
	}
	if size < 1 {
		return nn, ErrInvalidFrameHeader
	}
	f.Header.Size = int(size)

	// Process the flags.
	flags := uint32(hdr[8])<<8 | uint32(hdr[9])
	f.Header.Flags = FrameFlags(c.frameFlags.Decode(flags))

	// Read the rest of the frame into a buffer.
	buf := make([]byte, size)
	n, err = io.ReadFull(r, buf)
	nn += n
	if err != nil {
		return n, err
	}

	// Strip unsync codes if the frame is unsynchronized but the tag isn't.
	if (f.Header.Flags&FrameFlagUnsynchronized) != 0 && (t.Flags&TagFlagUnsync) == 0 {
		buf = removeUnsyncCodes(buf)
	}

	// Initialize the scanner.
	c.buf, c.n, c.err = buf, nn, nil

	// Scan additional header fields indicated by flags.
	if f.Header.Flags != 0 {
		c.scanFlaggedHeaderData(&f.Header)
		if c.err != nil {
			return c.n, c.err
		}
	}

	// Select a frame payload type based on the ID.
	typ := c.payloadTypes.Lookup(f.Header.ID)

	// Instantiate a new frame payload using reflection.
	v := reflect.New(typ)

	// Use the reflection type of the payload to process the frame's data.
	enc := EncodingISO88591
	for i := 0; i < typ.NumField(); i++ {
		fieldValue := v.Elem().Field(i)
		field := typ.Field(i)
		kind := field.Type.Kind()
		tags := getTags(field, "id3")
		switch kind {
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				c.scanByteSlice(tags, fieldValue)
			case reflect.String:
				c.scanStringSlice(tags, fieldValue, enc)
			default:
				c.err = ErrUnknownFieldType
			}

		case reflect.String:
			c.scanString(tags, fieldValue, enc)

		case reflect.Uint8:
			switch field.Type.Name() {
			case "frameID":
				// Skip
			case "Encoding":
				enc = Encoding(c.scanUint8(tags, fieldValue, 0, 3))
			case "PictureType":
				c.scanUint8(tags, fieldValue, 0, 20)
			case "GroupSymbol":
				c.scanUint8(tags, fieldValue, 0x80, 0xf0)
			default:
				c.err = ErrUnknownFieldType
			}

		default:
			c.err = ErrUnknownFieldType
		}
	}

	if c.err == nil {
		f.Payload = v.Interface().(FramePayload)
	}

	return c.n, c.err
}

func (c *codec24) scanFlaggedHeaderData(h *FrameHeader) {
	// If the frame is compressed, it must include a data length indicator.
	if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
		c.err = ErrInvalidFrameFlags
		return
	}

	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		gid := c.consumeByte()
		if c.err != nil || gid < 0x80 || gid > 0xf0 {
			c.err = ErrInvalidFrameHeader
			return
		}
		h.GroupID = GroupSymbol(gid)
	}

	if (h.Flags & FrameFlagEncrypted) != 0 {
		em := c.consumeByte()
		if c.err != nil || em < 0x80 || em > 0xf0 {
			c.err = ErrInvalidFrameHeader
			return
		}
		h.EncryptMethod = em
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		b := c.consumeBytes(4)
		if c.err != nil {
			c.err = ErrInvalidFrameHeader
		}
		h.DataLength, c.err = decodeSyncSafeUint32(b)
	}
}

func (c *codec24) scanString(tags tagList, v reflect.Value, enc Encoding) string {
	var s string
	if c.err != nil {
		return s
	}

	if tags.Lookup("iso88519") {
		enc = EncodingISO88591
	}

	if tags.Lookup("lang") {
		s = c.consumeString(3, EncodingISO88591)
	} else {
		s = c.consumeNextString(enc)
	}
	if c.err != nil {
		return s
	}

	v.SetString(s)
	return s
}

func (c *codec24) scanStringSlice(tags tagList, v reflect.Value, enc Encoding) []string {
	var ss []string
	if c.err != nil {
		return ss
	}

	ss = c.consumeStrings(enc)
	if c.err != nil {
		return ss
	}
	v.Set(reflect.ValueOf(ss))
	return ss
}

func (c *codec24) scanByteSlice(tags tagList, v reflect.Value) []byte {
	var b []byte
	if c.err != nil {
		return b
	}

	b = c.consumeAll()
	v.Set(reflect.ValueOf(b))
	return b
}

func (c *codec24) scanUint8(tags tagList, v reflect.Value, min uint8, max uint8) uint8 {
	var e uint8
	if c.err != nil {
		return e
	}

	e = c.consumeByte()
	if c.err != nil || e < min || e > max {
		c.err = ErrInvalidFrame
		return e
	}

	v.SetUint(uint64(e))
	return e
}

func (c *codec24) EncodeFrame(t *Tag, f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec24) consumeByte() byte {
	if c.err != nil {
		return 0
	}
	if len(c.buf) < 1 {
		c.err = errInsufficientBuffer
		return 0
	}
	b := c.buf[0]
	c.buf = c.buf[1:]
	return b
}

func (c *codec24) consumeBytes(n int) []byte {
	if c.err != nil {
		return make([]byte, n)
	}
	if len(c.buf) < n {
		c.err = errInsufficientBuffer
		return make([]byte, n)
	}
	b := c.buf[:n]
	c.buf = c.buf[n:]
	return b
}

func (c *codec24) consumeStrings(enc Encoding) []string {
	if c.err != nil {
		return []string{}
	}

	var ss []string
	ss, c.err = decodeStrings(c.buf, enc)
	if c.err != nil {
		return ss
	}

	c.buf = c.buf[:0]
	return ss
}

func (c *codec24) consumeString(len int, enc Encoding) string {
	var s string

	b := c.consumeBytes(len)
	if c.err != nil {
		return s
	}

	s, c.err = decodeString(b, EncodingISO88591)
	return s
}

func (c *codec24) consumeNextString(enc Encoding) string {
	var s string
	if c.err != nil {
		return s
	}

	s, c.buf, c.err = decodeNextString(c.buf, enc)
	return s
}

func (c *codec24) consumeAll() []byte {
	if c.err != nil {
		return []byte{}
	}

	b := c.buf
	c.buf = c.buf[:0]
	return b
}
