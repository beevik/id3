package id3

import "io"

type codec interface {
	Read(t *Tag, r io.Reader) (int64, error)
}
