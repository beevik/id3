package id3

import "reflect"

type typeMap map[string]reflect.Type

func newTypeMap(tag string) typeMap {
	m := make(typeMap, len(frameTypes))

	for _, ft := range frameTypes {
		if ft.NumField() < 1 {
			panic(errInvalidPayloadDef)
		}

		ff := ft.Field(0)
		if ff.Name != "frameID" {
			panic(errInvalidPayloadDef)
		}

		if id, ok := ff.Tag.Lookup(tag); ok {
			m[id] = ft
		}
	}

	return m
}

func (m typeMap) Lookup24(id string) reflect.Type {
	if id[0] == 'T' && id != "TXXX" {
		id = "T___"
	}
	if id[0] == 'W' && id != "WXXX" {
		id = "W___"
	}

	if typ, ok := m[id]; ok {
		return typ
	}

	return m["????"]
}
