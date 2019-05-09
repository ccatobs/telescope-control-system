package main

import (
	"fmt"
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

// Time returns the time of the next point in the pattern. Undefined if no more points.
func (iter ScanPatternIterator) Time() time.Time {
	return iter.t
}

// A RepeatingScanPattern executes an az,el pattern multiple times.
type RepeatingScanPattern struct {
	index, n, m int
	azs         []float64
	els         []float64
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

	iter.index++
	iter.t = t.Add(scan.dts[j])
	return nil
}

// NewAzimuthScanPattern scans back and forth in azimuth at constant elevation.
func NewAzimuthScanPattern(start time.Time, num int, el float64, az [2]float64, vel float64, turnaround time.Duration) *RepeatingScanPattern {
	const m = 5
	azs := make([]float64, 2*m)
	els := make([]float64, 2*m)
	dts := make([]time.Duration, 2*m)
	daz := (az[1] - az[0]) / (m - 1)
	dt := time.Duration(1e9*daz/vel) * time.Nanosecond
	for i := 0; i < m; i++ {
		azs[i] = az[0] + float64(i)*daz
		els[i] = el
		dts[i] = dt
	}
	for i := m; i < 2*m; i++ {
		azs[i] = az[1] - float64(i-m)*daz
		els[i] = el
		dts[i] = dt
	}
	dts[m-1] = turnaround
	dts[2*m-1] = turnaround
	return &RepeatingScanPattern{
		n:     num,
		m:     2 * m,
		azs:   azs,
		els:   els,
		dts:   dts,
		start: start,
	}
}

// A PathScanPattern follows a path of points.
type PathScanPattern struct {
	coordsys string
	points   [][3]float64
}

func NewPathScanPattern(coordsys string, points [][3]float64) *PathScanPattern {
	return &PathScanPattern{
		coordsys: coordsys,
		points:   points,
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

	var az, el float64
	switch path.coordsys {
	case "Horizon":
		az, el = x[1], x[2]
	case "ICRS":
		var err error
		az, el, err = RADec2AzEl(x[0], x[1], x[2])
		if err != nil {
			return err
		}
	}

	t := Unixtime2Time(x[0])
	p.Day = int32(t.YearDay())
	p.TimeOfDay = DaySeconds(t)
	p.AzPosition = az
	p.ElPosition = el

	iter.index++
	if i+1 < len(path.points) {
		iter.t = Unixtime2Time(path.points[i+1][0])
	}
	return nil
}

// A TrackScanPattern tracks a point on the celestial sphere.
type TrackScanPattern struct {
	tmin time.Time
	tmax time.Time
	ra   float64
	dec  float64
}

func NewTrackScanPattern(t0, t1 time.Time, ra, dec float64) (*TrackScanPattern, error) {
	return &TrackScanPattern{
		tmin: t0,
		tmax: t1,
		ra:   ra,
		dec:  dec,
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
	unixtime := float64(t.UnixNano()) * 1e-9
	az, el, err := RADec2AzEl(unixtime, track.ra, track.dec)
	if err != nil {
		return err
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
