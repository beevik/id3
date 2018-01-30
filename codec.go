package id3

import "io"

type codec interface {
	HeaderFlags() flagMap
	DecodeExtendedHeader(t *Tag, r io.Reader) (int, error)
	DecodeFrame(f *Frame, r io.Reader) (int, error)
	EncodeFrame(f *Frame, w io.Writer) (int, error)
}
