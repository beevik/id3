package id3

import (
	"fmt"
	"reflect"
)

// A reflector uses reflection to scan or output the contents of frame
// structures.
type reflector struct {
	version Version
	vdata   *versionData
}

func newReflector(v Version, vdata *versionData) *reflector {
	return &reflector{
		version: v,
		vdata:   vdata,
	}
}

// A property holds the reflection data necessary to update a property's
// value. Usually the property is a struct field.
type property struct {
	typ   reflect.Type
	value reflect.Value
	name  string
}

// The state structure keeps track of persistent state required while
// decoding a single frame.
type state struct {
	frameID     string     // current frame ID
	frameType   FrameType  // current frame type
	structStack valueStack // stack of active struct values
	fieldCount  int        // current frame's field count
	fieldIndex  int        // current frame field index
}

// ScanFrame uses reflection to scan the contents of an ID3 frame from a
// reader buffer.
func (rf *reflector) ScanFrame(r *reader, frameID string) (Frame, error) {
	state := state{frameID: frameID}

	typ := rf.vdata.frameTypes.LookupReflectType(frameID)

	p := property{
		typ:   typ,
		value: reflect.New(typ),
		name:  "",
	}

	rf.scanStruct(r, p, &state)
	if r.err != nil {
		return nil, r.err
	}

	f := p.value.Interface().(Frame)
	return f, nil
}

// SetFrameHeader uses reflection to update a frame's header.
func (rf *reflector) SetFrameHeader(f Frame, h *FrameHeader) {
	source := reflect.ValueOf(h).Elem()
	target := reflect.ValueOf(f).Elem().Field(0)
	target.Set(source)
}

// OutputFrame uses reflection to output the contents of an ID3 frame to
// a writer buffer.
func (rf *reflector) OutputFrame(w *writer, f Frame) (frameID string, err error) {
	state := state{}

	p := property{
		typ:   reflect.TypeOf(f).Elem(),
		value: reflect.ValueOf(f).Elem(),
		name:  "",
	}

	rf.outputStruct(w, p, &state)
	if w.err != nil {
		return "", w.err
	}

	frameID = rf.vdata.frameTypes.LookupFrameID(state.frameType)
	return frameID, nil
}

func (rf *reflector) scanStruct(r *reader, p property, state *state) {
	if p.typ.Name() == "FrameHeader" {
		return
	}

	state.structStack.push(p.value.Elem())
	if state.structStack.depth() == 1 {
		state.fieldCount = p.typ.NumField()
	}

	for ii, n := 0, p.typ.NumField(); ii < n; ii++ {
		if state.structStack.depth() == 1 {
			state.fieldIndex = ii
		}

		field := p.typ.Field(ii)

		fp := property{
			typ:   field.Type,
			value: p.value.Elem().Field(ii),
			name:  field.Name,
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			rf.scanUint8(r, fp, state)

		case reflect.Uint16:
			rf.scanUint16(r, fp, state)

		case reflect.Uint32:
			rf.scanUint32(r, fp, state)

		case reflect.Uint64:
			rf.scanUint64(r, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				rf.scanByteSlice(r, fp, state)
			case reflect.Uint32:
				rf.scanUint32Slice(r, fp, state)
			case reflect.String:
				rf.scanStringSlice(r, fp, state)
			case reflect.Struct:
				rf.scanStructSlice(r, fp, state)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			rf.scanString(r, fp, state)

		case reflect.Struct:
			rf.scanStruct(r, fp, state)

		default:
			panic(errUnknownFieldType)
		}
	}

	state.structStack.pop()
}

func (rf *reflector) scanUint8(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	if p.typ.Name() == "FrameType" {
		state.frameType = rf.vdata.frameTypes.LookupFrameType(state.frameID)
		p.value.SetUint(uint64(state.frameType))
		return
	}

	bounds, hasBounds := rf.vdata.bounds[p.name]

	value := r.ConsumeByte()
	if r.err != nil {
		return
	}

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		r.err = bounds.err
		return
	}

	p.value.SetUint(uint64(value))
}

func (rf *reflector) scanUint16(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	var value uint16
	switch p.name {
	case "BPM":
		value = uint16(r.ConsumeByte())
		if value == 0xff {
			value += uint16(r.ConsumeByte())
		}
	default:
		b := r.ConsumeBytes(2)
		value = uint16(b[0])<<8 | uint16(b[1])
	}

	if r.err != nil {
		return
	}

	p.value.SetUint(uint64(value))
}

func (rf *reflector) scanUint32(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	b := r.ConsumeBytes(4)

	var value uint64
	for _, bb := range b {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (rf *reflector) scanUint64(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	var b []byte
	switch p.name {
	case "Counter":
		b = r.ConsumeAll()
	default:
		panic(errUnknownFieldType)
	}

	var value uint64
	for _, bb := range b {
		value = (value << 8) | uint64(bb)
	}

	p.value.SetUint(value)
}

func (rf *reflector) scanByteSlice(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	b := r.ConsumeAll()
	p.value.Set(reflect.ValueOf(b))
}

func (rf *reflector) scanUint32Slice(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	if p.name != "IndexOffsets" {
		panic(errUnknownFieldType)
	}

	sf := state.structStack.first()
	length := uint32(sf.FieldByName("IndexedDataLength").Uint())
	bits := uint32(sf.FieldByName("BitsPerIndex").Uint())

	var offsets []uint32

	ff := r.ConsumeAll()
	switch bits {
	case 8:
		offsets = make([]uint32, 0, len(ff))
		for _, f := range ff {
			frac := uint32(f)
			offset := (frac*length + (1 << 7)) >> 8
			if offset > length {
				offset = length
			}
			offsets = append(offsets, offset)
		}

	case 16:
		offsets = make([]uint32, 0, len(ff)/2)
		for ii := 0; ii < len(ff); ii += 2 {
			frac := uint32(ff[ii])<<8 | uint32(ff[ii+1])
			offset := (frac*length + (1 << 15)) >> 16
			if offset > length {
				offset = length
			}
			offsets = append(offsets, offset)
		}

	default:
		r.err = ErrInvalidBits
		return
	}

	p.value.Set(reflect.ValueOf(offsets))
}

func (rf *reflector) scanStringSlice(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	sf := state.structStack.first()
	enc := Encoding(sf.FieldByName("Encoding").Uint())
	ss := r.ConsumeStrings(enc)
	if r.err != nil {
		return
	}
	p.value.Set(reflect.ValueOf(ss))
}

func (rf *reflector) scanStructSlice(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	elems := make([]reflect.Value, 0)
	for i := 0; r.Len() > 0; i++ {
		etyp := p.typ.Elem()
		ep := property{
			typ:   etyp,
			value: reflect.New(etyp),
			name:  fmt.Sprintf("%s[%d]", p.name, i),
		}

		rf.scanStruct(r, ep, state)
		if r.err != nil {
			return
		}

		elems = append(elems, ep.value)
	}

	slice := reflect.MakeSlice(p.typ, len(elems), len(elems))
	for ii := range elems {
		slice.Index(ii).Set(elems[ii].Elem())
	}
	p.value.Set(slice)
}

func (rf *reflector) scanString(r *reader, p property, state *state) {
	if r.err != nil {
		return
	}

	switch p.name {
	case "FrameID":
		p.value.SetString(string(state.frameID))
		return
	case "Language":
		str := r.ConsumeFixedLengthString(3, EncodingISO88591)
		p.value.SetString(str)
		return
	}

	var enc Encoding
	switch p.typ.Name() {
	case "WesternString":
		enc = EncodingISO88591
	default:
		sf := state.structStack.first()
		enc = Encoding(sf.FieldByName("Encoding").Uint())
	}

	str := r.ConsumeNextString(enc)

	if r.err != nil {
		return
	}

	p.value.SetString(str)
}

func (rf *reflector) outputStruct(w *writer, p property, state *state) {
	if p.typ.Name() == "FrameHeader" {
		return
	}

	state.structStack.push(p.value)
	if state.structStack.depth() == 1 {
		state.fieldCount = p.typ.NumField()
	}

	for i, n := 0, p.typ.NumField(); i < n; i++ {
		if state.structStack.depth() == 1 {
			state.fieldIndex = i
		}

		field := p.typ.Field(i)

		fp := property{
			typ:   field.Type,
			value: p.value.Field(i),
			name:  field.Name,
		}

		switch field.Type.Kind() {
		case reflect.Uint8:
			rf.outputUint8(w, fp, state)

		case reflect.Uint16:
			rf.outputUint16(w, fp, state)

		case reflect.Uint32:
			rf.outputUint32(w, fp, state)

		case reflect.Uint64:
			rf.outputUint64(w, fp, state)

		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				rf.outputByteSlice(w, fp, state)
			case reflect.Uint32:
				rf.outputUint32Slice(w, fp, state)
			case reflect.String:
				rf.outputStringSlice(w, fp, state)
			case reflect.Struct:
				rf.outputStructSlice(w, fp, state)
			default:
				panic(errUnknownFieldType)
			}

		case reflect.String:
			rf.outputString(w, fp, state)

		case reflect.Struct:
			rf.outputStruct(w, fp, state)

		default:
			panic(errUnknownFieldType)
		}
	}

	state.structStack.pop()
}

func (rf *reflector) outputUint8(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	value := uint8(p.value.Uint())

	if p.typ.Name() == "FrameType" {
		state.frameType = FrameType(value)
		return
	}

	bounds, hasBounds := rf.vdata.bounds[p.name]

	if hasBounds && (value < uint8(bounds.min) || value > uint8(bounds.max)) {
		w.err = bounds.err
		return
	}

	w.StoreByte(value)
	if w.err != nil {
		return
	}
}

func (rf *reflector) outputUint16(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	v := uint16(p.value.Uint())

	switch p.name {
	case "BPM":
		if v > 2*0xff {
			w.err = ErrInvalidBPM
			return
		}
		if v < 0xff {
			w.StoreByte(uint8(v))
		} else {
			w.StoreByte(0xff)
			w.StoreByte(uint8(v - 0xff))
		}
	default:
		b := []byte{byte(v >> 8), byte(v)}
		w.StoreBytes(b)
	}
}

func (rf *reflector) outputUint32(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	v := uint32(p.value.Uint())
	b := []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}

	w.StoreBytes(b)
}

func (rf *reflector) outputUint64(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	v := p.value.Uint()

	switch p.name {
	case "Counter":
		b := make([]byte, 0, 4)
		for v != 0 {
			b = append(b, byte(v&0xff))
			v = v >> 8
		}
		for len(b) < 4 {
			b = append(b, 0)
		}
		for i := len(b) - 1; i >= 0; i-- {
			w.StoreByte(b[i])
		}
	default:
		panic(errUnknownFieldType)
	}
}

func (rf *reflector) outputUint32Slice(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	if p.name != "IndexOffsets" {
		panic(errUnknownFieldType)
	}

	sf := state.structStack.first()
	length := uint32(sf.FieldByName("IndexedDataLength").Uint())
	bits := uint32(sf.FieldByName("BitsPerIndex").Uint())

	n := p.value.Len()
	slice := p.value.Slice(0, n)

	switch bits {
	case 8:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 8) / length
			if frac >= (1 << 8) {
				frac = (1 << 8) - 1
			}
			w.StoreByte(byte(frac))
		}

	case 16:
		for i := 0; i < n; i++ {
			offset := uint32(slice.Index(i).Uint())
			frac := (offset << 16) / length
			if frac >= (1 << 16) {
				frac = (1 << 16) - 1
			}
			b := []byte{byte(frac >> 8), byte(frac)}
			w.StoreBytes(b)
		}

	default:
		w.err = ErrInvalidBits
	}
}

func (rf *reflector) outputByteSlice(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	var b []byte
	reflect.ValueOf(&b).Elem().Set(p.value)
	w.StoreBytes(b)
}

func (rf *reflector) outputStringSlice(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	sf := state.structStack.first()
	enc := Encoding(sf.FieldByName("Encoding").Uint())

	var ss []string
	reflect.ValueOf(&ss).Elem().Set(p.value)
	w.StoreStrings(ss, enc)
}

func (rf *reflector) outputStructSlice(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	n := p.value.Len()
	slice := p.value.Slice(0, n)

	for i := 0; i < n; i++ {
		elem := slice.Index(i)

		ep := property{
			typ:   elem.Type(),
			value: elem,
			name:  fmt.Sprintf("%s[%d]", p.name, i),
		}

		rf.outputStruct(w, ep, state)
		if w.err != nil {
			return
		}
	}
}

func (rf *reflector) outputString(w *writer, p property, state *state) {
	if w.err != nil {
		return
	}

	v := p.value.String()

	switch p.name {
	case "FrameID":
		state.frameID = v
		return
	case "Language":
		w.StoreFixedLengthString(v, 3, EncodingISO88591)
		return
	}

	var enc Encoding
	switch p.typ.Name() {
	case "WesternString":
		enc = EncodingISO88591
	default:
		sf := state.structStack.first()
		enc = Encoding(sf.FieldByName("Encoding").Uint())
	}

	// Always terminate strings unless they are the last struct field
	// of the root level struct.
	term := state.structStack.depth() > 1 || (state.fieldIndex != state.fieldCount-1)
	w.StoreString(v, enc, term)
}
