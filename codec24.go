package id3

import (
	"encoding/binary"
	"io"
)

//
// codec24
//

type codec24 struct {
}

var frameCodecs = map[string]frameCodec{
	"T": &frameCodecText24{},
}

func (c *codec24) decode(t *Tag, r io.Reader) (int, error) {
	nn := 0
	for remain := t.Size; remain > 0; {

		// Create an empty frame.
		f := Frame{}

		// Decode the frame header.
		n, err := decodeFrameHeader24(&f.Header, r)
		nn += n
		if err != nil {
			return nn, err
		}

		// Select a frame codec based on the frame header's ID value.
		id := string(f.Header.ID)
		if id[0] == 'T' && id != "TXXX" {
			id = "T"
		}
		fc, ok := frameCodecs[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Read the frame's payload into a buffer.
		databuf := make([]byte, f.Header.Size-extraBytes24(&f.Header))
		n, err = r.Read(databuf)
		nn += n
		if err != nil {
			return nn, err
		}

		// Decode the contents of the buffer, generating the frame data.
		f.Data, err = fc.decode(&f.Header, databuf)
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
		fc, ok := frameCodecs[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Encode the frame data (not including the header) into a
		// new payload buffer.
		buf, err := fc.encode(&f.Header, f.Data)
		if err != nil {
			return nn, err
		}

		// Update the rame header's size field based on the contents of the
		// payload buffer.
		h := f.Header
		h.Size = uint32(len(buf)) + extraBytes24(&h)

		// Write the updated frame header to the output.
		n, err := encodeFrameHeader24(&h, w)
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

//
// v2.4 FrameHeader codec
//

func decodeFrameHeader24(h *FrameHeader, r io.Reader) (int, error) {
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

func encodeFrameHeader24(h *FrameHeader, w io.Writer) (int, error) {
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
func extraBytes24(h *FrameHeader) uint32 {
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
// frameCodecText24: v2.4 Text frame codec
//

type frameCodecText24 struct{}

func (c *frameCodecText24) decode(h *FrameHeader, buf []byte) (FrameData, error) {
	if buf[0] > 3 {
		return nil, ErrInvalidEncoding
	}

	f := &FrameDataText{}
	f.Encoding = Encoding(buf[0])

	var err error
	f.Text, err = decodeString(buf[1:], f.Encoding)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *frameCodecText24) encode(h *FrameHeader, d FrameData) ([]byte, error) {
	t := d.(*FrameDataText)

	buf := make([]byte, 0, len(t.Text)+1)
	buf = append(buf, byte(t.Encoding))

	b, err := encodeString(t.Text, t.Encoding)
	if err != nil {
		return nil, err
	}

	return append(buf, b...), nil
}
