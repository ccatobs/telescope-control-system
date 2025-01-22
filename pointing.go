package main

import (
	"time"
)

type Pointing struct {
	tOffset  time.Duration
	azOffset float64
	elOffset float64
	ref      Refraction
}

func NewPointing() Pointing {
	return Pointing{}
}

func (p Pointing) Sky2Raw(az, el float64) (float64, float64) {
	// refraction
	el = p.ref.SkyEl2ObsEl(el)

	return az + p.azOffset, el + p.elOffset
}

func (p Pointing) Track2Raw(t time.Time, az, el, vaz, vel float64) (time.Time, float64, float64, float64, float64) {
	az, el = p.Sky2Raw(az, el)
	return t.Add(p.tOffset), az, el, vaz, vel
}
