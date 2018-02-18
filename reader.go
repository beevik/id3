package id3

import (
	"io"
	"strings"
)

// A reader represents a buffer that may be consumed by the caller. The
// buffer is populated from an input stream.
type reader struct {
	r   io.Reader
	buf []byte
	n   int
	err error
}

func newReader(r io.Reader) *reader {
	return &reader{r: r, buf: make([]byte, 0, 64)}
}

// Bytes returns the contents of the reader's buffer without consuming them.
func (r *reader) Bytes() []byte {
	return r.buf
}

// Len returns the length of unread portion of the reader's buffer.
func (r *reader) Len() int {
	return len(r.buf)
}

// LoadFrom pulls exactly n bytes from a stream into the reader's buffer.
func (r *reader) Load(n int) (int, error) {
	l := len(r.buf)
	r.buf = append(r.buf, make([]byte, n)...)

	var nn int
	nn, r.err = io.ReadFull(r.r, r.buf[l:])
	r.n += nn

	if nn < n {
		r.err = io.ErrUnexpectedEOF
	}
	return nn, r.err
}

// ReplaceBuffer replaces the contents of the reader's buffer with the
// provided byte slice.
func (r *reader) ReplaceBuffer(p []byte) {
	r.buf = p
}

// ConsumeByte consumes a single byte from the reader's buffer and returns it.
func (r *reader) ConsumeByte() byte {
	if r.err != nil {
		return 0
	}
	if len(r.buf) < 1 {
		r.err = io.ErrUnexpectedEOF
		return 0
	}

	b := r.buf[0]
	r.buf = r.buf[1:]
	return b
}

// ConsumeBytes consumes exactly n bytes out of the reader's buffer and
// returns them as a byte slice.
func (r *reader) ConsumeBytes(n int) []byte {
	if r.err != nil {
		return make([]byte, n)
	}
	if len(r.buf) < n {
		r.err = io.ErrUnexpectedEOF
		return make([]byte, n)
	}

	b := r.buf[:n]
	r.buf = r.buf[n:]
	return b
}

// ConsumeFixedLengthString consumes a string of known length from the reader's
// buffer and returns it.
func (r *reader) ConsumeFixedLengthString(len int, enc Encoding) string {
	if r.err != nil {
		return strings.Repeat(" ", len)
	}

	var p []byte
	p = r.ConsumeBytes(len)
	if r.err != nil {
		return strings.Repeat("_", len)
	}

	var str string
	str, r.err = decodeString(p, enc)
	return str
}

// ConsumeNextString consumes the next null-terminated string from the reader's
// buffer and returns it.
func (r *reader) ConsumeNextString(enc Encoding) string {
	if r.err != nil {
		return ""
	}

	var str string
	str, r.buf, r.err = decodeNextString(r.buf, enc)
	return str
}

// ConsumeStrings consumes the remainder of the buffer as a series of
// null-terminated strings and returns them.
func (r *reader) ConsumeStrings(enc Encoding) []string {
	if r.err != nil {
		return []string{}
	}

	var ss []string
	ss, r.err = decodeStrings(r.buf, enc)
	if r.err != nil {
		return ss
	}

	r.buf = r.buf[:0]
	return ss
}

// ConsumeAll consumes the remaining contents of the reader's buffer
// and returns them as a byte slice.
func (r *reader) ConsumeAll() []byte {
	if r.err != nil {
		return []byte{}
	}

	p := r.buf
	r.buf = r.buf[:0]
	return p
}

// Consume exactly n bytes from the reader's buffer and place them into
// a new reader.
func (r *reader) ConsumeIntoNewReader(n int) *reader {
	if r.err != nil {
		return &reader{r: r.r, buf: nil}
	}
	if len(r.buf) < n {
		r.err = io.ErrUnexpectedEOF
		return &reader{r: r.r, buf: nil}
	}

	b := r.buf[:n]
	r.buf = r.buf[n:]
	return &reader{r: r.r, buf: b}
}
