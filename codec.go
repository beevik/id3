package id3

import "io"

type codec interface {
	decodeFrame(f *Frame, r io.Reader) (int, error)
	encodeFrame(f *Frame, w io.Writer) (int, error)
}
