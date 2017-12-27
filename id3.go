package id3

import (
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
func (t *Tag) ReadFrom(r io.Reader) (n int64, err error) {
	// Attempt to read the 10-byte ID3 header.
	hdr := make([]byte, 10)
	var hn int
	hn, err = r.Read(hdr)
	n += int64(hn)
	if hn < 10 || err != nil {
		return 0, ErrInvalidTag
	}

	// Make sure the tag id is ID3.
	if string(hdr[0:3]) != "ID3" {
		return 0, ErrInvalidTag
	}

	// Process the version number (2.2, 2.3, or 2.4).
	t.Version = hdr[3]
	if t.Version < 2 || t.Version > 4 {
		return 0, ErrInvalidVersion
	}
	if hdr[4] != 0 {
		return 0, ErrInvalidVersion
	}

	// Process the header flags.
	t.Flags = hdr[5]

	// If the "unsync" flag is set, then use an unsync reader to remove any
	// sync codes.
	if (t.Flags & TagFlagUnsync) != 0 {
		r = newUnsyncReader(r)
	}

	// Process the tag size.
	t.Size, err = readSyncSafeUint32(hdr[6:10])
	if err != nil {
		return n, err
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
	var cn int64
	cn, err = codec.Read(t, r)
	n += cn
	if err != nil {
		return n, err
	}

	return n, nil
}

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
