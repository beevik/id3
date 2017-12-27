package id3

import (
	"io"
)

type codec22 struct {
}

func (c *codec22) Read(t *Tag, r io.Reader) (int64, error) {
	return 0, nil
}
