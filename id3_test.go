package id3

import (
	"bytes"
	"testing"
)

func TestHeader(t *testing.T) {
	var cases = []struct {
		bytes []byte
		tag   Tag
		err   string
	}{
		{[]byte{}, Tag{}, "invalid id3 tag"},
		{[]byte{0x48, 0x44, 0x33, 0x04, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, Tag{}, "invalid id3 tag"},
		{[]byte{0x49, 0x44, 0x33, 0x04, 0x00, 0x00, 0x00, 0x00, 0x39}, Tag{}, "invalid id3 tag"},
		{[]byte{0x49, 0x44, 0x33, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}, Tag{}, "invalid sync code"},
		{[]byte{0x49, 0x44, 0x33, 0x04, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00}, Tag{}, "invalid sync code"},
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
		out, err := decodeSyncSafeUint32(c.input)

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
		err := encodeSyncSafeUint32(buf, c.output)

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
		{[]byte{0xff, 0xe0}, []byte{0xff, 0x00, 0xe0}},
		{[]byte{0xff, 0xef}, []byte{0xff, 0x00, 0xef}},
		{[]byte{0xff, 0xf0}, []byte{0xff, 0x00, 0xf0}},
		{[]byte{0xff, 0xff}, []byte{0xff, 0x00, 0xff}},
		{[]byte{0xff, 0x00}, []byte{0xff, 0x00, 0x00}},
		{[]byte{0xff, 0xff, 0xff}, []byte{0xff, 0x00, 0xff, 0xff}},
		{[]byte{0xff, 0xff, 0xff, 0xff}, []byte{0xff, 0x00, 0xff, 0xff, 0x00, 0xff}},
		{[]byte{0x00, 0x01, 0x02, 0x03}, []byte{0x00, 0x01, 0x02, 0x03}},
		{[]byte{0xff, 0xfe, 0xff, 0xfe, 0xff, 0xfe}, []byte{0xff, 0x00, 0xfe, 0xff, 0x00, 0xfe, 0xff, 0x00, 0xfe}},
	}

	for i, c := range cases {
		u := addUnsyncCodes(c.synced)
		if bytes.Compare(u, c.unsynced) != 0 {
			t.Errorf("case %d:\n  unsync failed. got %v, expected %v\n", i, u, c.unsynced)
		}

		s := removeUnsyncCodes(c.unsynced)
		if bytes.Compare(s, c.synced) != 0 {
			t.Errorf("case %d:\n  deUnsync failed. got %v, expected %v\n", i, s, c.synced)
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
		b, err := encodeString(c.input, c.encoding)

		if err != nil {
			t.Errorf("case %v\n  got error '%v'", i, c.err)
		}
		if bytes.Compare(b, c.output) != 0 {
			t.Errorf("case %v\n  got '%v', expected '%v'\n", i, b, c.output)
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
		s, err := decodeString(c.input, c.encoding)

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
			t.Errorf("case %v\n  got '%s' (%v), expected '%s (%v)'\n", i, s, []byte(s), c.output, []byte(c.output))
		}
	}
}

// func TestEncodeStrings(t *testing.T) {
// 	var enc = []Encoding{
// 		EncodingISO88591,
// 		EncodingUTF8,
// 		EncodingUTF16,
// 		EncodingUTF16BOM,
// 	}
// 	var text = [][]string{
// 		{},
// 		{"foo"},
// 		{"foo", "bar", "xyz"},
// 		{"a", "b", "c", "d", "e", "f"},
// 	}

// 	for i, e := range enc {
// 		for _, tt := range text {
// 			ss1 := tt
// 			b, err := encodeStrings(ss1, e)
// 			if err != nil {
// 				t.Error(err)
// 			}

// 			ss2, err := decodeStrings(b, e)
// 			if err != nil {
// 				t.Error(err)
// 			}

// 			equal := true
// 			if len(ss1) != len(ss2) {
// 				equal = false
// 			} else {
// 				for i := 0; i < len(ss1); i++ {
// 					if ss1[i] != ss2[i] {
// 						equal = false
// 						break
// 					}
// 				}
// 			}
// 			if !equal {
// 				t.Errorf("case %d:\n  mismatch. Expected '%v', got '%v'\n", i, ss1, ss2)
// 			}
// 		}
// 	}
// }

func TestFrame(t *testing.T) {
	var inbuf = []byte{
		0x49, 0x44, 0x33, 0x04, 0x00, 0x40, 0x00, 0x00,
		0x01, 0x34, 0x00, 0x00, 0x00, 0x0c, 0x01, 0x20,
		0x05, 0x0e, 0x0d, 0x77, 0x71, 0x41, 0x43, 0x4f,
		0x4d, 0x4d, 0x00, 0x00, 0x00, 0x15, 0x00, 0x4d,
		0x90, 0xf0, 0x00, 0x00, 0x00, 0x0f, 0x03, 0x65,
		0x6e, 0x67, 0x66, 0x6f, 0x6f, 0x00, 0x63, 0x6f,
		0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x59, 0x4c,
		0x54, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x03,
		0x65, 0x6e, 0x67, 0x02, 0x02, 0x6c, 0x79, 0x72,
		0x69, 0x63, 0x73, 0x00, 0x61, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x62, 0x00, 0x00, 0x00, 0x00, 0x02,
		0x63, 0x00, 0x00, 0x00, 0x00, 0x03, 0x50, 0x43,
		0x4e, 0x54, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xaa, 0xbb, 0xcc,
		0x54, 0x49, 0x54, 0x32, 0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x03, 0x59, 0x65, 0x6c, 0x6c, 0x6f,
		0x77, 0x20, 0x53, 0x75, 0x62, 0x6d, 0x61, 0x72,
		0x69, 0x6e, 0x65, 0x50, 0x52, 0x49, 0x56, 0x00,
		0x00, 0x00, 0x0a, 0x00, 0x00, 0x6f, 0x77, 0x6e,
		0x65, 0x72, 0x00, 0x00, 0x01, 0x02, 0x03, 0x41,
		0x53, 0x50, 0x49, 0x00, 0x00, 0x00, 0x15, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
		0xe8, 0x00, 0x05, 0x10, 0x00, 0x83, 0x19, 0x99,
		0x66, 0x66, 0xcc, 0xcc, 0xf3, 0x74,
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

	//hexdump.Dump(outbuf, hexdump.FormatGo, os.Stdout)

	if bytes.Compare(outbuf, inbuf) != 0 {
		t.Errorf("Tag write error: Different bytes encoded")
	}
}
