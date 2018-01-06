package id3

import (
	"encoding/binary"
	"io"
)

//
// codec24: ID3 v2.4 codec
//

type codec24 struct {
}

var frameCodecs24 = map[string]frameCodec{
	"T":    &frameText24{},
	"APIC": &frameAPIC24{},
}

func (c *codec24) decode(t *Tag, r io.Reader) (int, error) {
	nn := 0
	for remain := t.Size; remain > 0; {

		// Create an empty frame.
		f := Frame{}

		// Decode the frame header.
		n, err := c.decodeFrameHeader(&f.Header, r)
		nn += n
		if err != nil {
			return nn, err
		}

		// Select a frame codec based on the frame header's ID value.
		id := string(f.Header.ID)
		if id[0] == 'T' && id != "TXXX" {
			id = "T"
		}
		fc, ok := frameCodecs24[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Read the frame's payload into a buffer.
		databuf := make([]byte, f.Header.Size-c.extraBytes(&f.Header))
		n, err = r.Read(databuf)
		nn += n
		if err != nil {
			return nn, err
		}

		// Decode the contents of the buffer, generating the frame data.
		f.Payload, err = fc.decode(&f.Header, databuf)
		if err != nil {
			return nn, err
		}

		// Add the frame to the tag.
		t.Frames = append(t.Frames, f)

		remain -= f.Header.Size + 10
	}

	return nn, nil
}

func (c *codec24) encode(t *Tag, w io.Writer) (int, error) {
	nn := 0

	for _, f := range t.Frames {
		// Select a frame codec based on the frame's ID value.
		id := string(f.Header.ID)
		if id[0] == 'T' && id != "TXXX" {
			id = "T"
		}
		fc, ok := frameCodecs24[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Encode the frame data (not including the header) into a
		// new payload buffer.
		buf, err := fc.encode(&f.Header, f.Payload)
		if err != nil {
			return nn, err
		}

		// Update the frame header's size field based on the contents of the
		// payload buffer.
		h := f.Header
		h.Size = uint32(len(buf)) + c.extraBytes(&h)

		// Write the updated frame header to the output.
		n, err := c.encodeFrameHeader(&h, w)
		nn += n
		if err != nil {
			return nn, err
		}

		// Write the frame data buffer to the output.
		n, err = w.Write(buf)
		nn += n
		if err != nil {
			return nn, err
		}
	}

	return nn, nil
}

func (c *codec24) decodeFrameHeader(h *FrameHeader, r io.Reader) (int, error) {
	nn := 0

	// Read the 10-byte frame header.
	buf := make([]byte, 10)
	n, err := r.Read(buf)
	nn += n
	if n < 10 || err != nil {
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

func (c *codec24) encodeFrameHeader(h *FrameHeader, w io.Writer) (int, error) {
	nn := 0

	// Write the frame ID.
	idval := []byte(h.ID)
	n, err := w.Write(idval)
	nn += n
	if err != nil {
		return nn, err
	}

	// Create a 6-byte buffer for the rest of the header and store the
	// frame size in it.
	buf := make([]byte, 6)
	err = encodeSyncSafeUint32(buf[0:4], h.Size)
	if err != nil {
		return nn, err
	}

	// Generate the flags field.
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
		}
		if (h.Flags & FrameFlagCompressed) != 0 {
			flags |= 1 << 3
			if (h.Flags & FrameFlagHasDataLength) == 0 {
				return nn, ErrInvalidFrameFlags
			}
		}
		if (h.Flags & FrameFlagEncrypted) != 0 {
			flags |= 1 << 2
		}
		if (h.Flags & FrameFlagUnsynchronized) != 0 {
			flags |= 1 << 1
		}
		if (h.Flags & FrameFlagHasDataLength) != 0 {
			flags |= 1 << 0
		}
	}
	buf[4] = uint8(flags >> 8)
	buf[5] = uint8(flags)

	// Write the header.
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

//
// frameText24
//

type frameText24 struct{}

func (c *frameText24) decode(h *FrameHeader, b []byte) (FramePayload, error) {
	f := &FramePayloadText{}

	if b[0] > 3 {
		return nil, ErrInvalidEncoding
	}
	f.Encoding = Encoding(b[0])

	var err error
	f.Text, err = decodeStrings(b[1:], f.Encoding)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *frameText24) encode(h *FrameHeader, d FramePayload) ([]byte, error) {
	f := d.(*FramePayloadText)

	b := []byte{byte(f.Encoding)}

	sb, err := encodeStrings(f.Text, f.Encoding)
	if err != nil {
		return nil, err
	}

	return append(b, sb...), nil
}

//
// frameAPIC24
//

type frameAPIC24 struct{}

func (c *frameAPIC24) decode(h *FrameHeader, b []byte) (FramePayload, error) {
	f := &FramePayloadAPIC{}

	if b[0] > 3 {
		return nil, ErrInvalidEncoding
	}
	f.Encoding = Encoding(b[0])
	b = b[1:]

	if len(b) < 1 {
		return nil, ErrInvalidFrame
	}
	var err error
	var cnt int
	f.MimeType, cnt, err = decodeNextString(b, EncodingISO88591)
	if err != nil {
		return nil, err
	}
	b = b[cnt:]

	if len(b) < 1 {
		return nil, ErrInvalidFrame
	}
	f.Type = PictureType(b[0])

	b = b[1:]
	if len(b) < 1 {
		return nil, ErrInvalidFrame
	}
	f.Description, cnt, err = decodeNextString(b, f.Encoding)
	if err != nil {
		return nil, ErrInvalidFrame
	}

	b = b[cnt:]
	f.Data = []byte(b)

	return f, nil
}

func (c *frameAPIC24) encode(h *FrameHeader, d FramePayload) ([]byte, error) {
	f := d.(*FramePayloadAPIC)

	b := []byte{byte(f.Encoding)}

	sb, err := encodeString(f.MimeType, EncodingISO88591)
	if err != nil {
		return nil, err
	}
	b = append(b, sb...)
	b = append(b, 0)

	b = append(b, byte(f.Type))

	sb, err = encodeString(f.Description, f.Encoding)
	if err != nil {
		return nil, err
	}

	b = append(b, sb...)
	b = append(b, null[f.Encoding]...)

	b = append(b, f.Data...)

	return b, nil
}
