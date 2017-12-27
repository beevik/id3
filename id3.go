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
	version       uint8 // 3 or 4 (for 2.3 or 2.4)
	headerFlags   uint8
	headerFlagsEx uint16
	size          uint32
	frames        []frame
}

const (
	headerFlagUnsync       uint8 = 0x80
	headerFlagExtended           = 0x40
	headerFlagExperimental       = 0x20
	headerFlagFooter             = 0x10 // v2.4 only
)

// ReadFrom reads from a stream into a tag.
func (t *Tag) ReadFrom(r io.Reader) (n int64, err error) {
	b := bufio.NewReader(r)

	var hdr []byte
	hdr, err = b.Peek(10)
	if err != nil {
		return 0, ErrInvalidTag
	}

	// Make sure the tag id is ID3.
	if string(hdr[0:3]) != "ID3" {
		return 0, ErrInvalidTag
	}

	// Process the version number (2.3 or 2.4).
	t.version = hdr[3]
	if t.version != 3 && t.version != 4 {
		return 0, ErrInvalidVersion
	}
	if hdr[4] != 0 {
		return 0, ErrInvalidVersion
	}

	// Process the header flags.
	t.headerFlags = hdr[5]

	// If the "unsync" flag is set, then use an unsync reader to remove
	// sync codes.
	unsync := (t.headerFlags & headerFlagUnsync) != 0
	if unsync {
		r = newUnsyncReader(r)
	}

	// Process the tag size.
	t.size, err = readSyncUint32(hdr[6:10])
	if err != nil {
		return n, err
	}

	// Instantiate a version-appropriate codec to process the data.
	var codec codec
	if t.version == 3 {
		codec = new(codec23)
	} else {
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

func (t *Tag) readVersion23(r io.Reader) (n int64, err error) {
	if (t.headerFlags & 0x1f) != 0 {
		return n, ErrInvalidHeaderFlags
	}

	if (t.headerFlags & headerFlagExtended) != 0 {
		hdr := make([]byte, 10)
		var hn int
		hn, err = r.Read(hdr)
		n += int64(hn)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (t *Tag) readVersion24(r io.Reader) (n int64, err error) {
	if (t.headerFlags & 0x0f) != 0 {
		return n, ErrInvalidHeaderFlags
	}

	return 0, nil
}

func readSyncUint32(b []byte) (value uint32, err error) {
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

type frameHeader struct {
	id    [4]byte
	size  uint32
	flags uint16
}

// frameHeader.flags
const (
	frameFlagDiscardOnTagAltered  = 1 << 15
	frameFlagDiscardOnFileAltered = 1 << 14
	frameFlagReadOnly             = 1 << 13
	frameFlagCompressed           = 1 << 7
	frameFlagEncrypted            = 1 << 6
	frameFlagGroupInfo            = 1 << 5
)

type pictureType uint8

type frameHeaderAPIC struct {
	encoding    uint8
	mimeType    string
	pictureType pictureType
	description string
	data        []byte
}

const (
	picTypeOther             pictureType = 0
	picTypeIcon                          = 1
	picTypeIconOther                     = 2
	picTypeCoverFront                    = 3
	picTypeCoverBack                     = 4
	picTypeLeaflet                       = 5
	picTypeMedia                         = 6
	picTypeArtistLead                    = 7
	picTypeArtist                        = 8
	picTypeConductor                     = 9
	picTypeBand                          = 10
	picTypeComposer                      = 11
	picTypeLyricist                      = 12
	picTypeRecordingLocation             = 13
	picTypeDuringRecording               = 14
	picTypeDuringPerformance             = 15
	picTypeVideoCapture                  = 16
	picTypeFish                          = 17
	picTypeIlllustration                 = 18
	picTypeBandLogotype                  = 19
	picTypePublisherLogotype             = 20
)

type frame struct {
	header frameHeader
	data   []uint8
}
