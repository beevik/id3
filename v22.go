package id3

type codec22 struct {
}

func newCodec22() *codec22 {
	return &codec22{}
}

func (c *codec22) Decode(t *Tag, r *reader) error {
	return errUnimplemented
}

func (c *codec22) Encode(t *Tag, w *writer) error {
	return errUnimplemented
}
