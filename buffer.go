package id3

import (
	"io"
	"strings"
)

// The buffer may be used for reading or writing data. It tracks its
// errors and total bytes read/written internally. A buffer instance should be
// used only for reading or writing, never both.
type buffer struct {
	buf []byte
	n   int
	err error
}

func newBuffer() *buffer {
	return &buffer{buf: make([]byte, 0)}
}

func (b *buffer) Len() int {
	return len(b.buf)
}

func (b *buffer) Bytes() []byte {
	return b.buf
}

func (b *buffer) Read(r io.Reader, n int) {
	b.buf = make([]byte, n)
	_, b.err = io.ReadFull(r, b.buf)
	b.n += n
}

func (b *buffer) Replace(buf []byte) {
	b.buf = buf
}

func (b *buffer) ConsumeByte() byte {
	if b.err != nil {
		return 0
	}

	if len(b.buf) < 1 {
		b.err = errInsufficientBuffer
		return 0
	}

	bb := b.buf[0]
	b.buf = b.buf[1:]
	return bb
}

func (b *buffer) ConsumeBytes(n int) []byte {
	if b.err != nil {
		return make([]byte, n)
	}

	if len(b.buf) < n {
		b.err = errInsufficientBuffer
		return make([]byte, n)
	}

	bb := b.buf[:n]
	b.buf = b.buf[n:]
	return bb
}

func (b *buffer) ConsumeStrings(enc Encoding) []string {
	if b.err != nil {
		return []string{}
	}

	var ss []string
	ss, b.err = decodeStrings(b.buf, enc)
	if b.err != nil {
		return ss
	}

	b.buf = b.buf[:0]
	return ss
}

func (b *buffer) ConsumeFixedLengthString(len int, enc Encoding) string {
	if b.err != nil {
		return strings.Repeat("_", len)
	}

	bb := b.ConsumeBytes(len)
	if b.err != nil {
		return strings.Repeat("_", len)
	}

	var str string
	str, b.err = decodeString(bb, EncodingISO88591)
	return str
}

func (b *buffer) ConsumeNextString(enc Encoding) string {
	var str string

	if b.err != nil {
		return str
	}

	str, b.buf, b.err = decodeNextString(b.buf, enc)
	return str
}

func (b *buffer) ConsumeAll() []byte {
	if b.err != nil {
		return []byte{}
	}

	bb := b.buf
	b.buf = b.buf[:0]
	return bb
}

func (b *buffer) Write(w io.Writer) {
	var n int
	n, b.err = w.Write(b.buf)
	b.n += n
	b.buf = b.buf[:0]
}

func (b *buffer) AddByte(bb byte) {
	if b.err != nil {
		return
	}

	b.buf = append(b.buf, bb)
}

func (b *buffer) AddBytes(bb []byte) {
	if b.err != nil {
		return
	}

	b.buf = append(b.buf, bb...)
}

func (b *buffer) AddStrings(ss []string, enc Encoding) {
	if b.err != nil {
		return
	}

	bb, err := encodeStrings(ss, enc)
	if err != nil {
		b.err = err
		return
	}

	b.buf = append(b.buf, bb...)
}

func (b *buffer) AddFixedLengthString(s string, n int, enc Encoding) {
	if b.err != nil {
		return
	}

	if len(s) != n {
		b.err = ErrInvalidFixedLenString
		return
	}

	bb, err := encodeString(s, enc)
	if err != nil {
		b.err = err
		return
	}

	b.buf = append(b.buf, bb...)
}

func (b *buffer) AddString(s string, enc Encoding, term bool) {
	if b.err != nil {
		return
	}

	bb, err := encodeString(s, enc)
	if err != nil {
		b.err = err
		return
	}

	if term {
		bb = append(bb, null[enc]...)
	}

	b.buf = append(b.buf, bb...)
}
