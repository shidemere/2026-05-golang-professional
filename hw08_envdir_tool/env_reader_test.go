package main

import (
	"reflect"
	"testing"
)

func TestReadDir(t *testing.T) {
	want := Environment{
		"BAR":   {Value: "bar"},
		"EMPTY": {Value: ""},
		"FOO":   {Value: "   foo\nwith new line"},
		"HELLO": {Value: `"hello"`},
		"UNSET": {NeedRemove: true},
	}

	got, err := ReadDir("testdata/env")
	if err != nil {
		t.Fatalf("ReadDir returned an error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadDir result mismatch:\ngot:  %#v\nwant: %#v", got, want)
	}
}
