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
		{[]byte{}, Tag{}, "EOF"},
		{[]byte{0x48, 0x44, 0x33, 0x04, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, Tag{}, "invalid id3 tag"},
		{[]byte{0x49, 0x44, 0x33, 0x04, 0x00, 0x00, 0x00, 0x00, 0x39}, Tag{}, "unexpected EOF"},
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

func TestFrame(t *testing.T) {
	var inbuf = []byte{
		0x49, 0x44, 0x33, 0x04, 0x00, 0x40, 0x00, 0x00,
		0x01, 0x58, 0x00, 0x00, 0x00, 0x0c, 0x01, 0x20,
		0x05, 0x06, 0x3a, 0x2b, 0x09, 0x04, 0x43, 0x4f,
		0x4d, 0x4d, 0x00, 0x00, 0x00, 0x15, 0x00, 0x4d,
		0x90, 0xf0, 0x00, 0x00, 0x00, 0x0f, 0x03, 0x65,
		0x6e, 0x67, 0x66, 0x6f, 0x6f, 0x00, 0x63, 0x6f,
		0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x59, 0x4c,
		0x54, 0x00, 0x00, 0x00, 0x2b, 0x00, 0x00, 0x03,
		0x65, 0x6e, 0x67, 0x02, 0x02, 0x6c, 0x79, 0x72,
		0x69, 0x63, 0x73, 0x00, 0x69, 0x73, 0x20, 0x00,
		0x00, 0x00, 0x03, 0xe8, 0x61, 0x20, 0x73, 0x6f,
		0x6e, 0x67, 0x2e, 0x00, 0x00, 0x00, 0x07, 0xd1,
		0x54, 0x68, 0x69, 0x73, 0x20, 0x00, 0x00, 0x00,
		0x0b, 0xb8, 0x50, 0x43, 0x4e, 0x54, 0x00, 0x00,
		0x00, 0x08, 0x00, 0x00, 0x12, 0x34, 0x56, 0x78,
		0x90, 0xaa, 0xbb, 0xcc, 0x54, 0x49, 0x54, 0x32,
		0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x03, 0x59,
		0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x20, 0x53, 0x75,
		0x62, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x54,
		0x58, 0x58, 0x58, 0x00, 0x00, 0x00, 0x0e, 0x00,
		0x00, 0x03, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x00,
		0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x50,
		0x52, 0x49, 0x56, 0x00, 0x00, 0x00, 0x0a, 0x00,
		0x00, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x00, 0x00,
		0x01, 0x02, 0x03, 0x41, 0x53, 0x50, 0x49, 0x00,
		0x00, 0x00, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x03, 0xe8, 0x00, 0x05, 0x10,
		0x00, 0x83, 0x19, 0x99, 0x66, 0xa7, 0xcc, 0xcc,
		0xf3, 0x74,
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

	//hexdump(outbuf, os.Stdout)

	if bytes.Compare(outbuf, inbuf) != 0 {
		t.Errorf("Tag write error: Different bytes encoded")
	}
}

func serialize(t *testing.T, f Frame) {
	tag1 := Tag{Version: Version2_4}
	tag1.Flags |= TagFlagHasCRC
	tag1.Padding = 512

	tag1.Frames = append(tag1.Frames, f)

	buf := bytes.NewBuffer([]byte{})
	n1, err := tag1.WriteTo(buf)
	if err != nil {
		t.Error(err)
	}

	if tag1.CRC == 0 {
		t.Error("CRC not computed")
	}

	a := bytes.NewBuffer(buf.Bytes())

	tag2 := Tag{}
	n2, err := tag2.ReadFrom(buf)
	if err != nil {
		t.Error(err)
	}

	if n1 != n2 {
		t.Error("Bytes read != bytes written")
	}

	b := bytes.NewBuffer([]byte{})
	_, err = tag2.WriteTo(b)
	if err != nil {
		t.Error(err)
	}

	if tag1.CRC != tag2.CRC {
		t.Error("CRC mismatch")
	}
	if tag1.Size != tag2.Size {
		t.Error("Size mismatch")
	}
	if tag1.Flags != tag2.Flags {
		t.Error("Flags mismatch")
	}
	if bytes.Compare(a.Bytes(), b.Bytes()) != 0 {
		t.Errorf("Bytes mismatched: %s\n", HeaderOf(f).FrameID)
	}
}

func TestTextFrames(t *testing.T) {
	for typ := FrameTypeTextGroupDescription; typ < FrameTypeTextCustom; typ++ {
		f := NewFrameText(typ, "Text frame contents")
		serialize(t, f)
	}
}

func TestTXXX(t *testing.T) {
	f := NewFrameTextCustom("description", "Text frame contents")
	serialize(t, f)
}

func TestURLFrames(t *testing.T) {
	for typ := FrameTypeURLArtist; typ < FrameTypeURLCustom; typ++ {
		f := NewFrameURL(typ, "http://www.example.com/request?id=10")
		serialize(t, f)
	}
}

func TestWXXX(t *testing.T) {
	f := NewFrameURLCustom("description", "http://www.example.com/request?id=10")
	serialize(t, f)
}

func TestAPIC(t *testing.T) {
	data := make([]byte, 1024)
	f := NewFrameAttachedPicture("image/jpeg", "description", PictureTypeCoverFront, data)
	serialize(t, f)
}

func TestAENC(t *testing.T) {
	data := make([]byte, 1024)
	f := NewFrameAudioEncryption("owner", 1000, 32768, data)
	serialize(t, f)
}

func TestASPI(t *testing.T) {
	f := NewFrameAudioSeekPointIndex(30, 3100)
	for i := 100; i >= 0; i-- {
		f.AddIndexOffset(uint32(i * 30))
	}
	f.AddIndexOffset(1505)
	f.AddIndexOffset(3100)

	for i := 1; i < int(f.IndexPoints); i++ {
		if f.IndexOffsets[i-1] > f.IndexOffsets[i] {
			t.Error("ASPI indexes out of order")
		}
	}
	if f.IndexPoints != 103 {
		t.Error("ASPI frame index points incorrect")
	}

	serialize(t, f)
}

func TestCOMM(t *testing.T) {
	f := NewFrameComment("eng", "description", "This is the comment")
	serialize(t, f)
}

func TestENCR(t *testing.T) {
	data := make([]byte, 128)
	f := NewFrameEncryptionMethodRegistration("owner", 0x90, data)
	serialize(t, f)
}

func TestGRID(t *testing.T) {
	data := make([]byte, 1024)
	f := NewFrameGroupID("owner", 0x85, data)
	serialize(t, f)
}

func TestUSLT(t *testing.T) {
	f := NewFrameLyricsUnsync("eng", "descriptor", "These are the\nlyrics!")
	serialize(t, f)
}

func TestSYLT(t *testing.T) {
	f := NewFrameLyricsSync("eng", "descriptor", TimeStampMilliseconds, LyricContentTypeLyrics)
	f.AddSync(10000, "line 3")
	f.AddSync(5000, "line 1")
	f.AddSync(7500, "line 2")
	f.AddSync(12000, "line 4")

	for i := 0; i < 3; i++ {
		if f.Sync[i].TimeStamp > f.Sync[i+1].TimeStamp {
			t.Error("SYLT syncs out of order")
		}
	}

	serialize(t, f)
}

var counts = []uint64{
	0x0000000000000000, 0x0000000000001000, 0x0000000010000000,
	0x0000001234567890, 0x00001234567890ab, 0x001234567890abcd,
	0x1234567890abcdef,
}

func TestPCNT(t *testing.T) {
	for _, c := range counts {
		f := NewFramePlayCount(c)
		serialize(t, f)
	}
}

func TestPOPM(t *testing.T) {
	for _, c := range counts {
		f := NewFramePopularimeter("johndoe@gmail.com", 80, c)
		serialize(t, f)
	}
}

func TestPRIV(t *testing.T) {
	data := make([]byte, 1024)
	f := NewFramePrivate("owner", data)
	serialize(t, f)
}

func TestSYTC(t *testing.T) {
	f := NewFrameSyncTempoCodes(TimeStampFrames)
	f.AddSync(120, 2000)
	f.AddSync(510, 500)
	f.AddSync(0, 1000)
	f.AddSync(257, 3000)

	for i := 0; i < 3; i++ {
		if f.Sync[i].TimeStamp > f.Sync[i+1].TimeStamp {
			t.Error("SYTC syncs out of order")
		}
	}

	serialize(t, f)
}

func TestUSER(t *testing.T) {
	f := NewFrameTermsOfUse("eng", "Terms of Use")
	serialize(t, f)
}

func TestUFID(t *testing.T) {
	f := NewFrameUniqueFileID("owner", "b28f6045-9958-44b5-9da8-34703f5ffa13")
	serialize(t, f)
}
