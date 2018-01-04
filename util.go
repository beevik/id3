package id3

import (
	"unicode/utf16"
	"unicode/utf8"
)

// Read a sync-safe uint32 from a buffer of 4 or 5 bytes.
func readSyncSafeUint32(b []byte) (value uint32, err error) {
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

// Write a sync-safe uint32 to a buffer of 4 or 5 bytes.
func writeSyncSafeUint32(b []byte, value uint32) error {
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

// Decode a string stored in a byte slice.
func decodeString(b []byte, enc Encoding) (string, error) {
	switch enc {
	case EncodingISO88591:
		runes := make([]rune, 0, len(b))
		for _, c := range b {
			if c == 0 {
				break
			}
			runes = append(runes, rune(c))
		}
		return string(runes), nil

	case EncodingUTF16BOM:
		fallthrough

	case EncodingUTF16:
		if len(b) >= 2 && b[0] == 0xfe && b[1] == 0xff {
			b = b[2:]
		}
		if (len(b) & 1) != 0 {
			return "", ErrBadText
		}
		u := make([]uint16, len(b)/2)
		for i, j := 0, 0; i < len(b); i, j = i+2, j+1 {
			u[j] = uint16(b[i])<<8 | uint16(b[i+1])
			if u[j] == 0 {
				u = u[:j]
				break
			}
		}
		return string(utf16.Decode(u)), nil

	case EncodingUTF8:
		for i := 0; i < len(b); i++ {
			if b[i] == 0 {
				b = b[:i]
				break
			}
		}
		if !utf8.Valid(b) {
			return "", ErrBadText
		}
		return string(b), nil

	default:
		return "", ErrBadText
	}
}

// Encode a string to a byte slice.
func encodeString(s string, enc Encoding) ([]byte, error) {
	var b []byte

	switch enc {
	case EncodingISO88591:
		b = make([]byte, 0, len(s))
		for _, r := range s {
			if r > 0xff {
				r = '.'
			}
			b = append(b, byte(r))
		}
		return b, nil

	case EncodingUTF16BOM:
		b = make([]byte, 0, len(s)*2)
		b = append(b, []byte{0xfe, 0xff}...)
		fallthrough

	case EncodingUTF16:
		u := utf16.Encode([]rune(s))
		for i, j := 0, 0; i < len(u); i, j = i+1, j+2 {
			b = append(b, []byte{byte(u[i] >> 8), byte(u[i])}...)
		}
		return b, nil

	case EncodingUTF8:
		return []byte(s), nil

	default:
		return nil, ErrBadText
	}
}
