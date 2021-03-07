package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ccatp/antenna-control-unit/datasets"
)

// A ScanPattern represents an abstract scan pattern generator.
type ScanPattern interface {
	Iterator() *ScanPatternIterator
	// Done returns true if there are no more points, false otherwise.
	Done(*ScanPatternIterator) bool
	// Next retrieves the next point in the pattern.
	Next(*ScanPatternIterator, *datasets.TimePositionTransfer) error
}

type ScanPatternIterator struct {
	index int
	t     time.Time
}

// A RepeatingScanPattern executes an az,el pattern multiple times.
type RepeatingScanPattern struct {
	index, n, m int
	azs         []float64
	els         []float64
	vazs        []float64
	vels        []float64
	fazs        []int8
	fels        []int8
	dts         []time.Duration
	start       time.Time
}

func (scan RepeatingScanPattern) Iterator() *ScanPatternIterator {
	return &ScanPatternIterator{t: scan.start}
}

func (scan RepeatingScanPattern) Done(iter *ScanPatternIterator) bool {
	return iter.index == scan.n*scan.m
}

func (scan RepeatingScanPattern) Next(iter *ScanPatternIterator, p *datasets.TimePositionTransfer) error {
	t := iter.t
	j := iter.index % scan.m

	p.Day = int32(t.YearDay())
	p.TimeOfDay = DaySeconds(t)
	p.AzPosition = scan.azs[j]
	p.ElPosition = scan.els[j]
	p.AzVelocity = scan.vazs[j]
	p.ElVelocity = scan.vels[j]
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

// A PathScanPattern follows a path of points.
type PathScanPattern struct {
	coordsys string
	points   [][5]float64
	t0       float64
}

func NewPathScanPattern(coordsys string, points [][5]float64) *PathScanPattern {
	return &PathScanPattern{
		coordsys: coordsys,
		points:   points,
		t0:       float64(time.Now().Unix()),
	}
}

func (path PathScanPattern) Iterator() *ScanPatternIterator {
	return &ScanPatternIterator{}
}

func (path PathScanPattern) Done(iter *ScanPatternIterator) bool {
	return iter.index == len(path.points)
}

func (path PathScanPattern) Next(iter *ScanPatternIterator, p *datasets.TimePositionTransfer) error {
	i := iter.index
	x := path.points[i]
	ut := x[0]
	if ut < 1 {
		return fmt.Errorf("path starts too soon")
	} else if ut < 100000 {
		ut += path.t0
	} else if ut < path.t0 {
		return fmt.Errorf("path point is in the past")
	}

	var az, el, vaz, vel float64
	switch path.coordsys {
	case "Horizon":
		az, el, vaz, vel = x[1], x[2], x[3], x[4]
	case "ICRS":
		var err error
		az, el, err = RADec2AzEl(ut, x[1], x[2])
		// XXX:TBD velocities
		log.Printf("%f RA:%3.2f DEC:%3.2f AZ:%3.2f EL:%3.2f", ut, x[1], x[2], az, el)
		if err != nil {
			return err
		}
	}

	t := Unixtime2Time(ut)
	p.Day = int32(t.YearDay())
	p.TimeOfDay = DaySeconds(t)
	p.AzPosition = az
	p.ElPosition = el
	p.AzVelocity = vaz
	p.ElVelocity = vel

	iter.index++
	return nil
}

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

func (track TrackScanPattern) Next(iter *ScanPatternIterator, p *datasets.TimePositionTransfer) error {
	t := iter.t

	// convert ra,dec to az,el
	var az, el float64

	switch track.coordsys {
	case "Horizon":
		az, el = track.ra, track.dec
	case "ICRS":
		var err error
		unixtime := float64(t.UnixNano()) * 1e-9
		az, el, err = RADec2AzEl(unixtime, track.ra, track.dec)
		log.Printf("%f RA:%3.2f DEC:%3.2f AZ:%3.2f EL:%3.2f", unixtime, track.ra, track.dec, az, el)
		if err != nil {
			return err
		}
	}

	p.Day = int32(t.YearDay())
	p.TimeOfDay = DaySeconds(t)
	p.AzPosition = az
	p.ElPosition = el

	remaining := track.tmax.Sub(t)
	if remaining < 0 {
		return fmt.Errorf("track pattern bad time")
	}

	dt := 100 * time.Second
	if 0 < remaining && remaining < dt {
		dt = remaining
	}
	iter.t = t.Add(dt)
	return nil
}
