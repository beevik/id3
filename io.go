package id3

import (
	"bytes"
	"io"
	"strings"
)

//
// inputBuf
//

type inputBuf struct {
	buf []byte
	n   int
	err error
}

func newInputBuf() *inputBuf {
	return &inputBuf{}
}

func (i *inputBuf) Len() int {
	return len(i.buf)
}

func (i *inputBuf) Read(r io.Reader, n int) {
	i.buf = make([]byte, n)
	_, i.err = io.ReadFull(r, i.buf)
	i.n += n
}

func (i *inputBuf) Replace(buf []byte) {
	i.buf = buf
}

func (i *inputBuf) ConsumeByte() byte {
	if i.err != nil {
		return 0
	}

	if len(i.buf) < 1 {
		i.err = errInsufficientBuffer
		return 0
	}

	b := i.buf[0]
	i.buf = i.buf[1:]
	return b
}

func (i *inputBuf) ConsumeBytes(n int) []byte {
	if i.err != nil {
		return make([]byte, n)
	}

	if len(i.buf) < n {
		i.err = errInsufficientBuffer
		return make([]byte, n)
	}

	b := i.buf[:n]
	i.buf = i.buf[n:]
	return b
}

func (i *inputBuf) ConsumeStrings(enc Encoding) []string {
	if i.err != nil {
		return []string{}
	}

	var ss []string
	ss, i.err = decodeStrings(i.buf, enc)
	if i.err != nil {
		return ss
	}

	i.buf = i.buf[:0]
	return ss
}

func (i *inputBuf) ConsumeFixedLengthString(len int, enc Encoding) string {
	if i.err != nil {
		return strings.Repeat("_", len)
	}

	b := i.ConsumeBytes(len)
	if i.err != nil {
		return strings.Repeat("_", len)
	}

	var str string
	str, i.err = decodeString(b, EncodingISO88591)
	return str
}

func (i *inputBuf) ConsumeNextString(enc Encoding) string {
	var str string

	if i.err != nil {
		return str
	}

	str, i.buf, i.err = decodeNextString(i.buf, enc)
	return str
}

func (i *inputBuf) ConsumeAll() []byte {
	if i.err != nil {
		return []byte{}
	}

	b := i.buf
	i.buf = i.buf[:0]
	return b
}

//
// outputBuf
//

type outputBuf struct {
	buf *bytes.Buffer
	n   int
	err error
}

func newOutputBuf() *outputBuf {
	return &outputBuf{
		buf: bytes.NewBuffer([]byte{}),
		n:   0,
		err: nil,
	}
}

func (o *outputBuf) Len() int {
	return len(o.buf.Bytes())
}

func (o *outputBuf) Bytes() []byte {
	return o.buf.Bytes()
}

func (o *outputBuf) AddByte(b byte) {
	if o.err != nil {
		return
	}

	o.err = o.buf.WriteByte(b)
	if o.err == nil {
		o.n++
	}
}

func (o *outputBuf) AddBytes(b []byte) {
	if o.err != nil {
		return
	}

	n, err := o.buf.Write(b)
	o.n += n
	o.err = err
}

func (o *outputBuf) AddStrings(ss []string, enc Encoding) {
	if o.err != nil {
		return
	}

	b, err := encodeStrings(ss, enc)
	if err != nil {
		o.err = err
		return
	}

	n, err := o.buf.Write(b)
	o.n += n
	o.err = err
}

func (o *outputBuf) AddFixedLengthString(s string, n int, enc Encoding) {
	if o.err != nil {
		return
	}

	if len(s) != n {
		o.err = ErrInvalidFixedLenString
		return
	}

	b, err := encodeString(s, enc)
	if err != nil {
		o.err = err
		return
	}

	n, err = o.buf.Write(b)
	o.n += n
	o.err = err
}

func (o *outputBuf) AddString(s string, enc Encoding, term bool) {
	if o.err != nil {
		return
	}

	b, err := encodeString(s, enc)
	if err != nil {
		o.err = err
		return
	}

	if term {
		b = append(b, null[enc]...)
	}

	n, err := o.buf.Write(b)
	o.n += n
	o.err = err
}
