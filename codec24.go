package id3

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

//
// codec24
//

type codec24 struct {
}

var frameCodecs = map[string]frameCodec{
	"T":    &frameCodecText_24{},
	"APIC": &frameCodecAPIC_24{},
}

func (c *codec24) Read(t *Tag, r *bufio.Reader) (int, error) {
	n := 0

	remain := t.Size
	for remain > 0 {
		id, err := r.Peek(4)
		if err != nil {
			return n, err
		}
	}
	return 0, nil
}

func (c *codec24) Write(t *Tag, w *bufio.Writer) (int, error) {
	return 0, nil
}

//
// v2.4 FrameHeader codec
//

func (h *FrameHeader) read24(r *bufio.Reader) (int, error) {
	buf := make([]byte, 10)
	n, err := r.Read(buf)
	if n < 10 || err != nil {
		return n, err
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
		}
	}
	return n, nil
}

func (h *FrameHeader) write24(w *bufio.Writer) (int, error) {
	nn := 0

	n, err := w.WriteString(h.IDvalue)
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
	return nn, err
}

//
// FrameText
//

type frameCodecText_24 struct {
}

func (c *frameCodecText_24) decode(buf bytes.Buffer) (Frame, error) {
	nn := 0

	n, err := f.FrameHeader.read24(r)
	nn += n
	if err != nil {
		return nn, err
	}

	enc, err := r.ReadByte()
	nn++
	if err != nil {
		return nn, err
	}
	f.Encoding = Encoding(enc)
	if f.Encoding < 0 || f.Encoding > 3 {
		return nn, ErrBadEncoding
	}

	len := int(f.FrameHeader.Size) - 10
	n, f.Text, err = readEncodedString(r, len, f.Encoding)
	n += nn
	return nn, err
}

func (c *frameCodecText_24) encode(frame Frame, buf bytes.Buffer) error {
	// Encode the text data.
	b := bytes.NewBuffer(make([]byte, 0, len(f.Text)))
	len, err := writeEncodedString(b, f.Text, f.Encoding)
	if err != nil {
		return 0, err
	}

	// Update the frame size based on the string's encoded length.
	f.FrameHeader.Size = uint32(len) + 1

	nn := 0

	n, err := f.FrameHeader.write24(w)
	nn += n
	if err != nil {
		return nn, err
	}

	err = w.WriteByte(byte(f.Encoding))
	if err != nil {
		return nn, err
	}
	nn++

	n, err = w.Write(b.Bytes())
	nn += n
	return nn, nil
}

//
// frameCodecAPIC_24
//

type frameCodecAPIC_24 struct {
}

func (c *frameCodecAPIC_24) decode(buf bytes.Buffer) (Frame, error) {
	return nil, nil
}

func (c *frameCodecAPIC_24) encode(frame Frame, buf bytes.Buffer) error {
	return nil
}
