package main

import (
	"math"
)

// A Refraction represents the atmospheric refraction model
// dZ = A tan Z + B tan^3 Z.
//
// Z is the "observed" zenith distance (i.e. affected by refraction)
// and dZ is what to add to Z to give the "topocentric" (i.e. in vacuo)
// zenith distance.
//
type Refraction struct {
	a float64
	b float64
}

// NewRefraction determines the constants A and B in the atmospheric refraction model.
//
//   phpa     pressure at the observer (hPa = millibar)
//   tc       ambient temperature at the observer (deg C)
//   rh       relative humidity at the observer (range 0-1)
//   wl       wavelength (micrometers)
//
// https://github.com/liberfa/erfa/blob/b1c4e0ccd11f3adb66508ba6c16e5ed214154af2/src/refco.c
//
func NewRefraction(phpa, tc, rh, wl float64) (Refraction, error) {
	ps := math.Pow(10.0, (0.7859+0.03477*tc)/(1.0+0.00412*tc)) * (1.0 + phpa*(4.5e-6+6e-10*tc*tc))
	pw := rh * ps / (1.0 - (1.0-rh)*ps/phpa)
	tk := tc + 273.15
	gamma := (77.6890e-6*phpa - (6.3938e-6-0.375463/tk)*pw) / tk
	beta := 4.4474e-6 * tk
	beta += -0.0074 * pw * beta
	a := gamma * (1.0 - beta)
	b := -gamma * (beta - gamma/2.0)
	return Refraction{a, b}, nil
}

func (ref Refraction) ObsEl2SkyEl(obsEl float64) float64 {
	// A*tan(z)+B*tan^3(z) model
	z := deg2rad(90 - obsEl)
	tz := math.Tan(z)
	dz := (ref.a + ref.b*tz*tz) * tz
	return obsEl - rad2deg(dz)
}

func (ref Refraction) SkyEl2ObsEl(skyEl float64) float64 {
	// A*tan(z)+B*tan^3(z) model, with Newton-Raphson correction
	z := deg2rad(90 - skyEl)
	sz := math.Sin(z)
	cz := math.Cos(z)
	tz := sz / cz
	w := ref.b * tz * tz
	dz := (ref.a + w) * tz / (1.0 + (ref.a+3.0*w)/(cz*cz))
	return skyEl + rad2deg(dz)
}
