package id3

import (
	"io"
)

type codec22 struct {
}

func newCodec22() *codec22 {
	return &codec22{}
}

func (c *codec22) Decode(t *Tag, r *reader) (int, error) {
	return 0, errUnimplemented
}

func (c *codec22) Encode(t *Tag, w *writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec22) EncodeHeader(t *Tag, w io.Writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec22) EncodeExtendedHeader(t *Tag, w io.Writer) (int, error) {
	return 0, errUnimplemented
}

func (c *codec22) EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error) {
	return 0, errUnimplemented
}
