package id3

import (
	"bytes"
	"io"
)

// A Tag represents an entire ID3 tag, including zero or more frames.
type Tag struct {
	Version uint8   // 2, 3 or 4 (for 2.2, 2.3 or 2.4)
	Flags   uint8   // See TagFlag* list
	Size    uint32  // Size not including the header
	Frames  []Frame // All ID3 frames included in the tag
}

// Possible flags associated with an ID3 tag.
const (
	TagFlagUnsync       uint8 = 1 << 7
	TagFlagExtended           = 1 << 6
	TagFlagExperimental       = 1 << 5
	TagFlagFooter             = 1 << 4
)

func getCodec(v uint8) codec {
	switch v {
	case 2:
		return new(codec22)
	case 3:
		return new(codec23)
	case 4:
		return new(codec24)
	default:
		panic("invalid codec version")
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

	// Process the tag size.
	t.Size, err = decodeSyncSafeUint32(hdr[6:10])
	if err != nil {
		return nn, err
	}

	// Choose a version-appropriate codec to process the data.
	codec := getCodec(t.Version)

	// Decode the tag's frames.
	for remain := t.Size; remain > 0; {
		f := Frame{}

		n, err = codec.decodeFrame(&f, r)
		nn += int64(n)
		if err == errPaddingEncountered {
			pad := make([]byte, remain-4)
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
		remain -= f.Header.Size + 10
	}

	return nn, nil
}

// WriteTo writes an ID3 tag to an output stream. It returns the number of
// bytes written and any error encountered during encoding.
func (t *Tag) WriteTo(w io.Writer) (int64, error) {
	codec := getCodec(t.Version)

	// Create a temporary buffer to hold encoded frames.
	framebuf := bytes.NewBuffer([]byte{})
	var wf io.Writer = framebuf
	if (t.Flags & TagFlagUnsync) != 0 {
		wf = newUnsyncWriter(wf)
	}

	// Encode the tag's frames into the temporary buffer.
	var size uint32
	for _, f := range t.Frames {
		n, err := codec.encodeFrame(&f, wf)
		size += uint32(n)
		if err != nil {
			return 0, err
		}
	}

	// Create a buffer holding the 10-byte header.
	hdr := []byte{'I', 'D', '3', t.Version, 0, t.Flags, 0, 0, 0, 0}
	err := encodeSyncSafeUint32(hdr[6:10], size)
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
