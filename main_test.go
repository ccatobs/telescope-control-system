package main

import (
	"fmt"
	"maps"
	"strings"
	"testing"
)

func TestDecodeCmdBody(t *testing.T) {
	var cmd moveToCmd
	body := `{"azimuth": 120, "elevation": 45, "tags": {"obs": "abc", "note": "x y"}}`
	tags, err := decodeCmdBody(strings.NewReader(body), &cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cmd != (moveToCmd{Azimuth: 120, Elevation: 45}) {
		t.Errorf("got cmd %#v", cmd)
	}
	if !maps.Equal(tags, Tags{"obs": "abc", "note": "x y"}) {
		t.Errorf("got tags %#v", tags)
	}
}

func TestDecodeCmdBodyErrors(t *testing.T) {
	for _, body := range []string{
		``,                                    // missing body
		`{"azimuth": 120, "bogus": 1}`,        // unknown field
		`{"azimuth": 120, "tags": {"n": 1}}`,  // non-string tag value
		`{"azimuth": 120, "tags": ["a"]}`,     // tags not an object
		`{"azimuth": 120, "tags": {"": "a"}}`, // empty tag key
	} {
		var cmd moveToCmd
		_, err := decodeCmdBody(strings.NewReader(body), &cmd)
		if err == nil {
			t.Errorf("%q: expected error", body)
		}
	}
}

func TestTagsCheck(t *testing.T) {
	ok := Tags{strings.Repeat("k", maxTagKeyLen): strings.Repeat("v", maxTagValueLen)}
	if err := ok.Check(); err != nil {
		t.Errorf("tags at limits: %v", err)
	}

	many := Tags{}
	for i := range maxTags + 1 {
		many[fmt.Sprint(i)] = "x"
	}
	for _, tags := range []Tags{
		many,
		{"": "x"},
		{strings.Repeat("k", maxTagKeyLen+1): "x"},
		{"k": strings.Repeat("v", maxTagValueLen+1)},
	} {
		if err := tags.Check(); err == nil {
			t.Errorf("expected error for tags with %d keys", len(tags))
		}
	}
}

func TestCheckNoBody(t *testing.T) {
	if err := checkNoBody(strings.NewReader("")); err != nil {
		t.Errorf("empty body: %v", err)
	}
	if err := checkNoBody(strings.NewReader(`{"tags": {"reason": "wind"}}`)); err == nil {
		t.Error("non-empty body: expected error")
	}
}
