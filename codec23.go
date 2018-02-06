package id3

import (
	"io"
)

type codec23 struct {
	payloadTypes typeMap
}

func newCodec23() *codec23 {
	return &codec23{
		payloadTypes: newTypeMap("v23"),
	}
}

func (c *codec23) HeaderFlags() flagMap {
	return flagMap{}
}

func (c *codec23) DecodeExtendedHeader(t *Tag, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec23) DecodeFrame(t *Tag, f *FrameHolder, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec23) EncodeFrame(t *Tag, f *FrameHolder, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
