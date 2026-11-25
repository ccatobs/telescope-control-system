package main

import (
	"fmt"
	"maps"
	"math"
	"net/http/httptest"
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

// Parameter names match case-insensitively, as with encoding/json v1.
func TestDecodeCmdBodyCaseInsensitive(t *testing.T) {
	var cmd trackCmd
	body := `{"START_TIME": 1, "Stop_Time": 2, "RA": 3, "Dec": 4, "coordsys": "ICRS"}`
	if _, err := decodeCmdBody(strings.NewReader(body), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd != (trackCmd{StartTime: 1, StopTime: 2, RA: 3, Dec: 4, Coordsys: "ICRS"}) {
		t.Errorf("got cmd %#v", cmd)
	}

	// but underscores still matter
	if _, err := decodeCmdBody(strings.NewReader(`{"starttime": 1}`), &cmd); err == nil {
		t.Error("starttime matched start_time")
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

func TestStatusResponse(t *testing.T) {
	w := httptest.NewRecorder()
	statusResponse(w, nil)
	if got, want := w.Body.String(), `{"status":"ok","command":null}`+"\n"; got != want {
		t.Errorf("idle: got %q, want %q", got, want)
	}

	w = httptest.NewRecorder()
	statusResponse(w, &runningCmd{
		ID:        7,
		Endpoint:  "/move-to",
		Params:    moveToCmd{Azimuth: 120, Elevation: 45},
		Tags:      Tags{"scan": "s42"},
		StartedAt: 1.5,
	})
	want := `{"status":"ok","command":{"id":7,"endpoint":"/move-to","params":{"azimuth":120,"elevation":45},"tags":{"scan":"s42"},"started_at":1.5}}` + "\n"
	if got := w.Body.String(); got != want {
		t.Errorf("running: got %q, want %q", got, want)
	}
}

func TestStatusParamsPath(t *testing.T) {
	w := httptest.NewRecorder()
	statusResponse(w, &runningCmd{
		ID:       8,
		Endpoint: "/path",
		Params: statusParams(pathCmd{
			Coordsys:  "ICRS",
			Points:    [][5]float64{{0, 103, -33, 0.05, -0.05}, {60, 106, -36, 0.05, -0.05}},
			StartTime: 1615586629,
		}),
		StartedAt: 1.5,
	})
	want := `{"status":"ok","command":{"id":8,"endpoint":"/path","params":{"coordsys":"ICRS","num_points":2,"start_time":1615586629},"started_at":1.5}}` + "\n"
	if got := w.Body.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSetSPEMCmdCheck(t *testing.T) {
	for _, body := range []string{
		`{}`,                // no parameters
		`{"XX": 1}`,         // unknown parameter
		`{"IA": "1"}`,       // not a number
		`{"IA": 1, "x": 2}`, // unknown parameter
	} {
		var cmd setSPEMCmd
		_, err := decodeCmdBody(strings.NewReader(body), &cmd)
		if err == nil {
			err = cmd.Check()
		}
		if err == nil {
			t.Errorf("%q: expected error", body)
		}
	}
	if err := (setSPEMCmd{"IA": math.Inf(1)}).Check(); err == nil {
		t.Error("Inf: expected error")
	}
	if err := (setSPEMCmd{"IA": 1, "LEVE": -2}).Check(); err != nil {
		t.Error(err)
	}
}
