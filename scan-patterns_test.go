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
// housekeeping/devutils/control/src/data/scan_50_turns_scans.dat,
// except for the turnarounds between lines, which the ACU generates.
func TestHolographyScanPattern(t *testing.T) {
	// t, x, y, vx, vy, flag
	want := [][6]float64{
		// lead-in
		{0, 0, 0, 0, 0, 0},
		{0.5, -0.035471, 0, -0.212824, 0, 0},
		{1.0, -0.267514, 0, -0.691938, 0, 0},
		{1.5, -0.660615, 0, -0.809524, 0, 0},
		{2.0, -1.006097, 0, -0.501462, 0, 0},
		{2.5, -1.107386, 0, 0.072890, 0, 0},
		// start & end of the first & last lines
		{3, -1, 0, 0.285714, 0, 1},
		{10, 1, 0, 0.285714, 0, 2},
		{13, 0, -1, 0, 0.285714, 1},
		{20, 0, 1, 0, 0.285714, 2},
		{23, -0.998027, -0.062791, 0.285150, 0.017940, 1},
		{30, 0.998027, 0.062791, 0.285150, 0.017940, 2},
		{493, 0.998027, -0.062791, -0.285150, 0.017940, 1},
		{500, -0.998027, 0.062791, -0.285150, 0.017940, 1},
		// lead-out
		{500.5, -1.105201, 0.069533, -0.072746, 0.004577, 0},
		{501.0, -1.004111, 0.063173, 0.500472, -0.031487, 0},
		{501.5, -0.659311, 0.041480, 0.807926, -0.050830, 0},
		{502.0, -0.266986, 0.016797, 0.690572, -0.043447, 0},
		{502.5, -0.035401, 0.002227, 0.212404, -0.013363, 0},
		{503, 0, 0, 0, 0, 0},
	}
	const az, el = 120, 45
	t0 := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	pattern, err := NewHolographyScanPattern(t0, 50, az, el, 1, 2.0/7, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	got := map[float64]ScanPatternSample{}
	iter := pattern.Iterator()
	var prev time.Time
	n := 0
	for !pattern.Done(iter) {
		var p ScanPatternSample
		if err := pattern.Next(iter, &p); err != nil {
			t.Fatal(err)
		}
		if n > 0 && p.T.Sub(prev) < minProgramTrackStep {
			t.Errorf("t=%v: only %v after the previous point", p.T.Sub(t0), p.T.Sub(prev))
		}
		got[math.Round(p.T.Sub(t0).Seconds()*1e6)/1e6] = p
		prev = p.T
		n++
	}
	// 50 lines of 5 points, plus 29 points each in the lead-in & -out, plus the center twice
	if want := 50*5 + 2*29 + 2; n != want {
		t.Errorf("got %d points, want %d", n, want)
	}
	for _, w := range want {
		p, ok := got[w[0]]
		if !ok {
			t.Errorf("no point at t=%g", w[0])
			continue
		}
		for i, x := range []float64{p.Az - az, p.El - el, p.AzVel, p.ElVel} {
			if math.Abs(x-w[i+1]) > 2e-6 {
				t.Errorf("t=%g: got %+v, want %v", w[0], p, w)
				break
			}
		}
		if f := int8(w[5]); p.AzFlag != f || p.ElFlag != f {
			t.Errorf("t=%g: flags %d,%d, want %d", w[0], p.AzFlag, p.ElFlag, f)
		}
	}
}

func TestJerkTransition(t *testing.T) {
	for _, c := range [][5]float64{
		{0, 0, -1, 2.0 / 7, 3},    // lead-in
		{1, 2.0 / 7, 0, 0, 3},     // lead-out
		{0, 0, 2, 0, 3},           // move from rest to rest
		{5, 0.5, 5, -0.5, 2},      // reversal in place
		{0, 0.3, 0.3 * 3, 0.3, 3}, // constant velocity
	} {
		p0, v0, p1, v1, T := c[0], c[1], c[2], c[3], c[4]
		tr, err := newJerkTransition(p0, v0, p1, v1, T)
		if err != nil {
			t.Errorf("%v: %v", c, err)
			continue
		}
		if p, v := tr.at(0); math.Abs(p-p0) > 1e-9 || math.Abs(v-v0) > 1e-9 {
			t.Errorf("%v: start (%g,%g)", c, p, v)
		}
		if p, v := tr.at(T); math.Abs(p-p1) > 1e-9 || math.Abs(v-v1) > 1e-9 {
			t.Errorf("%v: end (%g,%g), want (%g,%g)", c, p, v, p1, v1)
		}
	}
}
