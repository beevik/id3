package id3

import (
	"bytes"
)

func addUnsyncCodes(buf []byte) []byte {
	if len(buf) == 0 {
		return buf
	}

	out := bytes.NewBuffer(make([]byte, 0, len(buf)))
	out.WriteByte(buf[0])
	prev := buf[0]
	l := len(buf)
	for i := 1; i < l; i++ {
		if prev == 0xff && (buf[i] == 0 || (buf[i]&0xe0) == 0xe0) {
			out.WriteByte(0)
			out.WriteByte(buf[i])
			prev = 0
		} else {
			out.WriteByte(buf[i])
			prev = buf[i]
		}
	}
	return out.Bytes()
}

func removeUnsyncCodes(buf []byte) []byte {
	if len(buf) == 0 {
		return buf
	}

	out := bytes.NewBuffer(make([]byte, 0, len(buf)))
	out.WriteByte(buf[0])
	prev := buf[0]
	l := len(buf)
	for i := 1; i < l; i++ {
		if prev != 0xff || buf[i] != 0 {
			out.WriteByte(buf[i])
		}
		prev = buf[i]
	}
	return out.Bytes()
}
