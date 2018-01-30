package id3

import "io"

type codec interface {
	HeaderFlags() flagMap
	DecodeExtendedHeader(t *Tag, r io.Reader) (int, error)
	DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error)
	EncodeFrame(t *Tag, f *Frame, w io.Writer) (int, error)
}
