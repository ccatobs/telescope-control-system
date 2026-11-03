package main

import (
	"fmt"
	"log/slog"
	"math"
	"time"
)

type ScanPatternSample struct {
	T      time.Time `json:"t"`
	Az     float64   `json:"az"`
	El     float64   `json:"el"`
	AzVel  float64   `json:"vaz"`
	ElVel  float64   `json:"vel"`
	AzFlag int8      `json:"azFlag"`
	ElFlag int8      `json:"elFlag"`
}

// A ScanPattern represents an abstract scan pattern generator.
type ScanPattern interface {
	Iterator() *ScanPatternIterator
	// Done returns true if there are no more points, false otherwise.
	Done(*ScanPatternIterator) bool
	// Next retrieves the next point in the pattern.
	Next(*ScanPatternIterator, *ScanPatternSample) error
}

type ScanPatternIterator struct {
	index int
	t     time.Time
}

// A RepeatingScanPattern executes an az,el pattern multiple times.
type RepeatingScanPattern struct {
	n, m  int
	azs   []float64
	els   []float64
	vazs  []float64
	vels  []float64
	fazs  []int8
	fels  []int8
	dts   []time.Duration
	start time.Time
}

func (scan RepeatingScanPattern) Iterator() *ScanPatternIterator {
	return &ScanPatternIterator{t: scan.start}
}

func (scan RepeatingScanPattern) Done(iter *ScanPatternIterator) bool {
	return iter.index == scan.n*scan.m
}

func (scan RepeatingScanPattern) Next(iter *ScanPatternIterator, p *ScanPatternSample) error {
	t := iter.t
	j := iter.index % scan.m

	p.T = t
	p.Az = scan.azs[j]
	p.El = scan.els[j]
	p.AzVel = scan.vazs[j]
	p.ElVel = scan.vels[j]
	p.AzFlag = scan.fazs[j]
	p.ElFlag = scan.fels[j]

	iter.index++
	iter.t = t.Add(scan.dts[j])
	return nil
}

// NewAzimuthScanPattern scans back and forth in azimuth at constant elevation.
func NewAzimuthScanPattern(start time.Time, num int, el float64, az [2]float64, speed float64, turnaround time.Duration) *RepeatingScanPattern {
	const m = 5
	azs := make([]float64, 2*m)
	els := make([]float64, 2*m)
	vazs := make([]float64, 2*m)
	vels := make([]float64, 2*m)
	fazs := make([]int8, 2*m)
	fels := make([]int8, 2*m)
	dts := make([]time.Duration, 2*m)
	daz := (az[1] - az[0]) / (m - 1)
	vel := math.Copysign(speed, daz)
	dt := time.Duration(1e9*daz/vel) * time.Nanosecond
	for i := 0; i < m; i++ {
		azs[i] = az[0] + float64(i)*daz
		els[i] = el
		vazs[i] = vel
		vels[i] = 0
		fazs[i] = 1 // linear interpolation
		fels[i] = 0
		dts[i] = dt
	}
	for i := m; i < 2*m; i++ {
		azs[i] = az[1] - float64(i-m)*daz
		els[i] = el
		vazs[i] = -vel
		vels[i] = 0
		fazs[i] = 1 // linear interpolation
		fels[i] = 0
		dts[i] = dt
	}
	dts[m-1] = turnaround
	dts[2*m-1] = turnaround
	fazs[m-1] = 2 // turnaround flag
	fazs[2*m-1] = 2
	return &RepeatingScanPattern{
		n:     num,
		m:     2 * m,
		azs:   azs,
		els:   els,
		vazs:  vazs,
		vels:  vels,
		fazs:  fazs,
		fels:  fels,
		dts:   dts,
		start: start,
	}
}

// NewHolographyScanPattern scans num straight lines through (az,el), each from -radius
// to +radius at constant speed. Lines come in perpendicular pairs, and each pair
// is rotated 180/num degrees from the last. The ACU generates the turnarounds.
// The pattern starts & ends at rest at (az,el), so the ACU can ease into & out of it.
func NewHolographyScanPattern(start time.Time, num int, az, el, radius, speed float64, turnaround time.Duration) *RepeatingScanPattern {
	const m = 5 // points per line
	n := num*m + 2
	azs := make([]float64, n)
	els := make([]float64, n)
	vazs := make([]float64, n)
	vels := make([]float64, n)
	fazs := make([]int8, n)
	fels := make([]int8, n)
	dts := make([]time.Duration, n)

	// start at rest at the center, then turn onto the first line
	azs[0], els[0] = az, el
	fazs[0], fels[0] = 2, 2 // turnaround flag
	dts[0] = turnaround

	ds := 2 * radius / (m - 1)
	dt := Seconds2Duration(ds / speed)
	for k := 0; k < num; k++ {
		theta := float64(k/2)*180/float64(num) + float64(k%2)*90
		sin, cos := math.Sincos(theta * math.Pi / 180)
		for i := 0; i < m; i++ {
			j := 1 + k*m + i
			s := -radius + float64(i)*ds
			azs[j] = az + s*cos
			els[j] = el + s*sin
			vazs[j] = speed * cos
			vels[j] = speed * sin
			fazs[j] = 1 // linear interpolation
			fels[j] = 1
			dts[j] = dt
		}
		j := 1 + k*m + m - 1
		dts[j] = turnaround
		fazs[j] = 2 // turnaround flag
		fels[j] = 2
	}

	// end at rest at the center
	azs[n-1], els[n-1] = az, el
	fazs[n-1], fels[n-1] = 1, 1

	return &RepeatingScanPattern{
		n:     1,
		m:     n,
		azs:   azs,
		els:   els,
		vazs:  vazs,
		vels:  vels,
		fazs:  fazs,
		fels:  fels,
		dts:   dts,
		start: start,
	}
}

// A PathScanPattern follows a path of points.
type PathScanPattern struct {
	coordsys string
	points   [][5]float64
	t0       time.Time
}

func NewPathScanPattern(t0 time.Time, points [][5]float64, coordsys string) *PathScanPattern {
	return &PathScanPattern{
		coordsys: coordsys,
		points:   points,
		t0:       t0,
	}
}

func (path PathScanPattern) Iterator() *ScanPatternIterator {
	return &ScanPatternIterator{}
}

func (path PathScanPattern) Done(iter *ScanPatternIterator) bool {
	return iter.index == len(path.points)
}

func (path PathScanPattern) Next(iter *ScanPatternIterator, p *ScanPatternSample) error {
	i := iter.index
	x := path.points[i]
	t := path.t0.Add(Seconds2Duration(x[0]))

	var az, el, vaz, vel float64
	switch path.coordsys {
	case "Horizon":
		az, el, vaz, vel = x[1], x[2], x[3], x[4]
	case "ICRS", "Galactic":
		var err error
		az, el, vaz, vel, err = celestial2AzElVel(path.coordsys, t, x[1], x[2])
		slog.Debug("path point", "t", t, "coordsys", path.coordsys, "lon", x[1], "lat", x[2], "az", az, "el", el, "vaz", vaz, "vel", vel)
		if err != nil {
			return err
		}
	}

	p.T = t
	p.Az = az
	p.El = el
	p.AzVel = vaz
	p.ElVel = vel

	iter.index++
	return nil
}

const (
	// time between track points; the ACU spline-interpolates between them
	trackStep = 10 * time.Second
	// "The minimum time interval between two samples is 0.05 s."
	minProgramTrackStep = 50 * time.Millisecond
)

// A TrackScanPattern tracks a point on the celestial sphere.
type TrackScanPattern struct {
	tmin     time.Time
	tmax     time.Time
	ra       float64
	dec      float64
	coordsys string
}

func NewTrackScanPattern(t0, t1 time.Time, ra, dec float64, coordsys string) (*TrackScanPattern, error) {
	return &TrackScanPattern{
		tmin:     t0,
		tmax:     t1,
		ra:       ra,
		dec:      dec,
		coordsys: coordsys,
	}, nil
}

func (track TrackScanPattern) Iterator() *ScanPatternIterator {
	return &ScanPatternIterator{t: track.tmin}
}

func (track TrackScanPattern) Done(iter *ScanPatternIterator) bool {
	return iter.t.After(track.tmax)
}

func (track TrackScanPattern) Next(iter *ScanPatternIterator, p *ScanPatternSample) error {
	t := iter.t

	// convert ra,dec to az,el
	var az, el, vaz, vel float64

	switch track.coordsys {
	case "Horizon":
		az, el = track.ra, track.dec
	case "ICRS", "Galactic":
		var err error
		az, el, vaz, vel, err = celestial2AzElVel(track.coordsys, t, track.ra, track.dec)
		slog.Debug("track point", "t", t, "coordsys", track.coordsys, "lon", track.ra, "lat", track.dec, "az", az, "el", el, "vaz", vaz, "vel", vel)
		if err != nil {
			return err
		}
	}

	p.T = t
	p.Az = az
	p.El = el
	p.AzVel = vaz
	p.ElVel = vel

	remaining := track.tmax.Sub(t)
	if remaining < 0 {
		return fmt.Errorf("track pattern bad time")
	}

	// Step by trackStep, landing the last point on tmax. If the final step
	// would be shorter than minProgramTrackStep, stretch the one before it.
	dt := trackStep
	if 0 < remaining && remaining < trackStep+minProgramTrackStep {
		dt = remaining
	}
	iter.t = t.Add(dt)
	return nil
}

// celestial2AzEl converts ICRS (ra,dec) or Galactic (l,b) coordinates to az,el at time t.
func celestial2AzEl(coordsys string, t time.Time, lon, lat float64) (float64, float64, error) {
	ut := Time2Unixtime(t)
	switch coordsys {
	case "ICRS":
		return RADec2AzEl(ut, lon, lat)
	case "Galactic":
		return GalLonLat2AzEl(ut, lon, lat)
	}
	return 0, 0, fmt.Errorf("unknown coordsys %q", coordsys)
}

// celestial2AzElVel is like celestial2AzEl, but also returns the az,el velocities [deg/s],
// estimated by central difference.
func celestial2AzElVel(coordsys string, t time.Time, lon, lat float64) (az, el, vaz, vel float64, err error) {
	const h = 0.5 // [s]
	az, el, err = celestial2AzEl(coordsys, t, lon, lat)
	if err != nil {
		return
	}
	az0, el0, err := celestial2AzEl(coordsys, t.Add(-Seconds2Duration(h)), lon, lat)
	if err != nil {
		return
	}
	az1, el1, err := celestial2AzEl(coordsys, t.Add(Seconds2Duration(h)), lon, lat)
	if err != nil {
		return
	}
	daz := math.Remainder(az1-az0, 360) // handle wrap at 0/360
	vaz = daz / (2 * h)
	vel = (el1 - el0) / (2 * h)
	return
}
