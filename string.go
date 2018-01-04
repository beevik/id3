package id3

import (
	"unicode/utf16"
	"unicode/utf8"
)

// Encoding represents the type of encoding used on a text string with an
// ID3 frame.
type Encoding uint8

// Possible values used to indicate the type of text encoding.
const (
	EncodingISO88591 Encoding = 0
	EncodingUTF16BOM          = 1
	EncodingUTF16             = 2
	EncodingUTF8              = 3
)

// Null terminators used by each encoding.
var null = [4][]byte{
	[]byte{0},    // EncodingISO88591
	[]byte{0, 0}, // EncodingUTF16BOM
	[]byte{0, 0}, // EncodingUTF16
	[]byte{0},    // EncodingUTF8
}

// EncodedText represents an array of zero or more text strings serialized
// using one of the four available text encodings.
type EncodedText struct {
	Encoding Encoding
	Strings  []string
}

// Decode a byte buffer containing zero or more encoded, null-terminated
// strings.
func (s *EncodedText) Decode(buf []byte) error {
	if len(buf) < 1 || buf[0] > 3 {
		return ErrInvalidEncodedString
	}

	var err error
	s.Encoding = Encoding(buf[0])
	s.Strings, err = decodeStrings(buf[1:], s.Encoding)
	return err
}

// Decode a string stored in a byte slice.
func decodeString(b []byte, enc Encoding) (string, error) {
	s, _, err := decodeStringHelper(b, enc)
	return s, err
}

// Decode one or more null-terminated strings stored in a byte slice.
func decodeStrings(b []byte, enc Encoding) ([]string, error) {
	ss := make([]string, 0, 1)
	for len(b) > 0 {
		s, c, err := decodeStringHelper(b, enc)
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
		b = b[c:]
	}
	return ss, nil
}

func decodeStringHelper(b []byte, enc Encoding) (s string, consumed int, err error) {
	consumed = len(b)

	switch enc {
	case EncodingISO88591:
		runes := make([]rune, 0, len(b))
		for i, c := range b {
			if c == 0 {
				consumed = i + 1
				break
			}
			runes = append(runes, rune(c))
		}
		return string(runes), consumed, nil

	case EncodingUTF8:
		for i := 0; i < len(b); i++ {
			if b[i] == 0 {
				b = b[:i]
				consumed = i + 1
				break
			}
		}
		if !utf8.Valid(b) {
			return "", 0, ErrBadText
		}
		return string(b), consumed, nil

	case EncodingUTF16BOM:
		fallthrough

	case EncodingUTF16:
		start := 0
		if len(b) >= 2 && b[0] == 0xfe && b[1] == 0xff {
			start = 2
		}
		if (len(b) & 1) != 0 {
			return "", 0, ErrBadText
		}
		u := make([]uint16, 0, len(b)/2)
		j := 0
		for i := start; i < len(b); i += 2 {
			cp := uint16(b[i])<<8 | uint16(b[i+1])
			if cp == 0 {
				consumed = i + 2
				break
			}
			u = append(u, cp)
			j++
		}
		return string(utf16.Decode(u)), consumed, nil

	default:
		return "", 0, ErrBadText
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

	case EncodingUTF8:
		return []byte(s), nil

	case EncodingUTF16BOM:
		b = make([]byte, 0, len(s)*2)
		b = append(b, []byte{0xfe, 0xff}...)
		fallthrough

	case EncodingUTF16:
		if b == nil {
			b = make([]byte, 0, len(s)*2)
		}
		u := utf16.Encode([]rune(s))
		for i, j := 0, 0; i < len(u); i, j = i+1, j+2 {
			b = append(b, []byte{byte(u[i] >> 8), byte(u[i])}...)
		}
		return b, nil

	default:
		return nil, ErrBadText
	}
}

// Encode an array of stings to a byte slice. Separate multiple strings
// by appropriate null terminators.
func encodeStrings(ss []string, enc Encoding) ([]byte, error) {
	buf := make([]byte, 0)
	for i, s := range ss {
		b, err := encodeString(s, enc)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			buf = append(buf, null[enc]...)
		}
		buf = append(buf, b...)
	}
	return buf, nil
}
