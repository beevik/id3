package id3

import (
	"io"
)

type codec22 struct {
}

func (c *codec22) decode(t *Tag, r io.Reader) (int, error) {
	return 0, nil
}

func (c *codec22) encode(t *Tag, w io.Writer) (int, error) {
	return 0, nil
}

func (h *FrameHeader) read22(r io.Reader) (int, error) {
	buf := make([]byte, 6)
	n, err := r.Read(buf)
	if n < 6 || err != nil {
		return n, err
	}

	h.IDvalue = string(buf[0:3])
	h.Size = uint32(buf[3])<<16 + uint32(buf[4])<<8 + uint32(buf[5])
	h.Flags = 0

	return n, nil
}

func (h *FrameHeader) write22(w io.Writer) (int, error) {
	nn := 0

	idval := []byte(h.IDvalue)
	n, err := w.Write(idval)
	nn += n
	if err != nil {
		return nn, err
	}

	size := make([]byte, 3)
	size[0] = uint8(h.Size >> 16)
	size[1] = uint8(h.Size >> 8)
	size[2] = uint8(h.Size)
	n, err = w.Write(size)
	nn += n

	return nn, err
}
