package id3

import "io"

type codec interface {
	Read(t *Tag, r io.Reader) (int64, error)
	Write(t *Tag, w io.Writer) (int64, error)
}
