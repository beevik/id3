package id3

import (
	"bufio"
	"errors"
	"io"
)

var (
	// ErrInvalidTag indicates an invalid ID3 header.
	ErrInvalidTag = errors.New("invalid id3 tag")

	// ErrInvalidVersion indicates an invalid ID3 version number. Must be
	// 2.3 or 2.4.
	ErrInvalidVersion = errors.New("invalid id3 version")

	// ErrInvalidHeaderFlags indicates an invalid header flag was set in the
	// ID3 tag.
	ErrInvalidHeaderFlags = errors.New("invalid header flags")

	// ErrBadSync indicates an invalid synchro code was encountered.
	ErrBadSync = errors.New("invalid sync code")
)

// A Tag represents an entire ID3 tag, including zero or more frames.
type Tag struct {
	Version uint8 // 3 or 4 (for 2.3 or 2.4)
	Flags   uint8
	Size    uint32
	Frames  []Frame
}

// Possible flags associated with an ID3 tag.
const (
	TagFlagUnsync       uint8 = 1 << 7
	TagFlagExtended           = 1 << 6
	TagFlagExperimental       = 1 << 5
	TagFlagFooter             = 1 << 4
)

// ReadFrom reads from a stream into a tag.
func (t *Tag) ReadFrom(r io.Reader) (int64, error) {

	var nn int64

	// Attempt to read the 10-byte ID3 header.
	hdr := make([]byte, 10)
	n, err := r.Read(hdr)
	nn += int64(n)
	if n < 10 || err != nil {
		return nn, ErrInvalidTag
	}

	// Make sure the tag id is ID3.
	if string(hdr[0:3]) != "ID3" {
		return nn, ErrInvalidTag
	}

	// Process the version number (2.2, 2.3, or 2.4).
	t.Version = hdr[3]
	if t.Version < 2 || t.Version > 4 {
		return nn, ErrInvalidVersion
	}
	if hdr[4] != 0 {
		return nn, ErrInvalidVersion
	}

	// Process the header flags.
	t.Flags = hdr[5]

	// If the "unsync" flag is set, then use an unsync reader to remove any
	// sync codes.
	if (t.Flags & TagFlagUnsync) != 0 {
		r = newUnsyncReader(r)
	}

	// Use a buffered reader so we can peek ahead.
	b := bufio.NewReader(r)

	// Process the tag size.
	t.Size, err = readSyncSafeUint32(hdr[6:10])
	if err != nil {
		return nn, err
	}

	// Instantiate a version-appropriate codec to process the data.
	var codec codec
	switch t.Version {
	case 2:
		codec = new(codec22)
	case 3:
		codec = new(codec23)
	case 4:
		codec = new(codec24)
	}

	// Decode the data.
	n, err = codec.Read(t, b)
	nn += int64(n)
	if err != nil {
		return nn, err
	}

	return nn, nil
}
