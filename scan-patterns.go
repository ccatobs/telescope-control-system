package main

import (
	"fmt"
	"time"

	"github.com/ccatp/antenna-control-unit/datasets"
)

// A ScanPattern represents an abstract scan pattern generator.
type ScanPattern interface {
	// Done returns true if there are no more points, false otherwise.
	Done() bool
	// Next retrieves the next point in the pattern.
	Next(*datasets.TimePositionTransfer) error
	// Time returns the time of the next point in the pattern. Undefined if no more points.
	Time() time.Time
}

// A RepeatingScanPattern executes an az,el pattern multiple times.
type RepeatingScanPattern struct {
	index, n, m int
	azs         []float64
	els         []float64
	dts         []time.Duration
	t           time.Time
}

func (scan RepeatingScanPattern) Done() bool {
	return scan.index == scan.n*scan.m
}

func (scan *RepeatingScanPattern) Time() time.Time {
	return scan.t
}

func (scan *RepeatingScanPattern) Next(p *datasets.TimePositionTransfer) error {
	t := scan.t
	j := scan.index % scan.m

	p.Day = int32(t.YearDay())
	p.TimeOfDay = DaySeconds(t)
	p.AzPosition = scan.azs[j]
	p.ElPosition = scan.els[j]

	scan.index++
	scan.t = t.Add(scan.dts[j])
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
		n:   num,
		m:   2 * m,
		azs: azs,
		els: els,
		dts: dts,
		t:   start,
	}
}

// A TrackScanPattern tracks a point on the celestial sphere.
type TrackScanPattern struct {
	t    time.Time
	tmax time.Time
	ra   float64
	dec  float64
}

func NewTrackScanPattern(t0, t1 time.Time, ra, dec float64) (*TrackScanPattern, error) {
	return &TrackScanPattern{
		t:    t0,
		tmax: t1,
		ra:   ra,
		dec:  dec,
	}, nil
}

func (track TrackScanPattern) Done() bool {
	return track.t.After(track.tmax)
}

func (track *TrackScanPattern) Time() time.Time {
	return track.t
}

func (track *TrackScanPattern) Next(p *datasets.TimePositionTransfer) error {
	t := track.t

	// convert ra,dec to az,el
	unixtime := float64(track.t.UnixNano()) * 1e-9
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
	track.t = t.Add(dt)
	return nil
}
