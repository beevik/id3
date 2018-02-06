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

func (c *codec22) HeaderFlags() flagMap {
	return flagMap{}
}

func (c *codec22) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec22) DecodeFrame(t *Tag, f *FrameHolder, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec22) EncodeFrame(t *Tag, f *FrameHolder, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
