package id3

import "io"

type codec24 struct {
}

func (c *codec24) Read(t *Tag, r io.Reader) (int64, error) {
	return 0, nil
}

// A FrameAPIC contains an image.
type FrameAPIC struct {
	Size        uint8
	Flags       uint8
	Encoding    Encoding
	MimeType    string
	Type        PictureType
	Description string
	Data        []byte
}
