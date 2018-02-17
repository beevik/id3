package id3

import (
	"io"
)

type codec23 struct {
}

func newCodec23() *codec23 {
	return &codec23{}
}

func (c *codec23) Decode(t *Tag, r *reader) (int, error) {
	return 0, errUnimplemented
}

func (c *codec23) Encode(t *Tag, w *writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec23) EncodeHeader(t *Tag, w io.Writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec23) EncodeExtendedHeader(t *Tag, w io.Writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec23) EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error) {
	return 0, errUnimplemented
}
