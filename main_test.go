package main

import (
	"fmt"
	"maps"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// Appends faster than the 1 Hz status update must not delay isDone.
func TestWaitForCmdWhileAppending(t *testing.T) {
	for name, isDone := range map[string]IsDoneFunc{
		"finished path": func(*Telescope) (bool, error) { return true, nil },
		"failed path":   func(*Telescope) (bool, error) { return true, fmt.Errorf("upload failed") },
	} {
		_, tel := newFakeACU(t, 0) // a full stack: each append gets 503
		path := &runningPath{id: 42, t0: time.Now().Add(20 * time.Second), coordsys: "Horizon", lastT: 10}
		appends := make(chan appendReq)
		finished := make(chan struct{})
		go func() {
			waitForCmd(42, tel, isDone, func() {}, path, nil, appends)
			close(finished)
		}()

		// the client retries the append at 10 Hz
		deadline := time.After(3 * time.Second)
	retries:
		for {
			r := appendReq{[][5]float64{{11, 120, 45, 0, 0}}, make(chan appendResult)}
			select {
			case appends <- r:
				<-r.result
			case <-finished:
				break retries
			case <-deadline:
				t.Fatalf("%s: no status update within 3 s of appends at 10 Hz", name)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
