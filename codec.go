package id3

import "bufio"

type codec interface {
	decode(t *Tag, r *bufio.Reader) (int, error)
	encode(t *Tag, w *bufio.Writer) (int, error)
}
