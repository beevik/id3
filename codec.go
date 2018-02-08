package id3

import "io"

// Version defines the ID3 codec version (2.2, 2.3, or 2.4).
type Version uint8

const (
	v22 Version = 2 + iota // v2.2
	v23                    // v2.3
	v24                    // v2.4
)

type codec interface {
	HeaderFlags() flagMap
	DecodeExtendedHeader(t *Tag, r io.Reader) (int, error)
	DecodeFrame(t *Tag, f *FrameHolder, r io.Reader) (int, error)
	EncodeFrame(t *Tag, f *FrameHolder, w io.Writer) (int, error)
}
