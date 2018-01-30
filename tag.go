package id3

import (
	"bytes"
	"io"
)

// A Tag represents an entire ID3 tag, including zero or more frames.
type Tag struct {
	Version uint8   // 2, 3 or 4 (for 2.2, 2.3 or 2.4)
	Flags   uint32  // See TagFlag* list
	Size    int     // Size not including the header
	CRC     uint32  // Optional CRC code
	Frames  []Frame // All ID3 frames included in the tag
}

// All tag header flags, including extended header flags
const (
	TagFlagUnsync uint32 = 1 << iota
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

	// Choose a version-appropriate codec to process the data.
	codec, err := newCodec(t.Version)
	if err != nil {
		return nn, err
	}

	// Allow the codec to interpret the flags field.
	t.Flags = codec.HeaderFlags().Decode(hdr[5])

	// If the "unsync" flag is set, then use an unsync reader to remove any
	// sync codes.
	if (t.Flags & TagFlagUnsync) != 0 {
		r = newUnsyncReader(r)
	}

	// Process the tag size.
	size, err := decodeSyncSafeUint32(hdr[6:10])
	t.Size = int(size)
	if err != nil {
		return nn, err
	}
	remain := t.Size

	// Decode the extended header if it exists.
	if (t.Flags & TagFlagExtended) != 0 {
		n, err = codec.DecodeExtendedHeader(t, r)
		nn += int64(n)
		if err != nil {
			return nn, err
		}
		remain -= n
	}

	// Decode the tag's frames until there's no more tag data.
	for remain > 0 {
		f := Frame{}

		n, err = codec.DecodeFrame(&f, r)
		nn += int64(n)
		remain -= n

		// If we hit padding, we're done.
		if err == errPaddingEncountered {
			pad := make([]byte, remain)
			n, err = r.Read(pad)
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

	return nn, nil
}

// WriteTo writes an ID3 tag to an output stream. It returns the number of
// bytes written and any error encountered during encoding.
func (t *Tag) WriteTo(w io.Writer) (int64, error) {
	codec, err := newCodec(t.Version)
	if err != nil {
		return 0, err
	}

	// Create a temporary buffer to hold encoded frames.
	framebuf := bytes.NewBuffer([]byte{})
	var wf io.Writer = framebuf
	if (t.Flags & TagFlagUnsync) != 0 {
		wf = newUnsyncWriter(wf)
	}

	// Encode the tag's frames into the temporary buffer.
	var size uint32
	for _, f := range t.Frames {
		n, err := codec.EncodeFrame(&f, wf)
		size += uint32(n)
		if err != nil {
			return 0, err
		}
	}

	// Create a buffer holding the 10-byte header.
	flags := codec.HeaderFlags().Encode(t.Flags)
	hdr := []byte{'I', 'D', '3', t.Version, 0, flags, 0, 0, 0, 0}
	err = encodeSyncSafeUint32(hdr[6:10], size)
	if err != nil {
		return 0, err
	}

	var nn int64

	// Write the header to the output.
	n, err := w.Write(hdr)
	nn += int64(n)
	if err != nil {
		return nn, err
	}

	// Write the frames to the output.
	n, err = w.Write(framebuf.Bytes())
	nn += int64(n)
	return nn, err
}
