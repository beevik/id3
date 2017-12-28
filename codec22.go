package id3

import (
	"bufio"
)

type codec22 struct {
}

func (c *codec22) Read(t *Tag, r *bufio.Reader) (int, error) {
	return 0, nil
}

func (c *codec22) Write(t *Tag, w *bufio.Writer) (int, error) {
	return 0, nil
}

func (h *FrameHeader) read22(r *bufio.Reader) (int, error) {
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

func (h *FrameHeader) write22(w *bufio.Writer) (int, error) {
	nn := 0

	n, err := w.WriteString(h.IDvalue)
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
