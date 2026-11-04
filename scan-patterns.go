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
// is rotated 180/num degrees from the last. The ACU generates the turnarounds
// between lines, but the lead-in from rest at (az,el) to the first line, and the
// lead-out from the last line back to rest at (az,el), are sampled here, since
// the ACU only recognizes turnarounds that reverse direction in place.
func NewHolographyScanPattern(start time.Time, num int, az, el, radius, speed float64, turnaround time.Duration) (*RepeatingScanPattern, error) {
	const m = 5 // points per line
	p := &RepeatingScanPattern{n: 1, start: start}
	add := func(az, el, vaz, vel float64, flag int8, dt time.Duration) {
		p.azs = append(p.azs, az)
		p.els = append(p.els, el)
		p.vazs = append(p.vazs, vaz)
		p.vels = append(p.vels, vel)
		p.fazs = append(p.fazs, flag)
		p.fels = append(p.fels, flag)
		p.dts = append(p.dts, dt)
		p.m++
	}
	direction := func(k int) (float64, float64) {
		theta := float64(k/2)*180/float64(num) + float64(k%2)*90
		return math.Sincos(theta * math.Pi / 180)
	}

	// addLead adds the samples strictly between (az0,el0) at t=0 and (az1,el1) at t=turnaround.
	n := int(math.Ceil(turnaround.Seconds() / leadStep.Seconds()))
	step := turnaround / time.Duration(n)
	addLead := func(az0, el0, vaz0, vel0, az1, el1, vaz1, vel1 float64) error {
		T := turnaround.Seconds()
		trAz, err := newJerkTransition(az0, vaz0, az1, vaz1, T)
		if err != nil {
			return fmt.Errorf("azimuth: %w", err)
		}
		trEl, err := newJerkTransition(el0, vel0, el1, vel1, T)
		if err != nil {
			return fmt.Errorf("elevation: %w", err)
		}
		if err := trAz.checkLimits(azimuthAccelMax, azimuthJerkMax); err != nil {
			return fmt.Errorf("azimuth: %w", err)
		}
		if err := trEl.checkLimits(elevationAccelMax, elevationJerkMax); err != nil {
			return fmt.Errorf("elevation: %w", err)
		}
		for i := 1; i < n; i++ {
			t := float64(i) * step.Seconds()
			a, va := trAz.at(t)
			e, ve := trEl.at(t)
			add(a, e, va, ve, 0, step)
		}
		return nil
	}

	// lead-in: from rest at the center to the start of the first line
	sin, cos := direction(0)
	add(az, el, 0, 0, 0, step)
	err := addLead(az, el, 0, 0, az-radius*cos, el-radius*sin, speed*cos, speed*sin)
	if err != nil {
		return nil, fmt.Errorf("lead-in: %w", err)
	}

	ds := 2 * radius / (m - 1)
	dt := Seconds2Duration(ds / speed)
	for k := 0; k < num; k++ {
		sin, cos = direction(k)
		for i := 0; i < m-1; i++ {
			s := -radius + float64(i)*ds
			add(az+s*cos, el+s*sin, speed*cos, speed*sin, 1, dt) // linear interpolation
		}
		if k < num-1 {
			add(az+radius*cos, el+radius*sin, speed*cos, speed*sin, 2, turnaround) // turnaround flag
		} else {
			add(az+radius*cos, el+radius*sin, speed*cos, speed*sin, 1, step)
		}
	}

	// lead-out: from the end of the last line to rest at the center
	err = addLead(az+radius*cos, el+radius*sin, speed*cos, speed*sin, az, el, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("lead-out: %w", err)
	}
	add(az, el, 0, 0, 0, step)

	return p, nil
}

// time between lead-in & lead-out samples
const leadStep = 100 * time.Millisecond

// A jerkTransition moves one axis from (p0,v0) at t=0 to (p1,v1) at t=T, with zero
// acceleration at both ends, using a jerk of j, -j, j, switching at t1 and t1+T/2.
// This is the profile used by the turnarounds in the housekeeping holography
// pattern scan_50_turns_scans.dat.
type jerkTransition struct {
	p0, v0, j, t1, T float64
}

func newJerkTransition(p0, v0, p1, v1, T float64) (jerkTransition, error) {
	tr := jerkTransition{p0: p0, v0: v0, T: T}
	dv := v1 - v0
	dx := p1 - p0 - v0*T
	if dv == 0 && dx == 0 {
		return tr, nil
	}
	// for unit jerk, dv & dx at T as functions of t1
	unit := func(t1 float64) (float64, float64) {
		x, v := jerkTransition{j: 1, t1: t1, T: T}.at(T)
		return v, x
	}
	// find t1 in [0,T/2] where (dv,dx) is parallel to unit(t1),
	// choosing the root with the smallest jerk
	h := func(t1 float64) float64 {
		uv, ux := unit(t1)
		return uv*dx - ux*dv
	}
	const ngrid = 1000
	best := math.Inf(1)
	for i := 0; i < ngrid; i++ {
		a := T / 2 * float64(i) / ngrid
		b := T / 2 * float64(i+1) / ngrid
		ha, hb := h(a), h(b)
		if ha*hb > 0 {
			continue
		}
		for range 60 {
			c := (a + b) / 2
			if hc := h(c); ha*hc <= 0 {
				b = c
			} else {
				a, ha = c, hc
			}
		}
		t1 := (a + b) / 2
		uv, ux := unit(t1)
		j := (dv*uv + dx*ux) / (uv*uv + ux*ux)
		if math.Abs(j) < math.Abs(best) {
			best = j
			tr.t1 = t1
		}
	}
	if math.IsInf(best, 0) {
		return tr, fmt.Errorf("no constant-jerk transition in %g s", T)
	}
	tr.j = best
	return tr, nil
}

// at returns the position & velocity at time t in [0,T].
func (tr jerkTransition) at(t float64) (float64, float64) {
	p, v, a := tr.p0, tr.v0, 0.0
	t0 := 0.0
	for _, phase := range [...]struct{ end, j float64 }{
		{tr.t1, tr.j},
		{tr.t1 + tr.T/2, -tr.j},
		{tr.T, tr.j},
	} {
		if t <= t0 {
			break
		}
		// phases may be empty, e.g. if t1 = 0
		if dt := math.Min(t, phase.end) - t0; dt > 0 {
			p += v*dt + a*dt*dt/2 + phase.j*dt*dt*dt/6
			v += a*dt + phase.j*dt*dt/2
			a += phase.j * dt
		}
		t0 = phase.end
	}
	return p, v
}

func (tr jerkTransition) checkLimits(accelMax, jerkMax float64) error {
	if j := math.Abs(tr.j); j > jerkMax {
		return fmt.Errorf("jerk (%g) exceeds %g; try a longer turnaround_time", j, jerkMax)
	}
	// peak acceleration is at t1 or t1+T/2
	if a := math.Abs(tr.j) * math.Max(tr.t1, tr.T/2-tr.t1); a > accelMax {
		return fmt.Errorf("acceleration (%g) exceeds %g; try a longer turnaround_time", a, accelMax)
	}
	return nil
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
