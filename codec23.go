package id3

import (
	"io"
)

type codec23 struct {
}

func (c *codec23) Read(t *Tag, r io.Reader) (int64, error) {
	return 0, nil
}
