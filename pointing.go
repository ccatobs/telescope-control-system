package main

type Pointing struct {
	azOffset float64
	elOffset float64
	ref      Refraction
}

func NewPointing() Pointing {
	return Pointing{}
}

func (p Pointing) Sky2Raw(az, el, vaz, vel float64) (float64, float64, float64, float64) {
	// refraction
	el = p.ref.SkyEl2ObsEl(el)

	return az + p.azOffset, el + p.elOffset, vaz, vel
}
