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

type tag struct {
	version       uint8 // 3 or 4 (for 2.3 or 2.4)
	headerFlags   uint8
	headerFlagsEx uint16
	size          uint32
	frames        []frame
}

const (
	headerFlagUnsync       uint8 = 1 << 7
	headerFlagExtended           = 1 << 6
	headerFlagExperimental       = 1 << 5
	headerFlagFooter             = 1 << 4 // v2.4 only
)

func (t *tag) ReadFrom(r io.Reader) (n int64, err error) {
	// Read the 10-byte ID3 header.
	hdr := make([]byte, 10)
	var hn int
	hn, err = r.Read(hdr)
	n += int64(hn)
	if hn < 10 {
		return n, ErrInvalidTag
	}
	if err != nil {
		return n, err
	}

	// Make sure the tag id is ID3.
	if string(hdr[0:3]) != "ID3" {
		return n, ErrInvalidTag
	}

	// Parse the version number (2.3 or 2.4).
	t.version = hdr[3]
	if t.version != 3 && t.version != 4 {
		return n, ErrInvalidVersion
	}
	if hdr[4] != 0 {
		return n, ErrInvalidVersion
	}

	// Parse the header flags.
	t.headerFlags = hdr[5]
	if t.version == 3 && (t.headerFlags&0xe0) != 0 {
		return n, ErrInvalidHeaderFlags
	}
	if t.version == 4 && (t.headerFlags&0xf0) != 0 {
		return n, ErrInvalidHeaderFlags
	}

	// Parse the tag size.
	t.size, err = readSyncUint32(hdr[6:10])
	if err != nil {
		return n, err
	}

	// Process the rest of the tag using format v2.3 or v2.4.
	var fn int64
	if t.version == 3 {
		fn, err = t.readVersion23(r)
	} else {
		fn, err = t.readVersion24(r)
	}
	n += fn
	if err != nil {
		return n, err
	}

	return n, nil
}

func (t *tag) readVersion23(r io.Reader) (n int64, err error) {
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

func (t *tag) readVersion24(r io.Reader) (n int64, err error) {
	return 0, nil
}

func readSyncUint32(b []byte) (value uint32, err error) {
	if len(b) != 4 {
		return 0, ErrBadSync
	}

	for i := 0; i < 4; i++ {
		if b[i] >= 0x80 {
			return 0, ErrBadSync
		}
		value = (value * 0x80) + uint32(b[i])
	}
	return value, nil
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
