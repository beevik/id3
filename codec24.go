package id3

import "io"

type codec24 struct {
}

func (c *codec24) Read(t *Tag, r io.Reader) (int64, error) {
	return 0, nil
}

func (c *codec24) Write(t *Tag, w io.Writer) (int64, error) {
	return 0, nil
}

func createFrame24(id string) Frame {
	if id[0] == 'T' {
		return newFrameText(id)
	}

	switch id {
	case "APIC":
		return new(FrameAPIC)
	default:
		return nil
	}
}

//
// FrameText
//

type FrameText struct {
	FrameHeader
	TextID   string
	Encoding Encoding
	Data     string
}

func newFrameText(id string) *FrameText {
	return &FrameText{TextID: id}
}

func (f *FrameText) ID() string {
	return f.TextID
}

func (f *FrameText) ReadFrom(r io.Reader) (int64, error) {
	n, err := f.FrameHeader.ReadFrom(r)
	if err != nil {
		return n, err
	}
	return 0, nil
}

func (f *FrameText) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

//
// FrameAPIC
//

// A FrameAPIC contains an image.
type FrameAPIC struct {
	FrameHeader
	Encoding    Encoding
	MimeType    string
	Type        PictureType
	Description string
	Data        []byte
}

func (f *FrameAPIC) ID() string {
	return "APIC"
}

func (f *FrameAPIC) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}

func (f *FrameAPIC) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}
