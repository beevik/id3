package id3

import "io"

// Version defines the ID3 codec version (2.2, 2.3, or 2.4).
type Version uint8

// Allowed ID3 codec versions
const (
	Version2_2 Version = 2 + iota // v2.2
	Version2_3                    // v2.3
	Version2_4                    // v2.4
)

type codec interface {
	HeaderFlags() flagMap
	DecodeExtendedHeader(t *Tag, r io.Reader) (int, error)
	DecodeFrame(t *Tag, f *Frame, r io.Reader) (int, error)
	EncodeExtendedHeader(t *Tag, w io.Writer) (int, error)
	EncodeFrame(t *Tag, f Frame, w io.Writer) (int, error)
}
