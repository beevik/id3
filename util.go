package id3

import (
	"fmt"
	"io"
	"reflect"
)

func decodeUint32(b []byte) uint32 {
	if len(b) != 4 {
		panic("invalid uint32 size")
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

// Decode a sync-safe uint32 from a byte slice containing 4 or 5 bytes.
func decodeSyncSafeUint32(b []byte) (value uint32, err error) {
	l := len(b)
	if l < 4 || l > 5 {
		return 0, ErrInvalidSync
	}

	var tmp uint64
	for i := 0; i < l; i++ {
		if (b[i] & 0x80) != 0 {
			return 0, ErrInvalidSync
		}
		tmp = (tmp << 7) | uint64(b[i])
	}
	return uint32(tmp), nil
}

func encodeUint32(b []byte, value uint32) {
	if len(b) != 4 {
		panic("invalid uint32 size")
	}
	b[0] = byte(value >> 24)
	b[1] = byte(value >> 16)
	b[2] = byte(value >> 8)
	b[3] = byte(value)
}

// Encode a sync-safe uint32 into a byte slice containing 4 or 5 bytes.
func encodeSyncSafeUint32(b []byte, value uint32) error {
	l := len(b)
	if l < 4 || l > 5 {
		return ErrInvalidSync
	}
	if l == 4 && value > 0x0fffffff {
		return ErrInvalidSync
	}

	for i := l - 1; i >= 0; i-- {
		b[i] = uint8(value & 0x7f)
		value = value >> 7
	}
	return nil
}

func hexdump(b []byte, w io.Writer) {
	fmt.Fprintf(w, "var b = []byte{\n")

	for i := 0; i < len(b); i += 8 {
		r := i + 8
		if r > len(b) {
			r = len(b)
		}

		fmt.Fprintf(w, "\t")

		var j int
		for j = i; j < r-1; j++ {
			fmt.Fprintf(w, "0x%02x, ", b[j])
		}
		fmt.Fprintf(w, "0x%02x,\n", b[j])
	}

	fmt.Fprintf(w, "}\n")
}

//
// valueStack
//

type valueStack struct {
	stack []reflect.Value
}

func (v *valueStack) pop() reflect.Value {
	n := len(v.stack) - 1
	ret := v.stack[n]
	v.stack = v.stack[:n]
	return ret
}

func (v *valueStack) push(rv reflect.Value) {
	v.stack = append(v.stack, rv)
}

func (v *valueStack) first() reflect.Value {
	return v.stack[0]
}

func (v *valueStack) depth() int {
	return len(v.stack)
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

//
// boundsMap
//

type boundsMap map[string]struct {
	min int
	max int
	err error
}
