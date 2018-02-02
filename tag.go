package id3

import (
	"bytes"
	"hash/crc32"
	"io"
)

// A Tag represents an entire ID3 tag, including zero or more frames.
type Tag struct {
	Version      uint8    // 2, 3 or 4 (for 2.2, 2.3 or 2.4)
	Flags        TagFlags // Flags
	Size         int      // Size not including the header
	Padding      int      // Number of bytes of padding
	CRC          uint32   // Optional CRC code
	Restrictions uint16   // ID3 restrictions (v2.4 only)
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

func newCodec(v uint8) (codec, error) {
	switch v {
	case 2:
		return newCodec22(), nil
	case 3:
		return newCodec23(), nil
	case 4:
		return newCodec24(), nil
	default:
		return nil, ErrInvalidVersion
	}
}

// ReadFrom reads from a stream into an ID3 tag. It returns the number of
// bytes read and any error encountered during decoding.
func (t *Tag) ReadFrom(r io.Reader) (int64, error) {
	var nn int64

	// Attempt to read the 10-byte ID3 header.
	hdr := make([]byte, 10)
	n, err := io.ReadFull(r, hdr)
	nn += int64(n)
	if err != nil {
		return nn, ErrInvalidTag
	}

	// Make sure the tag id is "ID3".
	if hdr[0] != 'I' || hdr[1] != 'D' || hdr[2] != '3' {
		return nn, ErrInvalidTag
	}

	// Process the version number (2.2, 2.3, or 2.4).
	t.Version = hdr[3]

	// Choose a version-appropriate codec to process the data.
	codec, err := newCodec(t.Version)
	if err != nil {
		return nn, err
	}

	// Allow the codec to interpret the flags field.
	flags := uint32(hdr[5])
	t.Flags = TagFlags(codec.HeaderFlags().Decode(flags))

	// Process the tag size.
	size, err := decodeSyncSafeUint32(hdr[6:10])
	t.Size = int(size)
	if err != nil {
		return nn, err
	}

	// Read the rest of the tag into a buffer.
	buf := make([]byte, t.Size)
	n, err = io.ReadFull(r, buf)
	nn += int64(n)
	if err != nil {
		return nn, ErrInvalidTag
	}

	// If the "unsync" flag is set, remove all unsync codes from the buffer.
	if (t.Flags & TagFlagUnsync) != 0 {
		buf = removeUnsyncCodes(buf)
	}

	// Create a new reader for the remaining tag data.
	rb := bytes.NewBuffer(buf)

	// Decode the extended header if it exists.
	if (t.Flags & TagFlagExtended) != 0 {
		n, err = codec.DecodeExtendedHeader(t, rb)
		nn += int64(n)
		if err != nil {
			return nn, err
		}
	}

	// Check the CRC.
	if (t.Flags & TagFlagHasCRC) != 0 {
		crc := crc32.ChecksumIEEE(rb.Bytes())
		if crc != t.CRC {
			return nn, ErrInvalidCRC
		}
	}

	// Decode the tag's frames until the tag is exhausted.
	for rb.Len() > 0 {
		f := Frame{}

		n, err = codec.DecodeFrame(t, &f, rb)
		nn += int64(n)

		if err == errPaddingEncountered {
			t.Padding = rb.Len()
			pad := make([]byte, t.Padding)
			n, err = io.ReadFull(rb, pad)
			nn += int64(n)
			if err != nil {
				return nn, err
			}
			break
		}

		if err != nil {
			return nn, err
		}

		t.Frames = append(t.Frames, f)
	}

	// If there's a footer, validate it.
	if (t.Flags & TagFlagFooter) != 0 {
		footer := make([]byte, 10)
		n, err = io.ReadFull(r, footer)
		nn += int64(n)
		if err != nil {
			return nn, err
		}

		if footer[0] != '3' || footer[1] != 'D' || footer[2] != 'I' {
			return nn, ErrInvalidFooter
		}
		if bytes.Compare(footer[3:], hdr[3:]) != 0 {
			return nn, ErrInvalidFooter
		}
	}

	return nn, nil
}

// WriteTo writes an ID3 tag to an output stream. It returns the number of
// bytes written and any error encountered during encoding.
func (t *Tag) WriteTo(w io.Writer) (int64, error) {
	codec, err := newCodec(t.Version)
	if err != nil {
		return 0, err
	}

	// Create a temporary buffer to hold everything but the 10-byte header.
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	// Encode the tag's frames into the temporary buffer.
	for _, f := range t.Frames {
		_, err := codec.EncodeFrame(t, &f, buf)
		if err != nil {
			return 0, err
		}
	}

	// If the tag's unsync flag is set, add unsync codes to the frame buffer.
	b := buf.Bytes()
	if (t.Flags & TagFlagUnsync) != 0 {
		b = addUnsyncCodes(b)
	}

	// Create a buffer holding the 10-byte header.
	flags := uint8(codec.HeaderFlags().Encode(uint32(t.Flags)))
	hdr := []byte{'I', 'D', '3', t.Version, 0, flags, 0, 0, 0, 0}
	err = encodeSyncSafeUint32(hdr[6:10], uint32(len(b)))
	if err != nil {
		return 0, err
	}

	var nn int64

	// Write the header to the output stream.
	n, err := w.Write(hdr)
	nn += int64(n)
	if err != nil {
		return nn, err
	}

	// Write the frames to the output stream.
	n, err = w.Write(b)
	nn += int64(n)
	return nn, err
}
