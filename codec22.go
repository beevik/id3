package id3

import (
	"io"
)

type codec22 struct {
	payloadTypes typeMap
}

func newCodec22() *codec22 {
	return &codec22{
		payloadTypes: newTypeMap("v22"),
	}
}

func (c *codec22) decodeHeaderFlags(flags uint8) uint8 {
	return 0
}

func (c *codec22) decodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec22) decodeFrame(f *Frame, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec22) encodeFrame(f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
