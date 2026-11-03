package main

import (
	"math"
	"testing"
	"time"
)

// Track velocities should match the motion between consecutive points.
func TestTrackVelocities(t *testing.T) {
	t0 := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	for _, coordsys := range []string{"ICRS", "Galactic"} {
		pattern, err := NewTrackScanPattern(t0, t0.Add(time.Hour), 83.6, -5.4, coordsys)
		if err != nil {
			t.Fatal(err)
		}
		iter := pattern.Iterator()
		var prev ScanPatternSample
		for i := 0; !pattern.Done(iter); i++ {
			var p ScanPatternSample
			if err := pattern.Next(iter, &p); err != nil {
				t.Fatal(err)
			}
			if p.AzVel == 0 || p.ElVel == 0 {
				t.Fatalf("%s: zero velocity at %v", coordsys, p.T)
			}
			if i > 0 {
				dt := p.T.Sub(prev.T).Seconds()
				vaz := math.Remainder(p.Az-prev.Az, 360) / dt
				vel := (p.El - prev.El) / dt
				avgVaz := (p.AzVel + prev.AzVel) / 2
				avgVel := (p.ElVel + prev.ElVel) / 2
				if math.Abs(vaz-avgVaz) > 1e-5 || math.Abs(vel-avgVel) > 1e-5 {
					t.Errorf("%s at %v: velocity (%g,%g), want ~(%g,%g)", coordsys, p.T, avgVaz, avgVel, vaz, vel)
				}
			}
			prev = p
		}
	}
}

// Track points should be trackStep apart, never closer than the ACU minimum,
// and the last point should land on the stop time.
func TestTrackSpacing(t *testing.T) {
	t0 := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	for _, dur := range []time.Duration{
		time.Hour,
		time.Hour + 10*time.Millisecond,
		time.Hour + 5*time.Second,
		3 * time.Second,
	} {
		pattern, err := NewTrackScanPattern(t0, t0.Add(dur), 83.6, -5.4, "ICRS")
		if err != nil {
			t.Fatal(err)
		}
		var ts []time.Time
		iter := pattern.Iterator()
		for !pattern.Done(iter) {
			var p ScanPatternSample
			if err := pattern.Next(iter, &p); err != nil {
				t.Fatal(err)
			}
			ts = append(ts, p.T)
		}
		if last := ts[len(ts)-1]; !last.Equal(t0.Add(dur)) {
			t.Errorf("%v: last point at %v, want %v", dur, last, t0.Add(dur))
		}
		for i := 1; i < len(ts); i++ {
			dt := ts[i].Sub(ts[i-1])
			if dt < minProgramTrackStep || dt >= trackStep+minProgramTrackStep {
				t.Errorf("%v: step %d is %v", dur, i, dt)
			}
		}
	}
}

// The holography scan should reproduce the pattern in
// housekeeping/devutils/control/src/data/scan_50_minimal.dat.
func TestHolographyScanPattern(t *testing.T) {
	// t, x, y, vx, vy at the start & end of the pattern & of the first & last lines
	want := [][5]float64{
		{0, 0, 0, 0, 0},
		{3, -1, 0, 0.285714, 0},
		{10, 1, 0, 0.285714, 0},
		{13, 0, -1, 0, 0.285714},
		{20, 0, 1, 0, 0.285714},
		{23, -0.998027, -0.062791, 0.285150, 0.017940},
		{30, 0.998027, 0.062791, 0.285150, 0.017940},
		{493, 0.998027, -0.062791, -0.285150, 0.017940},
		{500, -0.998027, 0.062791, -0.285150, 0.017940},
		{503, 0, 0, 0, 0},
	}
	const az, el = 120, 45
	t0 := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	pattern := NewHolographyScanPattern(t0, 50, az, el, 1, 2.0/7, 3*time.Second)
	got := map[float64]ScanPatternSample{}
	iter := pattern.Iterator()
	n := 0
	for !pattern.Done(iter) {
		var p ScanPatternSample
		if err := pattern.Next(iter, &p); err != nil {
			t.Fatal(err)
		}
		got[math.Round(p.T.Sub(t0).Seconds()*1e6)/1e6] = p
		n++
	}
	if n != 50*5+2 {
		t.Errorf("got %d points, want %d", n, 50*5+2)
	}
	for _, w := range want {
		p, ok := got[w[0]]
		if !ok {
			t.Errorf("no point at t=%g", w[0])
			continue
		}
		for i, x := range []float64{p.Az - az, p.El - el, p.AzVel, p.ElVel} {
			if math.Abs(x-w[i+1]) > 2e-6 {
				t.Errorf("t=%g: got %v, want %v", w[0], p, w)
				break
			}
		}
		// the start and the last point of each line are turnarounds
		var f int8 = 1
		if int(w[0])%10 == 0 {
			f = 2
		}
		if p.AzFlag != f || p.ElFlag != f {
			t.Errorf("t=%g: flags %d,%d, want %d", w[0], p.AzFlag, p.ElFlag, f)
		}
	}
}
