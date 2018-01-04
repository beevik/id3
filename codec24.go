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
		// Peek at the frame header.
		hdrbuf := make([]byte, 10)
		n, err := r.Read(hdrbuf)
		nn += n
		if n < 10 {
			return nn, ErrInvalidFrameHeader
		}
		if err != nil {
			return nn, err
		}

		// Select a frame codec based on the frame header's ID value.
		id := string(hdrbuf[0:4])
		if id[0] == 'T' {
			id = "T"
		}
		fc, ok := frameCodecs[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Get the frame's size, not including the header.
		size, err := readSyncSafeUint32(hdrbuf[4:8])
		if err != nil {
			return nn, err
		}

		// Read the rest of the frame into a buffer
		framebuf := make([]byte, size)
		n, err = r.Read(framebuf)
		nn += n
		if err != nil {
			return nn, err
		}

		// Decode the contents of the buffer, generating a frame.
		f, err := fc.decode(bytes.NewBuffer(append(hdrbuf, framebuf...)))
		if err != nil {
			return nn, err
		}

		// Add to the tag's list of frames.
		t.Frames = append(t.Frames, f)

		remain -= size + 10
	}

	return nn, nil
}

func (c *codec24) encode(t *Tag, w io.Writer) (int, error) {
	nn := 0

	for _, f := range t.Frames {
		// Select a frame codec based on the frame's ID value.
		id := f.ID()
		if id[0] == 'T' {
			id = "T"
		}
		fc, ok := frameCodecs[id]
		if !ok {
			return nn, ErrUnknownFrameType
		}

		// Encode the frame into a new buffer.
		buf := bytes.NewBuffer([]byte{})
		err := fc.encode(f, buf)
		if err != nil {
			return nn, err
		}

		// Write the buffer to the output.
		n, err := w.Write(buf.Bytes())
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

	buf := make([]byte, 10)
	n, err := r.Read(buf)
	nn += n
	if n < 10 || err != nil {
		return nn, err
	}

	h.IDvalue = string(buf[0:4])
	h.Size, err = readSyncSafeUint32(buf[4:8])
	h.Flags = 0
	if err != nil {
		return n, err
	}

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
	}

	// If the frame is compressed or encrypted, it must include a data length
	// indicator.
	if (h.Flags&(FrameFlagCompressed|FrameFlagEncrypted)) != 0 &&
		(h.Flags&FrameFlagHasDataLength) == 0 {
		return nn, ErrInvalidFrameFlags
	}

	return nn, nil
}

func encodeFrameHeader24(h *FrameHeader, w io.Writer) (int, error) {
	nn := 0

	idval := []byte(h.IDvalue)
	n, err := w.Write(idval)
	nn += n
	if err != nil {
		return nn, err
	}

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

func (c *frameText24) decode(buf *bytes.Buffer) (Frame, error) {
	f := NewFrameText("")

	_, err := decodeFrameHeader24(&f.FrameHeader, buf)
	if err != nil {
		return nil, err
	}

	enc, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if enc < 0 || enc > 4 {
		return nil, ErrInvalidEncoding
	}
	f.Encoding = Encoding(enc)

	_, f.Text, err = readEncodedString(buf, buf.Len(), f.Encoding)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *frameText24) encode(frame Frame, buf *bytes.Buffer) error {
	f := frame.(*FrameText)

	tmp := bytes.NewBuffer([]byte{})
	err := tmp.WriteByte(byte(f.Encoding))
	if err != nil {
		return err
	}

	_, err = writeEncodedString(tmp, f.Text, f.Encoding)
	if err != nil {
		return err
	}

	f.FrameHeader.Size = uint32(len(tmp.Bytes()))

	_, err = encodeFrameHeader24(&f.FrameHeader, buf)
	if err != nil {
		return err
	}

	_, err = io.Copy(buf, tmp)
	return err
}
