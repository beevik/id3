package id3

import (
	"bytes"
	"io"
	"strings"
)

//
// input
//

type ibuf struct {
	buf []byte
	n   int
	err error
}

func (i *ibuf) Len() int {
	return len(i.buf)
}

func (i *ibuf) Read(r io.Reader, n int) {
	i.buf = make([]byte, n)
	_, i.err = io.ReadFull(r, i.buf)
	i.n += n
}

func (i *ibuf) Replace(buf []byte) {
	i.buf = buf
}

func (i *ibuf) ConsumeByte() byte {
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

func (i *ibuf) ConsumeBytes(n int) []byte {
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

func (i *ibuf) ConsumeStrings(enc Encoding) []string {
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

func (i *ibuf) ConsumeFixedLengthString(len int, enc Encoding) string {
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

func (i *ibuf) ConsumeNextString(enc Encoding) string {
	var str string

	if i.err != nil {
		return str
	}

	str, i.buf, i.err = decodeNextString(i.buf, enc)
	return str
}

func (i *ibuf) ConsumeAll() []byte {
	if i.err != nil {
		return []byte{}
	}

	b := i.buf
	i.buf = i.buf[:0]
	return b
}

//
// output
//

type obuf struct {
	buf *bytes.Buffer
	n   int
	err error
}

func newOutput() *obuf {
	return &obuf{
		buf: bytes.NewBuffer([]byte{}),
		n:   0,
		err: nil,
	}
}

func (o *obuf) Len() int {
	return len(o.buf.Bytes())
}

func (o *obuf) Bytes() []byte {
	return o.buf.Bytes()
}

func (o *obuf) WriteByte(b byte) error {
	if o.err != nil {
		return o.err
	}

	o.err = o.buf.WriteByte(b)
	if o.err == nil {
		o.n++
	}
	return o.err
}

func (o *obuf) WriteBytes(b []byte) error {
	if o.err != nil {
		return o.err
	}

	n, err := o.buf.Write(b)
	o.n += n
	o.err = err
	return o.err
}

func (o *obuf) WriteStrings(ss []string, enc Encoding) error {
	if o.err != nil {
		return o.err
	}

	b, err := encodeStrings(ss, enc)
	if err != nil {
		o.err = err
		return o.err
	}

	n, err := o.buf.Write(b)
	o.n += n
	o.err = err
	return o.err
}

func (o *obuf) WriteFixedLengthString(s string, n int, enc Encoding) error {
	if o.err != nil {
		return o.err
	}

	if len(s) != n {
		o.err = ErrInvalidFixedLenString
		return o.err
	}

	b, err := encodeString(s, enc)
	if err != nil {
		o.err = err
		return o.err
	}

	n, err = o.buf.Write(b)
	o.n += n
	o.err = err
	return o.err
}

func (o *obuf) WriteString(s string, enc Encoding, term bool) error {
	if o.err != nil {
		return o.err
	}

	b, err := encodeString(s, enc)
	if err != nil {
		o.err = err
		return o.err
	}

	if term {
		b = append(b, null[enc]...)
	}

	n, err := o.buf.Write(b)
	o.n += n
	o.err = err
	return o.err
}
