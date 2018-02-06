package id3

import (
	"reflect"
	"strings"
)

//
// typeMap
//

type typeMap map[string]reflect.Type

func newTypeMap(tag string) typeMap {
	m := make(typeMap, len(frameTypes))

	for _, ft := range frameTypes {
		if ft.NumField() < 1 {
			panic(errInvalidPayloadDef)
		}

		ff := ft.Field(0)
		if ff.Type.Name() != "FrameID" {
			panic(errInvalidPayloadDef)
		}

		if id, ok := ff.Tag.Lookup(tag); ok {
			m[id] = ft
		}
	}

	return m
}

func (m typeMap) Lookup(id string) reflect.Type {
	if id[0] == 'T' && id != "TXXX" && id != "TXX" {
		id = "T"
	}
	if id[0] == 'W' && id != "WXXX" && id != "WXX" {
		id = "W"
	}

	if typ, ok := m[id]; ok {
		return typ
	}

	return m["?"]
}

//
// tagList
//

type tagList map[string]bool

func (t tagList) Lookup(s string) bool {
	_, ok := t[s]
	return ok
}

var emptyTagList = make(tagList)

// Return a table of all tags on a struct field.
func getTags(t reflect.StructTag, key string) tagList {
	tag, ok := t.Lookup(key)
	if !ok {
		return emptyTagList
	}

	l := make(tagList)
	for _, t := range strings.Split(tag, ",") {
		l[t] = true
	}
	return l
}
