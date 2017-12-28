package id3

import "bufio"

type codec interface {
	Read(t *Tag, r *bufio.Reader) (int, error)
	Write(t *Tag, w *bufio.Writer) (int, error)
}
