package id3

import (
	"io"
	"strings"
)

type scanner struct {
	buf []byte
	n   int
	err error
}

func (s *scanner) Len() int {
	return len(s.buf)
}

func (s *scanner) Read(r io.Reader, n int) {
	s.buf = make([]byte, n)
	_, s.err = io.ReadFull(r, s.buf)
	s.n += n
}

func (s *scanner) Replace(buf []byte) {
	s.buf = buf
}

func (s *scanner) ConsumeByte() byte {
	if s.err != nil {
		return 0
	}

	if len(s.buf) < 1 {
		s.err = errInsufficientBuffer
		return 0
	}

	b := s.buf[0]
	s.buf = s.buf[1:]
	return b
}

func (s *scanner) ConsumeBytes(n int) []byte {
	if s.err != nil {
		return make([]byte, n)
	}

	if len(s.buf) < n {
		s.err = errInsufficientBuffer
		return make([]byte, n)
	}

	b := s.buf[:n]
	s.buf = s.buf[n:]
	return b
}

func (s *scanner) ConsumeStrings(enc Encoding) []string {
	if s.err != nil {
		return []string{}
	}

	var ss []string
	ss, s.err = decodeStrings(s.buf, enc)
	if s.err != nil {
		return ss
	}

	s.buf = s.buf[:0]
	return ss
}

func (s *scanner) ConsumeFixedLenString(len int, enc Encoding) string {
	if s.err != nil {
		return strings.Repeat("_", len)
	}

	b := s.ConsumeBytes(len)
	if s.err != nil {
		return strings.Repeat("_", len)
	}

	var str string
	str, s.err = decodeString(b, EncodingISO88591)
	return str
}

func (s *scanner) ConsumeNextString(enc Encoding) string {
	var str string

	if s.err != nil {
		return str
	}

	str, s.buf, s.err = decodeNextString(s.buf, enc)
	return str
}

func (s *scanner) ConsumeAll() []byte {
	if s.err != nil {
		return []byte{}
	}

	b := s.buf
	s.buf = s.buf[:0]
	return b
}
