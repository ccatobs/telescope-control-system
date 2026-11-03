package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

func TestDatasetsEncodable(t *testing.T) {
	for name, newRec := range acuDatasets {
		rec := newRec()
		if binary.Size(rec) <= 0 {
			t.Errorf("%s: not fixed-size", name)
		}
		// the ACU's JSON rendering (NaNs as null, coded values as names)
		if _, ok := rec.(json.Marshaler); !ok {
			t.Errorf("%s: no MarshalJSON", name)
		}
		if _, err := json.Marshal(rec); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestDatasetGet(t *testing.T) {
	sent := datasets.Shutter{Time: 1.5, Year: 2026, ShutterOpen: true, DistanceToSun: 42}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if got := req.URL.Query().Get("identifier"); got != "DataSets.Shutter" {
			t.Errorf("identifier = %q", got)
		}
		var buf bytes.Buffer
		binary.Write(&buf, binary.LittleEndian, &sent)
		w.Write(buf.Bytes())
	}))
	defer srv.Close()

	host, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	acu := NewACU(host, port, port, port)
	rec := acuDatasets["Shutter"]()
	if err := acu.DatasetGet("Shutter", rec); err != nil {
		t.Fatal(err)
	}
	if got := *rec.(*datasets.Shutter); got != sent {
		t.Errorf("got %+v, want %+v", got, sent)
	}
}

func TestSPEMParameterSet(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		if id := q.Get("identifier"); id != "DataSets.CmdSPEMParameter" {
			t.Errorf("identifier = %q", id)
		}
		got = append(got, q.Get("command")+"="+q.Get("parameter"))
	}))
	defer srv.Close()

	host, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	tel := NewTelescope(NewACU(host, port, port, port))
	cmd := setSPEMCmd{"NPAE": 1e6, "IA": -0.25, "LEL_velocity": 1e-7}
	if err := cmd.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err := cmd.Start(t.Context(), tel); err != nil {
		t.Fatal(err)
	}
	want := []string{"Set Spem_IA=-0.25", "Set Spem_NPAE=1000000", "Set LEL_velocity=0.0000001"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

// spemParams should match the fields of the CmdSPEMParameter dataset.
func TestSPEMParamsMatchDataset(t *testing.T) {
	var names []string
	typ := reflect.TypeFor[datasets.CmdSPEMParameter]()
	for i := range typ.NumField() {
		tag := typ.Field(i).Tag.Get("json")
		if name, ok := strings.CutPrefix(tag, "Parameter "); ok {
			names = append(names, name)
		} else if tag == "LEL velocity" {
			names = append(names, "LEL_velocity")
		}
	}
	if !slices.Equal(names, spemParams) {
		t.Errorf("dataset has %q, spemParams has %q", names, spemParams)
	}
}

// The ACU may drop overlapping command connections, so commands must be serialized.
func TestCommandsSerialized(t *testing.T) {
	var inflight, maxInflight atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		n := inflight.Add(1)
		defer inflight.Add(-1)
		for {
			m := maxInflight.Load()
			if n <= m || maxInflight.CompareAndSwap(m, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}))
	defer srv.Close()

	host, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	acu := NewACU(host, port, port, port)
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			if err := acu.ModeSet("Stop"); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() {
			if _, err := acu.post("/UploadPtStack", "text/plain", strings.NewReader("x")); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if m := maxInflight.Load(); m != 1 {
		t.Errorf("%d commands in flight at once", m)
	}
}
