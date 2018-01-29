package id3

import "io"

type codec interface {
	decodeHeaderFlags(flags uint8) uint8
	decodeExtendedHeader(t *Tag, r io.Reader) (int, error)
	decodeFrame(f *Frame, r io.Reader) (int, error)
	encodeFrame(f *Frame, w io.Writer) (int, error)
}
