package id3

type codec23 struct {
}

func newCodec23() *codec23 {
	return &codec23{}
}

func (c *codec23) Decode(t *Tag, r *reader) error {
	return errUnimplemented
}

func (c *codec23) Encode(t *Tag, w *writer) error {
	return errUnimplemented
}
