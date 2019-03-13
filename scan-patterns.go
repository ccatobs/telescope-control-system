package main

import (
	"time"

	"github.com/ccatp/antenna-control-unit/datasets"
)

// A ScanPattern represents an abstract scan pattern generator.
type ScanPattern interface {
	// Next retrieves the next point in the pattern. Returns false if there are no more points.
	Next(*datasets.TimePositionTransfer) bool
	// Time returns the time of the next point in the pattern, or of the final point if done.
	Time() time.Time
}

type RepeatingScanPattern struct {
	index, n, m int
	azs         []float64
	els         []float64
	dts         []time.Duration
	t           time.Time
}

func (scan *RepeatingScanPattern) Time() time.Time {
	return scan.t
}

func (scan *RepeatingScanPattern) Next(p *datasets.TimePositionTransfer) bool {
	t := scan.t
	i := scan.index / scan.m
	j := scan.index % scan.m

	if i == scan.n {
		return false
	}

	p.Day = int32(t.YearDay())
	p.TimeOfDay = DaySeconds(t)
	p.AzPosition = scan.azs[j]
	p.ElPosition = scan.els[j]

	scan.index++
	scan.t = t.Add(scan.dts[j])
	return true
}

func AzimuthScanPattern(start time.Time, num int, el float64, az [2]float64, vel float64, turnaround time.Duration) *RepeatingScanPattern {
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

type DaisyScanPattern struct {
	// XXX:TBD
}

func (daisy *DaisyScanPattern) Next(p *datasets.TimePositionTransfer) bool {
	// XXX:TBD
	return false
}
