package id3

import (
	"io"
)

type codec23 struct {
}

func (c *codec23) Read(t *Tag, r io.Reader) (int64, error) {
	return 0, nil
}

func (c *codec23) Write(t *Tag, w io.Writer) (int64, error) {
	return 0, nil
}
