package main

import (
	"encoding/binary"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

// statusRec builds an ACU status record for time t, as the ACU reports it.
func statusRec(t time.Time) datasets.StatusGeneral8100 {
	year, doy, tod := acuTimeParts(t)
	return datasets.StatusGeneral8100{
		Year: uint32(year),
		Time: float64(doy) + tod/(24*60*60),
	}
}

func TestStatusTime(t *testing.T) {
	t0 := time.Date(2026, 12, 31, 23, 59, 59, 500_000_000, time.UTC)
	rec := statusRec(t0)
	got := statusTime(&rec)
	if d := got.Sub(t0).Abs(); d > time.Microsecond {
		t.Errorf("statusTime: got %v, expected %v", got, t0)
	}
}

func TestCalibrateTime(t *testing.T) {
	for _, offset := range []time.Duration{-3 * time.Second, 0, 3 * time.Second} {
		tel := NewTelescope(nil)
		tel.rec = statusRec(time.Now().Add(offset))
		if err := tel.CalibrateTime(); err != nil {
			t.Fatalf("CalibrateTime(%v): %v", offset, err)
		}
		if d := (tel.pointing.tOffset - offset).Abs(); d > 100*time.Millisecond {
			t.Errorf("CalibrateTime(%v): got offset %v", offset, tel.pointing.tOffset)
		}
	}
}

func TestCalibrateTimeNoStatus(t *testing.T) {
	tel := NewTelescope(nil)
	if err := tel.CalibrateTime(); err == nil {
		t.Error("CalibrateTime with no ACU status: expected error")
	}
}

// fakeProgramTrackACU returns an ACU whose requests are recorded in *log.
// If uploadFails, uploads are rejected.
func fakeProgramTrackACU(t *testing.T, log *[]string, uploadFails bool) *ACU {
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		mu.Lock()
		*log = append(*log, strings.TrimSpace(req.URL.Path+" "+q.Get("command")+" "+q.Get("parameter")))
		mu.Unlock()
		switch {
		case req.URL.Path == "/UploadPtStack":
			if uploadFails {
				w.Write([]byte("Failed: Command failed!"))
			}
		case q.Get("identifier") == "DataSets.StatusGeneral8100":
			binary.Write(w, binary.LittleEndian, &datasets.StatusGeneral8100{QtyOfFreeProgramTrackStackPositions: maxFreeProgramTrackStack})
		case q.Get("identifier") == "DataSets.StatusCCatDetailed8100":
			binary.Write(w, binary.LittleEndian, &datasets.StatusCCatDetailed8100{})
		}
	}))
	t.Cleanup(srv.Close)
	host, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	return NewACU(host, port, port, port)
}

// The ACU won't enter program track mode with an empty stack.
func TestStartPatternUploadsFirst(t *testing.T) {
	var log []string
	tel := NewTelescope(fakeProgramTrackACU(t, &log, false))
	pattern := NewAzimuthScanPattern(time.Now().Add(15*time.Second), 1, 45, [2]float64{100, 110}, 1, 5*time.Second)
	if _, err := startPattern(t.Context(), tel, pattern); err != nil {
		t.Fatal(err)
	}
	upload := slices.Index(log, "/UploadPtStack")
	mode := slices.Index(log, "/Command SetAzElMode ProgramTrack")
	if upload < 0 || mode < 0 || mode < upload {
		t.Errorf("want upload before ProgramTrack mode, got %q", log)
	}
}

// Failures should be reset before uploading a new program track.
func TestStartPatternFailureReset(t *testing.T) {
	var log []string
	tel := NewTelescope(fakeProgramTrackACU(t, &log, false))
	pattern := NewAzimuthScanPattern(time.Now().Add(15*time.Second), 1, 45, [2]float64{100, 110}, 1, 5*time.Second)
	if _, err := startPattern(t.Context(), tel, pattern); err != nil {
		t.Fatal(err)
	}
	stop := slices.Index(log, "/Command Stop")
	reset := slices.Index(log, "/Command Failure Reset")
	clear := slices.Index(log, "/Command Clear Stack")
	if stop < 0 || reset < 0 || clear < 0 || !(stop < reset && reset < clear) {
		t.Errorf("want Stop, Failure Reset, Clear Stack, got %q", log)
	}
}

func TestStartPatternUploadFails(t *testing.T) {
	var log []string
	tel := NewTelescope(fakeProgramTrackACU(t, &log, true))
	pattern := NewAzimuthScanPattern(time.Now().Add(15*time.Second), 1, 45, [2]float64{100, 110}, 1, 5*time.Second)
	if _, err := startPattern(t.Context(), tel, pattern); err == nil {
		t.Error("expected upload error")
	}
	if slices.Contains(log, "/Command SetAzElMode ProgramTrack") {
		t.Errorf("entered ProgramTrack mode after failed upload: %q", log)
	}
}
