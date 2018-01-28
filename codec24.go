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
}

var frameTypeTable24 = map[string]reflect.Type{
	"????": reflect.TypeOf(FramePayloadUnknown{}),
	"T___": reflect.TypeOf(FramePayloadText{}),
	"TXXX": reflect.TypeOf(FramePayloadTXXX{}),
	"APIC": reflect.TypeOf(FramePayloadAPIC{}),
	"UFID": reflect.TypeOf(FramePayloadUFID{}),
	"USER": reflect.TypeOf(FramePayloadUSER{}),
	"USLT": reflect.TypeOf(FramePayloadUSLT{}),
	"GRID": reflect.TypeOf(FramePayloadGRID{}),
}

func (c *codec24) getFramePayloadType(id string) reflect.Type {
	if id[0] == 'T' {
		id = "T___"
	}
	if typ, ok := frameTypeTable24[id]; ok {
		return typ
	}
	return frameTypeTable24["????"]
}

func (c *codec24) decodeFrame(f *Frame, r io.Reader) (int, error) {
	nn := 0

	// Decode the frame header.
	n, err := c.decodeFrameHeader(&f.Header, r)
	nn += n
	if err != nil {
		return nn, err
	}

	// Select a frame payload type based on the ID.
	typ := c.getFramePayloadType(f.Header.ID)

	// Instantiate a new frame payload using reflection.
	v := reflect.New(typ)
	elem := v.Elem()

	// Create a frame payload parser.
	parser := newParser24(r, f.Header.Size)

	// Use the reflection type of the payload to process the frame's data.
	enc := EncodingISO88591
	fields := typ.NumField()
	for i := 0; i < fields; i++ {
		fieldValue := elem.Field(i)
		field := typ.Field(i)
		kind := field.Type.Kind()
		switch {

		case kind == reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				parser.readByteSlice(field, fieldValue)
			case reflect.String:
				parser.readStringSlice(field, fieldValue, enc)
			default:
				parser.err = ErrUnknownFieldType
			}

		case kind == reflect.String:
			parser.readString(field, fieldValue, enc)

		case kind == reflect.Uint8:
			switch field.Type.Name() {
			case "Encoding":
				enc = Encoding(parser.readUint8(field, fieldValue, 0, 3))
			case "PictureType":
				parser.readUint8(field, fieldValue, 0, 20)
			case "GroupSymbol":
				parser.readUint8(field, fieldValue, 0x80, 0xf0)
			default:
				parser.err = ErrUnknownFieldType
			}

		default:
			parser.err = ErrUnknownFieldType
		}
	}

	nn += parser.n
	if parser.err == nil {
		f.Payload = v.Interface().(FramePayload)
	}
	return nn, parser.err
}

func (c *codec24) decodeFrameHeader(h *FrameHeader, r io.Reader) (int, error) {
	nn := 0

	// Read the 4-byte frame ID.
	buf := make([]byte, 10)
	n, err := r.Read(buf[0:4])
	nn += n
	if err != nil {
		return nn, err
	}
	if buf[0] == 0 && buf[1] == 0 && buf[2] == 0 && buf[3] == 0 {
		return nn, errPaddingEncountered
	}

	// Read the rest of the 10-byte frame header.
	n, err = r.Read(buf[4:10])
	nn += n
	if err != nil {
		return nn, err
	}

	// Process the header ID and size.
	h.ID = string(buf[0:4])
	h.Size, err = decodeSyncSafeUint32(buf[4:8])
	if err != nil {
		return n, err
	}
	if h.Size < 1 {
		return n, ErrInvalidFrameHeader
	}

	// Process header flags and load additional data if necessary.
	h.Flags = 0
	flags := binary.BigEndian.Uint16(buf[8:10])
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
			gid, err := readByte(r)
			h.GroupID = GroupSymbol(gid)
			nn++
			if err != nil {
				return nn, err
			}
		}
		if (flags & (1 << 3)) != 0 {
			h.Flags |= FrameFlagCompressed
		}
		if (flags & (1 << 2)) != 0 {
			h.Flags |= FrameFlagEncrypted
			h.EncryptMethod, err = readByte(r)
			nn++
			if err != nil {
				return nn, err
			}
		}
		if (flags & (1 << 1)) != 0 {
			h.Flags |= FrameFlagUnsynchronized
		}
		if (flags & (1 << 0)) != 0 {
			h.Flags |= FrameFlagHasDataLength
			buf := make([]byte, 4)
			n, err = r.Read(buf)
			nn += n
			if err != nil {
				return nn, err
			}
			h.DataLength, err = decodeSyncSafeUint32(buf)
			if err != nil {
				return nn, err
			}
		}

		// If the frame is compressed, it must include a data length indicator.
		if (h.Flags&FrameFlagCompressed) != 0 && (h.Flags&FrameFlagHasDataLength) == 0 {
			return nn, ErrInvalidFrameFlags
		}
	}

	return nn, nil
}

func (c *codec24) encodeFrame(f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}

//
// parser24
//

type parser24 struct {
	buf []byte
	err error
	n   int
}

func newParser24(r io.Reader, size uint32) *parser24 {
	p := &parser24{}
	p.buf = make([]byte, int(size))
	p.n, p.err = r.Read(p.buf)
	return p
}

func (p *parser24) readString(f reflect.StructField, v reflect.Value, enc Encoding) string {
	var s string

	if p.err != nil {
		return s
	}

	if tagContains(f, "iso88519") {
		enc = EncodingISO88591
	}

	var b []byte

	if tagContains(f, "lang") {
		if len(p.buf) < 3 {
			p.err = ErrInvalidFrame
			return s
		}
		s, _, p.err = decodeNextString(p.buf[:3], EncodingISO88591)
		b = p.buf[3:]
	} else {
		s, b, p.err = decodeNextString(p.buf, enc)
	}

	if p.err != nil {
		return s
	}

	v.SetString(s)

	p.buf = b
	return s
}

func (p *parser24) readStringSlice(f reflect.StructField, v reflect.Value, enc Encoding) []string {
	var ss []string

	if p.err != nil {
		return ss
	}

	ss, p.err = decodeStrings(p.buf, enc)
	if p.err != nil {
		return ss
	}

	slice := reflect.MakeSlice(v.Type(), len(ss), len(ss))
	reflect.Copy(slice, reflect.ValueOf(ss))
	v.Set(slice)

	p.buf = p.buf[:0]
	return ss
}

func (p *parser24) readByteSlice(f reflect.StructField, v reflect.Value) []byte {
	var b []byte
	if p.err != nil {
		return b
	}

	b = p.buf
	slice := reflect.MakeSlice(v.Type(), len(b), len(b))
	reflect.Copy(slice, reflect.ValueOf(b))
	v.Set(slice)

	p.buf = p.buf[:0]
	return b
}

func (p *parser24) readUint8(f reflect.StructField, v reflect.Value, min uint8, max uint8) uint8 {
	var e uint8

	if p.err != nil {
		return e
	}

	if len(p.buf) < 1 {
		p.err = ErrInvalidFrame
		return e
	}

	e = p.buf[0]
	if e < min || e > max {
		p.err = ErrInvalidFrame
		return e
	}

	v.SetUint(uint64(e))

	p.buf = p.buf[1:]
	return e
}
