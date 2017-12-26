package id3

import (
	"bytes"
	"testing"
)

type testCase struct {
	bytes []byte
	tag   tag
	err   string
}

var headers = []testCase{
	{[]byte{0x48, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, tag{}, "invalid id3 tag"},
	{[]byte{}, tag{}, "invalid id3 tag"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x39}, tag{}, "invalid id3 tag"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}, tag{}, "invalid sync code"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00}, tag{}, "invalid sync code"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x39, 0x5d}, tag{3, 0, 0, 7389, []frame{}}, ""},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x33, 0x39, 0x5d}, tag{3, 0, 0, 0x0fecdcdd, []frame{}}, ""},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, tag{3, 0, 0, 0x0fffffff, []frame{}}, ""},
}

func TestHeader(t *testing.T) {
	for i, c := range headers {
		tag := new(tag)
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

		if tag.version != c.tag.version {
			t.Errorf("header case %v:\n  invalid header version: got %x, expected: %x\n",
				i, tag.version, c.tag.version)
		}
		if tag.headerFlags != c.tag.headerFlags {
			t.Errorf("header case %v:\n  invalid header flags: got %v expected %v\n",
				i, tag.headerFlags, c.tag.headerFlags)
		}
		if tag.size != c.tag.size {
			t.Errorf("header case %v:\n  invalid header size: got %v, expected %v\n",
				i, tag.size, c.tag.size)
		}
	}
}
