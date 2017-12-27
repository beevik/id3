package id3

import "io"

type codec24 struct {
}

func (c *codec24) Read(t *Tag, r io.Reader) (int64, error) {
	return 0, nil
}
