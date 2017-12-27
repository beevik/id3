package id3

import (
	"bytes"
	"io"
	"testing"
)

type testCase struct {
	bytes []byte
	tag   Tag
	err   string
}

var headers = []testCase{
	{[]byte{0x48, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, Tag{}, "invalid id3 tag"},
	{[]byte{}, Tag{}, "invalid id3 tag"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x39}, Tag{}, "invalid id3 tag"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}, Tag{}, "invalid sync code"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00}, Tag{}, "invalid sync code"},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x39, 0x5d}, Tag{3, 0, 0, 7389, []frame{}}, ""},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x33, 0x39, 0x5d}, Tag{3, 0, 0, 0x0fecdcdd, []frame{}}, ""},
	{[]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x7f, 0x7f, 0x7f, 0x7f}, Tag{3, 0, 0, 0x0fffffff, []frame{}}, ""},
}

func TestHeader(t *testing.T) {
	for i, c := range headers {
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

func TestUnsync(t *testing.T) {
	for bs := 1; bs < 12; bs++ {
		b := bytes.NewReader([]byte{0xff, 0x00, 0xff, 0x00})
		r := newUnsyncReader(b)

		buf := make([]byte, bs)
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

		if len(out) != 2 || out[0] != 0xff || out[1] != 0xff {
			t.Errorf("invalid unsync result: %v\n", out)
		}
	}
}
