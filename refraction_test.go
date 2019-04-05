package main

import (
	"math"
	"testing"
)

func TestRefraction(t *testing.T) {
	ref, err := NewRefraction(550, 10, 0.5, 1300)
	if err != nil {
		t.Error(err)
	}
	obsEl := 20.0 // zd = 70
	skyEl := ref.ObsEl2SkyEl(obsEl)

	// the correction should be ~100 arcsec (arxiv:1411.1617, fig 5)
	r := 3600 * (obsEl - skyEl)
	if math.Abs(r-100) > 1 {
		t.Error(r)
	}

	// test roundtrip
	obsEl2 := ref.SkyEl2ObsEl(skyEl)
	if math.Abs(obsEl-obsEl2) > 1e-7 {
		t.Error(obsEl, skyEl, obsEl2)
	}
}
