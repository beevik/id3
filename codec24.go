package id3

import (
	"bytes"
	"encoding/binary"
	"io"
)

//
// codec24
//

type codec24 struct {
}

var frameCodecs = map[string]frameCodec{
	"T": &frameText24{},
}

func (c *codec24) decode(t *Tag, r io.Reader) (int, error) {
	nn := 0
	for remain := t.Size; remain > 0; {

		// Read the frame header.
		var h FrameHeader
		n, err := decodeFrameHeader24(&h, r)
		nn += n
		if err != nil {
			return nn, err
		}

		// Select a frame codec based on the frame header's ID value.
		id := string(h.IDvalue)
		if id[0] == 'T' && id != "TXXX" {
			id = "T"
		}
		fc, ok := frameCodecs[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Read the rest of the frame into a buffer.
		framebuf := make([]byte, h.Size)
		n, err = r.Read(framebuf)
		nn += n
		if err != nil {
			return nn, err
		}

		// Decode the contents of the buffer, generating the frame data.
		d, err := fc.decode(&h, framebuf)
		if err != nil {
			return nn, err
		}

		// Add to the tag's list of frames.
		t.Frames = append(t.Frames, Frame{h, d})

		remain -= h.Size + 10
	}

	return nn, nil
}

func (c *codec24) encode(t *Tag, w io.Writer) (int, error) {
	nn := 0

	for _, f := range t.Frames {
		// Select a frame codec based on the frame's ID value.
		id := string(f.Header.IDvalue)
		if id[0] == 'T' && id != "TXXX" {
			id = "T"
		}
		fc, ok := frameCodecs[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Encode the frame data (not including the header) into a new buffer.
		buf, err := fc.encode(&f.Header, f.Data)
		if err != nil {
			return nn, err
		}

		// Update the frame header's size based on the contents of the
		// buffer.
		h := f.Header
		h.Size = uint32(len(buf)) + h.ExtraBytes()

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
	h.IDvalue = string(buf[0:4])
	h.Size, err = readSyncSafeUint32(buf[4:8])
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
			buf := make([]byte, 1)
			n, err = r.Read(buf)
			nn += n
			if err != nil {
				return nn, err
			}
			h.GroupID = buf[0]
		}
		if (flags & (1 << 3)) != 0 {
			h.Flags |= FrameFlagCompressed
		}
		if (flags & (1 << 2)) != 0 {
			h.Flags |= FrameFlagEncrypted
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
			h.DataLength, err = readSyncSafeUint32(buf)
			if err != nil {
				return nn, err
			}
		}

		// If the frame is compressed or encrypted, it must include a data
		// length indicator.
		if (h.Flags&(FrameFlagCompressed|FrameFlagEncrypted)) != 0 &&
			(h.Flags&FrameFlagHasDataLength) == 0 {
			return nn, ErrInvalidFrameFlags
		}
	}

	return nn, nil
}

func encodeFrameHeader24(h *FrameHeader, w io.Writer) (int, error) {
	nn := 0

	// Write the frame ID.
	idval := []byte(h.IDvalue)
	n, err := w.Write(idval)
	nn += n
	if err != nil {
		return nn, err
	}

	// Create a 6-byte buffer for the rest of the header and store the
	// frame size in it.
	buf := make([]byte, 6)
	err = writeSyncSafeUint32(buf[0:4], h.Size)
	if err != nil {
		return nn, err
	}

	var flags uint16
	if h.Flags != 0 {
		if (h.Flags&(FrameFlagCompressed|FrameFlagEncrypted)) != 0 &&
			(h.Flags&FrameFlagHasDataLength) == 0 {
			return nn, ErrInvalidFrameFlags
		}
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

	n, err = w.Write(buf)
	nn += n
	if err != nil {
		return nn, err
	}

	if (h.Flags & FrameFlagHasGroupInfo) != 0 {
		buf := []byte{h.GroupID}
		n, err = w.Write(buf)
		nn += n
		if err != nil {
			return nn, err
		}
	}

	if (h.Flags & FrameFlagHasDataLength) != 0 {
		buf := make([]byte, 4)
		err = writeSyncSafeUint32(buf, h.DataLength)
		if err != nil {
			return nn, err
		}
		n, err = w.Write(buf)
		nn += n
		if err != nil {
			return nn, err
		}
	}

	return nn, err
}

//
// frameText24: v2.4 Text frame codec
//

type frameText24 struct{}

func (c *frameText24) decode(h *FrameHeader, buf []byte) (FrameData, error) {
	if buf[0] > 3 {
		return nil, ErrInvalidEncoding
	}

	f := &FrameDataText{}
	f.Encoding = Encoding(buf[0])

	btmp := bytes.NewBuffer(buf[1:])
	var err error
	_, f.Text, err = readEncodedString(btmp, btmp.Len(), f.Encoding)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *frameText24) encode(h *FrameHeader, d FrameData) ([]byte, error) {
	t := d.(*FrameDataText)

	tmpbuf := bytes.NewBuffer([]byte{})
	err := tmpbuf.WriteByte(byte(t.Encoding))
	if err != nil {
		return nil, err
	}

	_, err = writeEncodedString(tmpbuf, t.Text, t.Encoding)
	if err != nil {
		return nil, err
	}

	return tmpbuf.Bytes(), nil
}
