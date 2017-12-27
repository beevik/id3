package id3

import (
	"io"
)

type unsyncReader struct {
	reader   io.Reader
	prevbyte uint8
}

func newUnsyncReader(r io.Reader) *unsyncReader {
	return &unsyncReader{r, 0}
}

func (r *unsyncReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil || n == 0 {
		return n, err
	}

	if r.prevbyte == 0xff && p[0] == 0x00 {
		copy(p[0:], p[1:n])
		n--
		r.prevbyte = 0
	}

	for i := 0; i < n-1; i++ {
		if p[i] == 0xff && p[i+1] == 0x00 {
			copy(p[i+1:], p[i+2:n])
			n--
		}
	}

	if n > 0 {
		r.prevbyte = p[n-1]
	}

	return n, nil
}
