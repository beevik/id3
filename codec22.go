package id3

import (
	"io"
)

type codec22 struct {
}

func newCodec22() *codec22 {
	return &codec22{}
}

func (c *codec22) DecodeHeader(t *Tag, r io.Reader) (int, error) {
	return 0, errUnimplemented
}

func (c *codec22) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	return 0, errUnimplemented
}

func (c *codec22) DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error) {
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
