package id3

import (
	"io"
	"reflect"
	"strings"
)

// Decode a sync-safe uint32 from a byte slice containing 4 or 5 bytes.
func decodeSyncSafeUint32(b []byte) (value uint32, err error) {
	l := len(b)
	if l < 4 || l > 5 {
		return 0, ErrBadSync
	}

	var tmp uint64
	for i := 0; i < l; i++ {
		if (b[i] & 0x80) != 0 {
			return 0, ErrBadSync
		}
		tmp = (tmp << 7) | uint64(b[i])
	}
	return uint32(tmp), nil
}

// Encode a sync-safe uint32 into a byte slice containing 4 or 5 bytes.
func encodeSyncSafeUint32(b []byte, value uint32) error {
	l := len(b)
	if l < 4 || l > 5 {
		return ErrBadSync
	}
	if l == 4 && value > 0x0fffffff {
		return ErrBadSync
	}

	for i := l - 1; i >= 0; i-- {
		b[i] = uint8(value & 0x7f)
		value = value >> 7
	}
	return nil
}

// Read a single byte from a reader.
func readByte(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	return buf[0], err
}

// Write a single byte to a writer.
func writeByte(w io.Writer, b byte) error {
	buf := make([]byte, 1)
	buf[0] = b
	_, err := w.Write(buf)
	return err
}

type tagMap map[string]bool

func (t tagMap) Lookup(s string) bool {
	_, ok := t[s]
	return ok
}

var emptyTags = make(tagMap)

// Return a table of all tags on a struct field.
func getTags(f reflect.StructField, key string) tagMap {
	tag, ok := f.Tag.Lookup(key)
	if !ok {
		return emptyTags
	}

	m := make(tagMap)
	for _, t := range strings.Split(tag, ",") {
		m[t] = true
	}
	return m
}
