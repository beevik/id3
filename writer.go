package id3

import (
	"io"
)

// A writer represents a buffer to which data is added. After adding
// data to the writer, it may be stored to a stream.
type writer struct {
	w   io.Writer
	buf []byte
	n   int
	err error
}

func newWriter(w io.Writer) *writer {
	return &writer{w: w, buf: make([]byte, 0, 64)}
}

// Len returns the number of unsaved bytes in the writer's buffer.
func (w *writer) Len() int {
	return len(w.buf)
}

// Bytes returns the current contents of the writer's buffer.  The returned
// slice is valid only until the next call to a writer function.
func (w *writer) Bytes() []byte {
	return w.buf
}

// SliceBuffer returns a slice of the reader's buffer starting at the
// offset and ending after length bytes.
func (w *writer) SliceBuffer(offset int, length int) []byte {
	return w.buf[offset : offset+length]
}

// ConsumeBytesFromOffset consumes all bytes in the writer's output buffer
// starting from the offset. It returns the consumed bytes.
func (w *writer) ConsumeBytesFromOffset(offset int) []byte {
	b := w.buf[offset:]
	w.buf = w.buf[:offset]
	return b
}

// SaveTo writes all unsaved bytes in the writer's buffer to the stream.
func (w *writer) Save() (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	var n int
	n, w.err = w.w.Write(w.buf)
	w.n += n
	return n, w.err
}

// StoreByte adds a single byte to the writer's buffer.
func (w *writer) StoreByte(b byte) {
	if w.err != nil {
		return
	}

	w.buf = append(w.buf, b)
}

// StoreBytes adds a slice of bytes to the writer's buffer.
func (w *writer) StoreBytes(b []byte) {
	if w.err != nil {
		return
	}

	w.buf = append(w.buf, b...)
}

// StoreStrings adds a series of encoded, null-terminated strings to the
// writer's buffer.
func (w *writer) StoreStrings(ss []string, enc Encoding) {
	if w.err != nil {
		return
	}

	b, err := encodeStrings(ss, enc)
	if err != nil {
		w.err = err
		return
	}

	w.buf = append(w.buf, b...)
}

// StoreFixedLengthString adds a string of known length to the writer's
// buffer.
func (w *writer) StoreFixedLengthString(s string, n int, enc Encoding) {
	if w.err != nil {
		return
	}

	if len(s) != n {
		w.err = ErrInvalidFixedLenString
		return
	}

	b, err := encodeString(s, enc)
	if err != nil {
		w.err = err
		return
	}

	w.buf = append(w.buf, b...)
}

// StoreString adds an encoded string to the writer's buffer. If requested,
// it will be null-terminated.
func (w *writer) StoreString(s string, enc Encoding, term bool) {
	if w.err != nil {
		return
	}

	b, err := encodeString(s, enc)
	if err != nil {
		w.err = err
		return
	}

	if term {
		b = append(b, null[enc]...)
	}

	w.buf = append(w.buf, b...)
}
