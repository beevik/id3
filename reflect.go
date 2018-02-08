package id3

import (
	"reflect"
	"strings"
)

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
