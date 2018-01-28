package id3

import (
	"bytes"
	"encoding/binary"
	"io"
)

// ID3 v2.4 codec
type codec24 struct {
}

// Per-frame codecs for v2.4
type frameCodec24 struct {
	decodePayload func(c *codec24, f *Frame, r io.Reader) (int, error)
	encodePayload func(c *codec24, f *Frame, w io.Writer) (int, error)
}

var frameCodecs24 = map[string]frameCodec24{
	"T???": {(*codec24).decodeTextFrame, (*codec24).encodeTextFrame},
	"TXXX": {(*codec24).decodeTXXXFrame, (*codec24).encodeTXXXFrame},
	"APIC": {(*codec24).decodeAPICFrame, (*codec24).encodeAPICFrame},
	"":     {(*codec24).decodeUnknownFrame, (*codec24).encodeUnknownFrame},
}

func getFrameCodec24(id string) frameCodec24 {
	// All text frames (except TXXX) use the same codec.
	if id[0] == 'T' && id != "TXXX" {
		id = "T???"
	}

	c, ok := frameCodecs24[id]
	if ok {
		return c
	}
	return frameCodecs24[""] // unknown frame codec
}

func (c *codec24) decodeFrame(f *Frame, r io.Reader) (int, error) {
	nn := 0

	// Decode the frame header.
	n, err := c.decodeFrameHeader(&f.Header, r)
	nn += n
	if err != nil {
		return nn, err
	}

	// Use the frame header ID to look up the appropriate frame codec, and
	// use the frame codec to decode the frame's payload.
	fc := getFrameCodec24(f.Header.ID)
	n, err = fc.decodePayload(c, f, r)
	nn += n
	return nn, err
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

func (c *codec24) decodeTextFrame(f *Frame, r io.Reader) (int, error) {
	p := &FramePayloadText{}

	b := make([]byte, f.Header.Size)
	n, err := r.Read(b)
	if err != nil {
		return n, err
	}

	if b[0] > 3 {
		return n, ErrInvalidEncoding
	}
	p.Encoding = Encoding(b[0])

	p.Text, err = decodeStrings(b[1:], p.Encoding)
	if err != nil {
		return n, err
	}

	f.Payload = p
	return n, nil
}

func (c *codec24) decodeTXXXFrame(f *Frame, r io.Reader) (int, error) {
	p := &FramePayloadTXXX{}

	b := make([]byte, f.Header.Size)
	n, err := r.Read(b)
	if err != nil {
		return n, err
	}

	if b[0] > 3 {
		return n, ErrInvalidEncoding
	}
	p.Encoding = Encoding(b[0])

	p.Description, b, err = decodeNextString(b[1:], p.Encoding)
	if err != nil {
		return n, err
	}

	p.Text, err = decodeString(b, p.Encoding)
	if err != nil {
		return n, err
	}

	f.Payload = p
	return n, nil
}

func (c *codec24) decodeAPICFrame(f *Frame, r io.Reader) (int, error) {
	p := &FramePayloadAPIC{}

	b := make([]byte, f.Header.Size)
	n, err := r.Read(b)
	if err != nil {
		return n, err
	}

	if b[0] > 3 {
		return n, ErrInvalidEncoding
	}
	p.Encoding = Encoding(b[0])
	b = b[1:]

	if len(b) < 1 {
		return n, ErrInvalidFrame
	}
	p.MimeType, b, err = decodeNextString(b, EncodingISO88591)
	if err != nil {
		return n, err
	}

	if len(b) < 1 {
		return n, ErrInvalidFrame
	}
	p.Type = PictureType(b[0])

	b = b[1:]
	if len(b) < 1 {
		return n, ErrInvalidFrame
	}
	p.Description, b, err = decodeNextString(b, p.Encoding)
	if err != nil {
		return n, ErrInvalidFrame
	}

	p.Data = []byte(b)

	f.Payload = p
	return n, nil
}

func (c *codec24) decodeUnknownFrame(f *Frame, r io.Reader) (int, error) {
	p := &FramePayloadUnknown{}

	b := make([]byte, f.Header.Size)
	n, err := r.Read(b)
	if err != nil {
		return n, err
	}

	p.Data = b
	f.Payload = p
	return n, nil
}

func (c *codec24) encodeFrame(f *Frame, w io.Writer) (int, error) {
	// Encode the payload into a temporary buffer, since we need to
	// compute the size of the payload before writing the frame header.
	buf := bytes.NewBuffer([]byte{})

	// Use the frame header ID to look up the appropriate frame codec, and
	// use the frame codec to encode the frame payload into the temporary
	// buffer.
	fc := getFrameCodec24(f.Header.ID)
	size, err := fc.encodePayload(c, f, buf)
	if err != nil {
		return 0, err
	}

	nn := 0

	// Write the frame header to the output.
	n, err := c.encodeFrameHeader(&f.Header, size, w)
	nn += n
	if err != nil {
		return nn, err
	}

	// Write the frame payload buffer to the output.
	n, err = w.Write(buf.Bytes())
	nn += n
	return nn, err
}

func (c *codec24) encodeFrameHeader(h *FrameHeader, size int, w io.Writer) (int, error) {
	nn := 0

	// Write the frame ID.
	idval := []byte(h.ID)
	n, err := w.Write(idval)
	nn += n
	if err != nil {
		return nn, err
	}

	// Generate the flags field (and adjust frame size if necessary).
	var flags uint16
	if h.Flags != 0 {
		if (h.Flags & FrameFlagDiscardOnTagAlteration) != 0 {
			flags |= 1 << 14
		}
		if (h.Flags & FrameFlagDiscardOnFileAlteration) != 0 {
			flags |= 1 << 13
		}
		if (h.Flags & FrameFlagReadOnly) != 0 {
			flags |= 1 << 12
		}
		if (h.Flags & FrameFlagHasGroupInfo) != 0 {
			flags |= 1 << 6
			size++
		}
		if (h.Flags & FrameFlagCompressed) != 0 {
			flags |= 1 << 3
			if (h.Flags & FrameFlagHasDataLength) == 0 {
				return nn, ErrInvalidFrameFlags
			}
		}
		if (h.Flags & FrameFlagEncrypted) != 0 {
			flags |= 1 << 2
			size++
		}
		if (h.Flags & FrameFlagUnsynchronized) != 0 {
			flags |= 1 << 1
		}
		if (h.Flags & FrameFlagHasDataLength) != 0 {
			flags |= 1 << 0
			size += 4
		}
	}

	// Create a 6-byte buffer for the rest of the header and store the
	// frame size and flags in it.
	buf := make([]byte, 6)
	err = encodeSyncSafeUint32(buf[0:4], uint32(size))
	if err != nil {
		return nn, err
	}
	buf[4] = uint8(flags >> 8)
	buf[5] = uint8(flags)

	// Write the size and flags.
	n, err = w.Write(buf)
	nn += n
	if err != nil {
		return nn, err
	}

	// Write any extra data indicated by flags.
	if h.Flags != 0 {
		ex := make([]byte, 0)
		if (h.Flags & FrameFlagHasGroupInfo) != 0 {
			ex = append(ex, h.GroupID)
		}
		if (h.Flags & FrameFlagEncrypted) != 0 {
			ex = append(ex, h.EncryptMethod)
		}
		if (h.Flags & FrameFlagHasDataLength) != 0 {
			len := make([]byte, 4)
			err = encodeSyncSafeUint32(len, h.DataLength)
			ex = append(ex, len...)
			if err != nil {
				return nn, err
			}
		}
		if len(ex) > 0 {
			n, err = w.Write(ex)
			nn += n
			if err != nil {
				return nn, err
			}
		}
	}

	return nn, err
}

// Return the total number of bytes required to store the frame header and any
// extra fields indicated by the header flags.
func (c *codec24) extraBytes(h *FrameHeader) uint32 {
	var n uint32
	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		n++
	}
	if (h.Flags & FrameFlagEncrypted) != 0 {
		n++
	}
	if (h.Flags & FrameFlagHasDataLength) != 0 {
		n += 4
	}
	return n
}

func (c *codec24) encodeTextFrame(f *Frame, w io.Writer) (int, error) {
	p := f.Payload.(*FramePayloadText)

	nn := 0

	n, err := w.Write([]byte{byte(p.Encoding)})
	nn += n
	if err != nil {
		return nn, err
	}

	sb, err := encodeStrings(p.Text, p.Encoding)
	if err != nil {
		return nn, err
	}

	n, err = w.Write(sb)
	nn += n
	return nn, err
}

func (c *codec24) encodeTXXXFrame(f *Frame, w io.Writer) (int, error) {
	p := f.Payload.(*FramePayloadTXXX)

	nn := 0

	n, err := w.Write([]byte{byte(p.Encoding)})
	nn += n
	if err != nil {
		return nn, err
	}

	sb, err := encodeString(p.Description, p.Encoding)
	if err != nil {
		return nn, err
	}
	sb = append(sb, null[p.Encoding]...)
	n, err = w.Write(sb)
	nn += n
	if err != nil {
		return nn, err
	}

	sb, err = encodeString(p.Text, p.Encoding)
	n, err = w.Write(sb)
	nn += n
	return nn, err
}

func (c *codec24) encodeAPICFrame(f *Frame, w io.Writer) (int, error) {
	p := f.Payload.(*FramePayloadAPIC)

	nn := 0

	n, err := w.Write([]byte{byte(p.Encoding)})
	nn += n
	if err != nil {
		return nn, err
	}

	sb, err := encodeString(p.MimeType, EncodingISO88591)
	if err != nil {
		return nn, err
	}
	sb = append(sb, 0)

	n, err = w.Write(sb)
	nn += n
	if err != nil {
		return nn, err
	}

	n, err = w.Write([]byte{byte(p.Type)})
	nn += n
	if err != nil {
		return nn, err
	}

	sb, err = encodeString(p.Description, p.Encoding)
	if err != nil {
		return nn, err
	}
	sb = append(sb, null[p.Encoding]...)

	n, err = w.Write(sb)
	nn += n
	if err != nil {
		return nn, err
	}

	n, err = w.Write(p.Data)
	nn += n
	return nn, err
}

func (c *codec24) encodeUnknownFrame(f *Frame, w io.Writer) (int, error) {
	p := f.Payload.(*FramePayloadUnknown)
	return w.Write(p.Data)
}
