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

func (c *codec23) decodeFrame(f *Frame, r io.Reader) (int, error) {
	return 0, ErrUnimplemented
}

func (c *codec23) encodeFrame(f *Frame, w io.Writer) (int, error) {
	return 0, ErrUnimplemented
}
