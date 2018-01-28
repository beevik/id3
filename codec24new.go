package id3

import (
	"encoding/binary"
	"io"
	"reflect"
	"strings"
)

type parser struct {
	buf []byte
	err error
	n   int
}

func newParser(r io.Reader, size uint32) *parser {
	p := &parser{}
	p.buf = make([]byte, int(size))
	p.n, p.err = r.Read(p.buf)
	return p
}

func tagContains(f reflect.StructField, s string) bool {
	if f.Tag == "" {
		return false
	}
	tag := string(f.Tag)
	if !strings.HasPrefix(tag, "id3:") {
		return false
	}
	tag = tag[5 : len(tag)-1]
	for _, t := range strings.Split(tag, ",") {
		if t == s {
			return true
		}
	}
	return false
}

func (p *parser) readString(f reflect.StructField, v reflect.Value, enc Encoding) {
	if p.err != nil {
		return
	}

	if tagContains(f, "iso88519") {
		enc = EncodingISO88591
	}

	var s string
	s, p.buf, p.err = decodeNextString(p.buf, enc)
	if p.err != nil {
		return
	}

	v.SetString(s)
}

func (p *parser) readStringSlice(f reflect.StructField, v reflect.Value, enc Encoding) {
	if p.err != nil {
		return
	}

	var ss []string
	ss, p.err = decodeStrings(p.buf, enc)
	if p.err != nil {
		return
	}

	p.buf = p.buf[:0]

	slice := reflect.MakeSlice(v.Type(), len(ss), len(ss))
	for i := 0; i < len(ss); i++ {
		slice.Index(i).SetString(ss[i])
	}
	v.Set(slice)
}

func (p *parser) readByteSlice(f reflect.StructField, v reflect.Value) {
	if p.err != nil {
		return
	}

	slice := reflect.MakeSlice(v.Type(), len(p.buf), len(p.buf))
	reflect.Copy(slice, reflect.ValueOf(p.buf))
	v.Set(slice)

	p.buf = p.buf[:0]
}

func (p *parser) readUint8(f reflect.StructField, v reflect.Value, max uint8) uint8 {
	if p.err != nil {
		return 0
	}

	if len(p.buf) < 1 {
		p.err = ErrInvalidFrame
		return 0
	}

	var e uint8
	e = p.buf[0]
	if e > max {
		p.err = ErrInvalidEncoding
		return 0
	}

	p.buf = p.buf[1:]
	v.SetUint(uint64(e))
	return e
}

var (
	fText = FramePayloadText{}
	fTXXX = FramePayloadTXXX{}
	fAPIC = FramePayloadAPIC{}
	fUFID = FramePayloadUFID{}
	fUSLT = FramePayloadUSLT{}
	fUnkn = FramePayloadUnknown{}
)

var table = map[string]reflect.Type{
	"T":    reflect.TypeOf(fText),
	"TXXX": reflect.TypeOf(fTXXX),
	"APIC": reflect.TypeOf(fAPIC),
	"UFID": reflect.TypeOf(fUFID),
	"USLT": reflect.TypeOf(fUSLT),
	"":     reflect.TypeOf(fUnkn),
}

type codec24new struct {
}

func (c *codec24new) decodeFrame(f *Frame, r io.Reader) (int, error) {
	nn := 0

	// Decode the frame header.
	n, err := c.decodeFrameHeader(&f.Header, r)
	nn += n
	if err != nil {
		return nn, err
	}

	// decode payload here
	id := string(f.Header.ID)
	if id[0] == 'T' {
		id = "T"
	}
	typ, ok := table[id]
	if !ok {
		typ = table[""]
	}

	parser := newParser(r, f.Header.Size)
	v := reflect.New(typ)
	elem := v.Elem()

	// New returns a Value representing a pointer to a new zero value for the
	// specified type. That is, the returned Value's Type is PtrTo(typ).
	//v := reflect.New(typ)
	enc := EncodingISO88591
	fields := typ.NumField()
	for i := 0; i < fields; i++ {
		field := typ.Field(i)
		switch {

		case field.Type.Kind() == reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				parser.readByteSlice(field, elem.Field(i))
			case reflect.String:
				parser.readStringSlice(field, elem.Field(i), enc)
			default:
				parser.err = ErrUnknownFieldType
			}

		case field.Type.Kind() == reflect.String:
			parser.readString(field, elem.Field(i), enc)

		case field.Type.Kind() == reflect.Uint8:
			switch field.Type.Name() {
			case "Encoding":
				enc = Encoding(parser.readUint8(field, elem.Field(i), 3))
			case "PictureType":
				parser.readUint8(field, elem.Field(i), 20)
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

func (c *codec24new) decodeFrameHeader(h *FrameHeader, r io.Reader) (int, error) {
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
			h.GroupID, err = readByte(r)
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

func (c *codec24new) encodeFrame(f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
