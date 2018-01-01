package id3

import (
	"bytes"
	"io"
	"testing"
)

func TestHeader(t *testing.T) {
	var cases = []struct {
		bytes []byte
		tag   Tag
		err   string
	}{
		{[]byte{}, Tag{}, "invalid id3 tag"},
		{[]byte{0x48, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, Tag{}, "invalid id3 tag"},
		{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x39}, Tag{}, "invalid id3 tag"},
		{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}, Tag{}, "invalid sync code"},
		{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00}, Tag{}, "invalid sync code"},
		{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x39, 0x5d}, Tag{3, 0, 7389, []Frame{}}, ""},
		{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x33, 0x39, 0x5d}, Tag{3, 0, 0x0fecdcdd, []Frame{}}, ""},
		{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, Tag{3, 0, 0x0fffffff, []Frame{}}, ""},
	}
	for i, c := range cases {
		tag := new(Tag)
		r := bytes.NewReader(c.bytes)
		_, err := tag.ReadFrom(r)

		if err == nil && c.err != "" {
			t.Errorf("header case %v:\n  expected error '%v', got success\n", i, c.err)
		}
		if err != nil && err.Error() != c.err {
			t.Errorf("header case %v:\n  got error '%v', expected error '%v'\n", i, err, c.err)
		}
		if err != nil && c.err == "" {
			t.Errorf("header case %v\n  got error '%v', expected success\n", i, err)
		}
		if err != nil && err.Error() == c.err {
			continue
		}

		if tag.Version != c.tag.Version {
			t.Errorf("header case %v:\n  invalid header version: got %x, expected: %x\n",
				i, tag.Version, c.tag.Version)
		}
		if tag.Flags != c.tag.Flags {
			t.Errorf("header case %v:\n  invalid header flags: got %v expected %v\n",
				i, tag.Flags, c.tag.Flags)
		}
		if tag.Size != c.tag.Size {
			t.Errorf("header case %v:\n  invalid header size: got %v, expected %v\n",
				i, tag.Size, c.tag.Size)
		}
	}
}

func TestSyncUint32(t *testing.T) {
	var cases = []struct {
		input  []byte
		output uint32
		err    string
	}{
		{[]byte{0x00}, 0, "invalid sync code"},
		{[]byte{0x00, 0x00, 0x00}, 0, "invalid sync code"},
		{[]byte{0x00, 0x00, 0x00, 0x00}, 0, ""},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00}, 0, ""},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, "invalid sync code"},
		{[]byte{0x00, 0x00, 0x00, 0x80}, 0, "invalid sync code"},
		{[]byte{0xff, 0x00, 0x00, 0x00}, 0, "invalid sync code"},
		{[]byte{0xff, 0x00, 0x00, 0x00, 0x00}, 0, "invalid sync code"},
		{[]byte{0x01, 0x01, 0x01, 0x01}, 0x00204081, ""},
		{[]byte{0x01, 0x02, 0x03, 0x04}, 0x00208184, ""},
		{[]byte{0x00, 0x05, 0x30, 0x00}, 0x00015800, ""},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0x05}, 0x1040c205, ""},
		{[]byte{0x7f, 0x7f, 0x7f, 0x7f}, 0x0fffffff, ""},
		{[]byte{0x00, 0x7f, 0x7f, 0x7f, 0x7f}, 0x0fffffff, ""},
		{[]byte{0x0f, 0x7f, 0x7f, 0x7f, 0x7f}, 0xffffffff, ""},
	}

	for i, c := range cases {
		out, err := readSyncSafeUint32(c.input)

		if err == nil && c.err != "" {
			t.Errorf("case %v:\n  expected error '%v', got success\n", i, c.err)
		} else if err != nil && err.Error() != c.err {
			t.Errorf("case %v:\n  got error '%v', expected error '%v'\n", i, err, c.err)
		} else if err != nil && c.err == "" {
			t.Errorf("case %v\n  got error '%v', expected success\n", i, err)
		} else if err != nil && err.Error() == c.err {
			continue
		}

		if out != c.output {
			t.Errorf("case %v\n  got %d. expected %d\n", i, out, c.output)
		}
	}

	for i, c := range cases {
		if c.err != "" {
			continue
		}
		buf := make([]byte, len(c.input))
		err := writeSyncSafeUint32(buf, c.output)

		if err != nil {
			t.Errorf("case %v\n  got error '%v', expected '%v'\n", i, err, c.input)
		} else if bytes.Compare(buf, c.input) != 0 {
			t.Errorf("case %v\n  got '%v', expected '%v'\n", i, buf, c.input)
		}
	}
}

func TestUnsync(t *testing.T) {
	var cases = []struct {
		synced   []byte
		unsynced []byte
	}{
		{[]byte{0x00}, []byte{0x00}},
		{[]byte{0xff}, []byte{0xff}},
		{[]byte{0xff, 0xc0}, []byte{0xff, 0xc0}},
		{[]byte{0xff, 0xf0}, []byte{0xff, 0x00, 0xf0}},
		{[]byte{0xff, 0xe0}, []byte{0xff, 0x00, 0xe0}},
		{[]byte{0xff, 0x00}, []byte{0xff, 0x00, 0x00}},
		{[]byte{0xff, 0xff}, []byte{0xff, 0x00, 0xff}},
		{[]byte{0xff, 0xff, 0xff}, []byte{0xff, 0x00, 0xff, 0xff}},
		{[]byte{0xff, 0xff, 0xff, 0xff}, []byte{0xff, 0x00, 0xff, 0xff, 0x00, 0xff}},
		{[]byte{0x00, 0x01, 0x02, 0x03}, []byte{0x00, 0x01, 0x02, 0x03}},
		{[]byte{0xff, 0xfe, 0xff, 0xfe, 0xff, 0xfe}, []byte{0xff, 0x00, 0xfe, 0xff, 0x00, 0xfe, 0xff, 0x00, 0xfe}},
	}

	// Test synced -> unsynced
	for i, c := range cases {
		for chunk := 1; chunk <= len(c.synced); chunk++ {
			b := bytes.NewBuffer([]byte{})
			w := newUnsyncWriter(b)
			for j := 0; j < len(c.synced); j += chunk {
				right := j + chunk
				if right > len(c.synced) {
					right = len(c.synced)
				}
				_, err := w.Write(c.synced[j:right])
				if err != nil {
					t.Error(err)
				}
			}
			if bytes.Compare(b.Bytes(), c.unsynced) != 0 {
				t.Errorf("case %d:\n  unsync write failed (chunk=%d). got: %v\n", i, chunk, b.Bytes())
			}
		}
	}

	// Test unsynced -> synced
	for i, c := range cases {
		for chunk := 1; chunk <= len(c.unsynced)*2; chunk++ {
			b := bytes.NewReader(c.unsynced)
			r := newUnsyncReader(b)

			buf := make([]byte, chunk)
			out := make([]byte, 0)
			for {
				n, err := r.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Error(err)
				}
				out = append(out, buf[:n]...)
			}

			if bytes.Compare(out, c.synced) != 0 {
				t.Errorf("case %d:\n  unsync read failed (chunk=%d). got %v\n", i, chunk, out)
			}
		}
	}
}

func TestStringEncode(t *testing.T) {
	var cases = []struct {
		encoding Encoding
		input    string
		output   []byte
		err      string
	}{
		{
			EncodingISO88591,
			"VX¡¢Æ",
			[]byte{0x56, 0x58, 0xa1, 0xa2, 0xc6},
			"",
		},
		{
			EncodingISO88591,
			"©𝌆☃",
			[]byte{0xa9, 0x2e, 0x2e},
			"",
		},
		{
			EncodingUTF8,
			"©𝌆☃",
			[]byte{0xc2, 0xa9, 0xf0, 0x9d, 0x8c, 0x86, 0xe2, 0x98, 0x83},
			"",
		},
		{
			EncodingUTF16,
			"©𝌆☃",
			[]byte{0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03},
			"",
		},
		{
			EncodingUTF16BOM,
			"©𝌆☃",
			[]byte{0xfe, 0xff, 0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03},
			"",
		},
	}

	for i, c := range cases {
		b := bytes.NewBuffer([]byte{})
		_, err := writeEncodedString(b, c.input, c.encoding)

		if err != nil {
			t.Errorf("case %v\n  got error '%v'", i, c.err)
		}
		if bytes.Compare(b.Bytes(), c.output) != 0 {
			t.Errorf("case %v\n  got '%v', expected '%v'\n", i, b.Bytes(), c.output)
		}
	}
}

func TestStringDecode(t *testing.T) {
	var cases = []struct {
		encoding Encoding
		input    []byte
		output   string
		err      string
	}{
		{
			EncodingISO88591,
			[]byte{0x56, 0x58, 0xa1, 0xa2, 0xc6},
			"VX¡¢Æ",
			"",
		},
		{
			EncodingISO88591,
			[]byte{0x56, 0x58, 0xa1, 0xa2, 0xc6, 0x00},
			"VX¡¢Æ",
			"",
		},
		{
			EncodingISO88591,
			[]byte{0x56, 0x58, 0xa1, 0xa2, 0xc6, 0x00, 0xff},
			"VX¡¢Æ",
			"",
		},
		{
			EncodingUTF8,
			[]byte{0xc2, 0xa9, 0xf0, 0x9d, 0x8c, 0x86, 0xe2, 0x98, 0x83},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF8,
			[]byte{0xc2, 0xa9, 0xf0, 0x9d, 0x8c, 0x86, 0xe2, 0x98, 0x83, 0x00},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF8,
			[]byte{0xc2, 0xa9, 0xf0, 0x9d, 0x8c, 0x86, 0xe2, 0x98, 0x83, 0x00, 0x80},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF16,
			[]byte{0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF16,
			[]byte{0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26},
			"",
			"invalid text string encountered",
		},
		{
			EncodingUTF16BOM,
			[]byte{0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF16BOM,
			[]byte{0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03, 0x00, 0x00},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF16,
			[]byte{0xfe, 0xff, 0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF16BOM,
			[]byte{0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03},
			"©𝌆☃",
			"",
		},
		{
			EncodingUTF16BOM,
			[]byte{0xfe, 0xff, 0x00, 0xa9, 0xd8, 0x34, 0xdf, 0x06, 0x26, 0x03, 0x00},
			"",
			"invalid text string encountered",
		},
	}

	for i, c := range cases {
		r := bytes.NewReader(c.input)
		_, s, err := readEncodedString(r, len(c.input), c.encoding)

		if err == nil && c.err != "" {
			t.Errorf("case %v:\n  expected error '%v', got success\n", i, c.err)
		} else if err != nil && err.Error() != c.err {
			t.Errorf("case %v:\n  got error '%v', expected error '%v'\n", i, err, c.err)
		} else if err != nil && c.err == "" {
			t.Errorf("case %v\n  got error '%v', expected success\n", i, err)
		} else if err != nil && err.Error() == c.err {
			continue
		}

		if s != c.output {
			t.Errorf("case %v\n  got '%s', expected '%s'\n", i, s, c.output)
		}
	}
}

func TestFrame(t *testing.T) {
	inbuf := []byte{0x49, 0x44, 0x33, 0x04, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x3b, 0x54, 0x49, 0x54, 0x32, 0x00, 0x00,
		0x00, 0x11, 0x00, 0x00, 0x00, 0x54, 0x68, 0x65,
		0x20, 0x44, 0x69, 0x73, 0x61, 0x70, 0x70, 0x6f,
		0x69, 0x6e, 0x74, 0x65, 0x64, 0x54, 0x50, 0x45,
		0x31, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
		0x58, 0x54, 0x43, 0x54, 0x41, 0x4c, 0x42, 0x00,
		0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x4e, 0x6f,
		0x6e, 0x73, 0x75, 0x63, 0x68,
	}

	tag := &Tag{}
	n, err := tag.ReadFrom(bytes.NewBuffer(inbuf))
	if err != nil {
		t.Errorf("Tag read error: %v\n", err)
	}
	if n != int64(len(inbuf)) {
		t.Errorf("Tag read error: Not all bytes processed")
	}

	b := bytes.NewBuffer([]byte{})
	n, err = tag.WriteTo(b)
	if err != nil {
		t.Errorf("Tag write error: %v\n", err)
	}
	outbuf := b.Bytes()
	if bytes.Compare(outbuf, inbuf) != 0 {
		t.Errorf("Tag write error: Different bytes encoded")
	}
}
