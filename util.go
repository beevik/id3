package id3

import (
	"io"
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

// Read an encoded text string using the specified encoding.
func readEncodedString(r io.Reader, len int, enc Encoding) (n int, s string, err error) {
	buf := make([]byte, len)
	n, err = r.Read(buf)
	if err != nil {
		return n, "", err
	}
	if n < len {
		return n, "", ErrBadText
	}

	switch enc {
	case EncodingISO88591:
		runes := make([]rune, len)
		for i := range buf {
			runes[i] = rune(buf[i])
		}
		return n, string(runes), nil

	case EncodingUTF16BOM:
		if len < 2 || buf[0] != 0xfe || buf[1] != 0xff {
			return n, "", ErrBadText
		}
		buf = buf[2:]
		len -= 2
		fallthrough

	case EncodingUTF16:
		if (len & 1) != 0 {
			return n, "", ErrBadText
		}
		u := make([]uint16, len/2)
		for i, j := 0, 0; i < len; i, j = i+2, j+1 {
			u[j] = uint16(buf[i])<<8 | uint16(buf[i+1])
		}
		return n, string(utf16.Decode(u)), nil

	case EncodingUTF8:
		if !utf8.Valid(buf) {
			return n, "", ErrBadText
		}
		return n, string(buf), nil

	default:
		return n, "", ErrBadText
	}
}

func writeEncodedString(w io.Writer, s string, enc Encoding) (n int, err error) {
	switch enc {
	case EncodingISO88591:
		b := make([]byte, 0, len(s))
		for _, r := range s {
			if r > 0xff {
				r = '.'
			}
			b = append(b, byte(r))
		}
		return w.Write(b)

	case EncodingUTF16BOM:
		n, err = w.Write([]byte{0xfe, 0xff})
		if err != nil {
			return n, ErrBadText
		}
		fallthrough

	case EncodingUTF16:
		u := utf16.Encode([]rune(s))
		b := make([]byte, len(u)*2)
		for i, j := 0, 0; i < len(u); i, j = i+1, j+2 {
			b[j] = byte(u[i] >> 8)
			b[j+1] = byte(u[i])
		}
		nn, err := w.Write(b)
		n += nn
		return n, err

	case EncodingUTF8:
		return w.Write([]byte(s))

	default:
		return 0, ErrBadText
	}
}
