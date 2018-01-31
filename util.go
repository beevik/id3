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

//
// tagList
//

type tagList map[string]bool

func (t tagList) Lookup(s string) bool {
	_, ok := t[s]
	return ok
}

var emptyTagList = make(tagList)

// Return a table of all tags on a struct field.
func getTags(f reflect.StructField, key string) tagList {
	tag, ok := f.Tag.Lookup(key)
	if !ok {
		return emptyTagList
	}

	l := make(tagList)
	for _, t := range strings.Split(tag, ",") {
		l[t] = true
	}
	return l
}

//
// flagMap
//

// A flag map maps encoded and decoded versions of flags to one another.
type flagMap []struct {
	encoded uint32 // the encoded representation of a flag
	decoded uint32 // the decoded representation of a flag
}

// Return the decoded representation of the encoded flags.
func (f flagMap) Decode(flags uint32) uint32 {
	if flags == 0 {
		return 0
	}
	var result uint32
	for _, e := range f {
		if (e.encoded & flags) != 0 {
			result |= e.decoded
		}
	}
	return result
}

// Return the encoded representation of the decoded flags.
func (f flagMap) Encode(flags uint32) uint32 {
	if flags == 0 {
		return 0
	}
	var result uint32
	for _, e := range f {
		if (e.decoded & flags) != 0 {
			result |= e.encoded
		}
	}
	return result
}
