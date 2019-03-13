package main

type Pointing struct {
	azOffset float64
	elOffset float64
}

func NewPointing() Pointing {
	return Pointing{}
}

func (p Pointing) Sky2Raw(az, el float64) (float64, float64) {
	return az + p.azOffset, el + p.elOffset
}
