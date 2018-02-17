package id3

import (
	"io"
)

// A Tag represents an entire ID3 tag, including zero or more frames.
type Tag struct {
	Version      Version  // ID3 codec version (2.2, 2.3, or 2.4)
	Flags        TagFlags // Flags
	Size         int      // Size not including the header
	Padding      int      // Number of bytes of padding
	CRC          uint32   // Optional CRC code
	Restrictions uint8    // ID3 restrictions (v2.4 only)
	Frames       []Frame  // All ID3 frames included in the tag
}

// TagFlags describe flags that may appear within an ID3 tag. Not all
// flags are supported by all versions of the ID3 codec.
type TagFlags uint32

// All possible TagFlags.
const (
	TagFlagUnsync TagFlags = 1 << iota
	TagFlagExtended
	TagFlagExperimental
	TagFlagFooter
	TagFlagIsUpdate
	TagFlagHasCRC
	TagFlagHasRestrictions
)

func newCodec(v Version) (codec, error) {
	switch v {
	case Version2_2:
		return newCodec22(), nil
	case Version2_3:
		return newCodec23(), nil
	case Version2_4:
		return newCodec24(), nil
	default:
		return nil, ErrInvalidVersion
	}
}

// NewTag creates a new ID3 tag object. Use this constructor when you
// wish to manually construct a new Tag.
func NewTag(v Version, f TagFlags) *Tag {
	return &Tag{
		Version: v,
		Flags:   f,
	}
}

// ReadFrom reads from a stream into an ID3 tag. It returns the number of
// bytes read and any error encountered during decoding.
func (t *Tag) ReadFrom(r io.Reader) (int64, error) {
	rr := newReader(r)

	// Read 3 bytes to check for the ID3 file id.
	if rr.Load(3); rr.err != nil {
		return int64(rr.n), rr.err
	}
	fileID := rr.Bytes()
	if fileID[0] != 'I' || fileID[1] != 'D' || fileID[2] != '3' {
		return int64(rr.n), ErrInvalidTag
	}

	// Read another byte to check the version number (2.2, 2.3, or 2.4).
	if rr.Load(1); rr.err != nil {
		return int64(rr.n), rr.err
	}
	t.Version = Version(rr.Bytes()[3])
	c, err := newCodec(t.Version)
	if err != nil {
		return int64(rr.n), err
	}

	// Decode the rest of the tag.
	err = c.Decode(t, rr)
	return int64(rr.n), err
}

// WriteTo writes an ID3 tag to an output stream. It returns the number of
// bytes written and any error encountered during encoding.
func (t *Tag) WriteTo(w io.Writer) (int64, error) {
	ww := newWriter(w)

	// Select a codec based on the ID3 version.
	c, err := newCodec(t.Version)
	if err != nil {
		return 0, err
	}

	err = c.Encode(t, ww)
	return int64(ww.n), err
}
