package id3

import (
	"io"
)

// An unsyncWriter can be used to insert "unsync codes" into a stream of bytes.
type unsyncWriter struct {
	writer   io.Writer
	prevbyte uint8
}

func needsUnsync(ch byte) bool {
	// A 11111111 followed by a 111xxxxx (or 00000000) requires an unsync code
	// between the two bytes.
	return ch == 0 || (ch&0xe0) == 0xe0
}

func newUnsyncWriter(b io.Writer) *unsyncWriter {
	return &unsyncWriter{b, 0}
}

func (w *unsyncWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Write runs of bytes that don't require "unsynchronizing". Between
	// runs, write a zero. Use "left" and "right" to keep track of the
	// bounds of the current run.
	var sn int
	left := 0
	for right := 0; right < len(p); right++ {
		if w.prevbyte == 0xff && needsUnsync(p[right]) {
			// Write the current run.
			sn, err = w.writer.Write(p[left:right])
			n += sn
			if err != nil {
				return n, err
			}

			// Insert 00000000 between 11111111 and 111xxxxx
			sn, err = w.writer.Write([]byte{0, p[right]})
			n += sn
			if err != nil {
				return n, err
			}

			// Start tracking a new run.
			left = right + 1
			w.prevbyte = 0
		} else {
			w.prevbyte = p[right]
		}
	}

	// Write the remainder run.
	sn, err = w.writer.Write(p[left:])
	n += sn
	return n, err
}

// An unsyncReader can be used to remove "unsync codes" from a stream of
// bytes.
type unsyncReader struct {
	reader   io.Reader
	prevbyte uint8
}

func newUnsyncReader(b io.Reader) *unsyncReader {
	return &unsyncReader{b, 0}
}

func (r *unsyncReader) Read(p []byte) (int, error) {
	// Fill the read buffer.
	n, err := r.reader.Read(p)
	if err != nil || n == 0 {
		return n, err
	}

	// Keep track of the last character read. If it's 0xff, the next Read
	// will need to check if the first byte is an unsync code.
	prevbyte := r.prevbyte
	if n > 0 {
		r.prevbyte = p[n-1]
	} else {
		r.prevbyte = 0
	}

	// If the previous read ended in 0xff, check if the first byte of this
	// read was an unsync code (0x00). If so, remove it.
	if prevbyte == 0xff && p[0] == 0x00 {
		copy(p[0:], p[1:n])
		n--
	}

	// Scan the buffer for 0xff followed by 0x00. Remove the 0x00 whenever
	// such a sequence is discovered.
	for i := 0; i < n-1; i++ {
		if p[i] == 0xff && p[i+1] == 0x00 {
			copy(p[i+1:], p[i+2:n])
			n--
		}
	}

	return n, nil
}
