package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

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
