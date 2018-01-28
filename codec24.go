package id3

import (
	"encoding/binary"
	"io"
	"reflect"
)

//
// codec24
//

type codec24 struct {
	payloadTypes typeMap
	buf          []byte // scan buffer
	n            int    // total bytes read or written from/to the stream
	err          error  // first error encountered when decoding or encoding
}

func newCodec24() *codec24 {
	return &codec24{
		payloadTypes: newTypeMap("v24"),
	}
}

func (c *codec24) decodeFrame(f *Frame, r io.Reader) (int, error) {
	// Read the first four bytes of the frame header to see if it's padding.
	c.buf = make([]byte, 10)
	c.n, c.err = r.Read(c.buf[0:4])
	if c.err != nil {
		return c.n, c.err
	}
	if c.buf[0] == 0 && c.buf[1] == 0 && c.buf[2] == 0 && c.buf[3] == 0 {
		return c.n, errPaddingEncountered
	}

	// Read the rest of the header.
	var n int
	n, c.err = r.Read(c.buf[4:10])
	c.n += n
	if c.err != nil {
		return c.n, c.err
	}

	// Examine the size field to figure out how much more data needs to
	// be read into the frame buffer.
	var size uint32
	size, c.err = decodeSyncSafeUint32(c.buf[4:8])
	if c.err != nil {
		return c.n, c.err
	}
	if size < 1 {
		return c.n, ErrInvalidFrameHeader
	}

	// Read the rest of the frame into the frame buffer.
	c.buf = append(c.buf, make([]byte, size)...)
	n, c.err = r.Read(c.buf[10:])
	c.n += n
	if c.err != nil {
		return c.n, c.err
	}

	// Scan the header's contents.
	c.scanFrameHeader(&f.Header)
	if c.err != nil {
		return c.n, c.err
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

func (c *codec24) scanFrameHeader(h *FrameHeader) {
	hdr := c.consumeBytes(10)
	if c.err != nil {
		c.err = ErrInvalidFrameHeader
		return
	}

	h.ID = string(hdr[0:4])

	var size uint32
	size, c.err = decodeSyncSafeUint32(hdr[4:8])
	if c.err != nil {
		c.err = ErrInvalidFrameHeader
		return
	}
	h.Size = int(size)

	h.Flags = 0
	flags := binary.BigEndian.Uint16(hdr[8:10])
	if flags != 0 {
		if (flags & (1 << 14)) != 0 {
			h.Flags |= FrameFlagDiscardOnTagAlteration
		}

		if (flags & (1 << 13)) != 0 {
			h.Flags |= FrameFlagDiscardOnFileAlteration
		}

		if (flags & (1 << 12)) != 0 {
			h.Flags |= FrameFlagReadOnly
		}

		if (flags & (1 << 6)) != 0 {
			h.Flags |= FrameFlagHasGroupInfo
			h.GroupID = GroupSymbol(c.consumeByte())
			if c.err != nil || h.GroupID < 0x80 || h.GroupID > 0xf0 {
				c.err = ErrInvalidFrameHeader
			}
		}

		if (flags & (1 << 3)) != 0 {
			h.Flags |= FrameFlagCompressed
		}

		if (flags & (1 << 2)) != 0 {
			h.Flags |= FrameFlagEncrypted
			h.EncryptMethod = c.consumeByte()
			if c.err != nil || h.EncryptMethod < 0x80 || h.EncryptMethod > 0xf0 {
				c.err = ErrInvalidFrameHeader
			}
		}

		if (flags & (1 << 1)) != 0 {
			h.Flags |= FrameFlagUnsynchronized
		}

		if (flags & (1 << 0)) != 0 {
			h.Flags |= FrameFlagHasDataLength
			b := c.consumeBytes(4)
			if c.err != nil {
				c.err = ErrInvalidFrameHeader
			}
			h.DataLength, c.err = decodeSyncSafeUint32(b)
		}

		// If the frame is compressed, it must include a data length indicator.
		if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
			c.err = ErrInvalidFrameFlags
		}
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

func (c *codec24) encodeFrame(f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
