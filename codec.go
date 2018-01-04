package id3

import "io"

type codec interface {
	decode(t *Tag, r io.Reader) (int, error)
	encode(t *Tag, w io.Writer) (int, error)
}
