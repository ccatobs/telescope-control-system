package main

import (
	"context"
	"encoding/binary"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

// fakeACU answers the ACU requests of a program-track upload,
// and records the commands and the uploaded points.
type fakeACU struct {
	mu       sync.Mutex
	free     int32 // free program-track stack positions
	commands []string
	uploads  [][]datasets.TimePositionTransfer
}

func newFakeACU(t *testing.T, free int32) (*fakeACU, *Telescope) {
	f := &fakeACU{free: free}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch req.URL.Path {
		case "/Values":
			switch id := req.URL.Query().Get("identifier"); id {
			case "DataSets.StatusGeneral8100":
				binary.Write(w, binary.LittleEndian, &datasets.StatusGeneral8100{QtyOfFreeProgramTrackStackPositions: f.free})
			case "DataSets.StatusCCatDetailed8100":
				binary.Write(w, binary.LittleEndian, &datasets.StatusCCatDetailed8100{})
			default:
				t.Errorf("unexpected dataset %s", id)
			}
		case "/UploadPtStack":
			file, _, err := req.FormFile("upload")
			if err != nil {
				t.Error(err)
				return
			}
			var pts []datasets.TimePositionTransfer
			for {
				var pt datasets.TimePositionTransfer
				if pt.ReadSSV(file) != nil {
					break
				}
				pts = append(pts, pt)
			}
			f.uploads = append(f.uploads, pts)
			f.free -= int32(len(pts))
		case "/Command":
			f.commands = append(f.commands, req.URL.RawQuery)
		default:
			t.Errorf("unexpected ACU request %s", req.URL)
		}
	}))
	t.Cleanup(srv.Close)
	host, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	return f, NewTelescope(NewACU(host, port, port, port))
}

func TestAppendPath(t *testing.T) {
	acu, tel := newFakeACU(t, 500)
	t0 := time.Now().Add(20 * time.Second)
	path := &runningPath{id: 42, t0: t0, coordsys: "Horizon", lastT: 10}

	res := appendPath(tel, path, [][5]float64{{11, 120, 45, 0, 0}, {12, 121, 45, 0.5, 0}})
	if res.status != http.StatusOK || res.err != nil {
		t.Fatalf("got status %d, err %v", res.status, res.err)
	}
	if res.id != 42 {
		t.Errorf("got id %d, want the id of the running path (42)", res.id)
	}

	if len(acu.commands) != 0 {
		t.Errorf("append sent ACU commands %q, want none (no Stop, no Clear+Stack)", acu.commands)
	}
	if len(acu.uploads) != 1 || len(acu.uploads[0]) != 2 {
		t.Fatalf("got uploads %v, want one upload of 2 points", acu.uploads)
	}
	for i, want := range []float64{11, 12} {
		pt := acu.uploads[0][i]
		day, tod := programTrackTime(t0.Add(Seconds2Duration(want)))
		if pt.Day != day || pt.TimeOfDay-tod > 1e-6 || tod-pt.TimeOfDay > 1e-6 {
			t.Errorf("point %d: got time (%d, %f), want start_time + %g s (%d, %f)", i, pt.Day, pt.TimeOfDay, want, day, tod)
		}
	}
	if acu.uploads[0][1].AzPosition != 121 {
		t.Errorf("point 1: got az %g, want 121", acu.uploads[0][1].AzPosition)
	}
}

func TestAppendPathNoPath(t *testing.T) {
	res := appendPath(NewTelescope(nil), nil, [][5]float64{{11, 120, 45, 0, 0}})
	if res.status != http.StatusConflict || res.err == nil {
		t.Errorf("got status %d, err %v; want %d", res.status, res.err, http.StatusConflict)
	}
	if res.id != 0 {
		t.Errorf("got id %d, want none", res.id)
	}
}

func TestAppendPathOutOfOrder(t *testing.T) {
	acu, tel := newFakeACU(t, 500)
	path := &runningPath{id: 42, t0: time.Now().Add(20 * time.Second), coordsys: "Horizon", lastT: 10}

	for name, points := range map[string][][5]float64{
		"no points":                   {},
		"at the last queued point":    {{10, 120, 45, 0, 0}},
		"before the last queued":      {{9, 120, 45, 0, 0}},
		"< 50 ms after the last":      {{10.01, 120, 45, 0, 0}},
		"out of order in the request": {{11, 120, 45, 0, 0}, {12, 120, 45, 0, 0}, {11.5, 120, 45, 0, 0}},
	} {
		res := appendPath(tel, path, points)
		if res.status != http.StatusBadRequest || res.err == nil {
			t.Errorf("%s: got status %d, err %v; want %d", name, res.status, res.err, http.StatusBadRequest)
		}
	}
	if len(acu.uploads) != 0 {
		t.Errorf("rejected points reached the ACU: %v", acu.uploads)
	}
}

func TestAppendPathLimits(t *testing.T) {
	acu, tel := newFakeACU(t, 500)
	path := &runningPath{id: 42, t0: time.Now().Add(20 * time.Second), coordsys: "Horizon", lastT: 10}

	ok := [5]float64{11, 120, 45, 0, 0}
	for name, bad := range map[string][5]float64{
		"azimuth":         {12, azimuthMax + 1, 45, 0, 0},
		"elevation":       {12, 120, elevationMin - 1, 0, 0},
		"azimuth speed":   {12, 120, 45, azimuthSpeedMax + 0.1, 0},
		"elevation speed": {12, 120, 45, 0, -elevationSpeedMax - 0.1},
	} {
		res := appendPath(tel, path, [][5]float64{ok, bad})
		if res.status != http.StatusBadRequest || res.err == nil {
			t.Errorf("%s: got status %d, err %v; want %d", name, res.status, res.err, http.StatusBadRequest)
		}
	}
	if len(acu.uploads) != 0 {
		t.Errorf("rejected points reached the ACU: %v", acu.uploads)
	}
}

// An append must not wait for room on the ACU stack: it runs in the main
// loop, which also serves /abort.
func TestAppendPathStackFull(t *testing.T) {
	acu, tel := newFakeACU(t, 1)
	path := &runningPath{id: 42, t0: time.Now(), coordsys: "Horizon", lastT: 10}

	start := time.Now()
	res := appendPath(tel, path, [][5]float64{{11, 120, 45, 0, 0}, {12, 120, 45, 0, 0}})
	if res.status != http.StatusServiceUnavailable || res.err == nil {
		t.Errorf("got status %d, err %v; want %d", res.status, res.err, http.StatusServiceUnavailable)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("append blocked for %v", d)
	}
	if len(acu.uploads) != 0 {
		t.Errorf("points reached the ACU: %v", acu.uploads)
	}
}

// Appended points must follow the path's own points on the ACU stack, so an
// append waits until the path's upload has finished.
func TestAppendPathUploadInProgress(t *testing.T) {
	acu, tel := newFakeACU(t, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel() // ends the path's upload
		for activeUploads.Load() > 0 {
			time.Sleep(time.Millisecond)
		}
	}()

	// a path of 2 points on a stack with room for 1: its upload waits for room
	cmd := pathCmd{Coordsys: "Horizon", Points: [][5]float64{{0, 120, 45, 0, 0}, {1, 120, 45, 0, 0}}, StartTime: Time2Unixtime(time.Now().Add(20 * time.Second))}
	if _, err := cmd.Start(ctx, tel); err != nil {
		t.Fatal(err)
	}
	acu.mu.Lock()
	acu.free = 500 // the ACU has made room, but the path's 2nd point is not queued yet
	acu.mu.Unlock()

	path := &runningPath{id: 42, t0: jsontime(cmd.StartTime), coordsys: "Horizon", lastT: 1}
	res := appendPath(tel, path, [][5]float64{{2, 120, 45, 0, 0}})
	if res.status != http.StatusServiceUnavailable || res.err == nil {
		t.Errorf("got status %d, err %v; want %d", res.status, res.err, http.StatusServiceUnavailable)
	}
}

// After the path's own upload failed, the stack holds a gap or no points of
// the path, so an append must not go through.
func TestAppendPathUploadFailed(t *testing.T) {
	acu, tel := newFakeACU(t, 10000)
	t.Cleanup(func() { uploadFailed.Store(false) })

	// 101 points pass Check, which reads the first 100; point 100 breaks the azimuth limit
	points := make([][5]float64, 101)
	for i := range points {
		points[i] = [5]float64{float64(i), 120, 45, 0, 0}
	}
	points[100][1] = azimuthMax + 1
	cmd := pathCmd{Coordsys: "Horizon", Points: points, StartTime: Time2Unixtime(time.Now().Add(20 * time.Second))}
	if err := cmd.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err := cmd.Start(context.Background(), tel); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(time.Second); activeUploads.Load() > 0; {
		if time.Now().After(deadline) {
			t.Fatal("path upload did not end")
		}
		time.Sleep(time.Millisecond)
	}

	path := &runningPath{id: 42, t0: jsontime(cmd.StartTime), coordsys: "Horizon", lastT: 100}
	res := appendPath(tel, path, [][5]float64{{101, 120, 45, 0, 0}})
	if res.status != http.StatusConflict || res.err == nil {
		t.Errorf("got status %d, err %v; want %d", res.status, res.err, http.StatusConflict)
	}
	if len(acu.uploads) != 0 {
		t.Errorf("points reached the ACU: %d uploads", len(acu.uploads))
	}
}

// The running path's last point stays on the stack, so more points than the
// rest of the stack never fit: a retry can't help.
func TestAppendPathTooMany(t *testing.T) {
	acu, tel := newFakeACU(t, maxFreeProgramTrackStack-1)
	path := &runningPath{id: 42, t0: time.Now().Add(20 * time.Second), coordsys: "Horizon", lastT: 10}

	points := make([][5]float64, maxFreeProgramTrackStack)
	for i := range points {
		points[i] = [5]float64{11 + float64(i), 120, 45, 0, 0}
	}
	res := appendPath(tel, path, points)
	if res.status != http.StatusBadRequest || res.err == nil {
		t.Errorf("got status %d, err %v; want %d", res.status, res.err, http.StatusBadRequest)
	}
	if len(acu.uploads) != 0 {
		t.Errorf("points reached the ACU: %d uploads", len(acu.uploads))
	}
}
