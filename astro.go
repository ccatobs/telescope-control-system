package main

// #include "erfa.h"
import "C"

import (
	"fmt"
	"math"
)

const (
	// telescope coordinates
	CCATP_ELEVATION_METERS   = 5611.8
	CCATP_LATITUDE_DEG       = -22.985639 // -22°59'08.30"
	CCATP_LATITUDE_RAD       = CCATP_LATITUDE_DEG * math.Pi / 180
	CCATP_LONGITUDE_EAST_DEG = -67.740278 // -67°44'25.00"
	CCATP_LONGITUDE_EAST_RAD = CCATP_LONGITUDE_EAST_DEG * math.Pi / 180

	// physical constants
	SPEED_OF_LIGHT_METERS_PER_SEC = 299792458.0

	// Julian date for unixtime = 0
	UNIX_JD_EPOCH = 2440587.5
)

func deg2rad(d float64) float64 {
	return d * math.Pi / 180
}

func rad2deg(r float64) float64 {
	return r * 180 / math.Pi
}

func AzEl2RADec(unixtime, az, el float64) (float64, float64, error) {
	utc1 := UNIX_JD_EPOCH
	utc2 := unixtime / 86400
	dut1 := 0.0
	xp := 0.0
	yp := 0.0

	freq_GHz := 100.0
	wavelength_um := (SPEED_OF_LIGHT_METERS_PER_SEC * 1e-3) / freq_GHz
	pressure_mbar := 500.0
	temperature_C := 0.0
	relative_humidity := 0.2

	var ra, dec C.double

	// Observed place at a groundbased site to to ICRS astrometric RA,Dec.
	stat := C.eraAtoc13(
		C.CString("A"),              // ob1 and ob2 are azimuth and zenith distance
		C.double(deg2rad(az)),       // observed Az (radians; Az is N=0,E=90)
		C.double(deg2rad(90-el)),    // observed ZD (radians)
		C.double(utc1),              // UTC as a 2-part...
		C.double(utc2),              // ...quasi Julian Date (Notes 3,4)
		C.double(dut1),              // UT1-UTC (seconds, Note 5)
		CCATP_LONGITUDE_EAST_RAD,    // longitude (radians, east +ve, Note 6)
		CCATP_LATITUDE_RAD,          // geodetic latitude (radians, Note 6)
		CCATP_ELEVATION_METERS,      // height above ellipsoid (m, geodetic Notes 6,8)
		C.double(xp),                // polar motion coordinates (radians, Note 7)
		C.double(yp),                // polar motion coordinates (radians, Note 7)
		C.double(pressure_mbar),     // pressure at the observer (hPa = mB, Note 8)
		C.double(temperature_C),     // ambient temperature at the observer (deg C)
		C.double(relative_humidity), // relative humidity at the observer (range 0-1)
		C.double(wavelength_um),     // wavelength (micrometers, Note 9)
		&ra,                         // ICRS astrometric RA (radians)
		&dec)                        // ICRS astrometric Dec (radians)

	var err error
	switch stat {
	case 0:
		// ok
	case 1:
		err = fmt.Errorf("eraAtoc13: dubious year")
	case -1:
		err = fmt.Errorf("eraAtoc13: unacceptable date")
	default:
		err = fmt.Errorf("eraAtoc13: unknow error")
	}

	return rad2deg(float64(ra)), rad2deg(float64(dec)), err
}

func RADec2AzEl(unixtime, ra, dec float64) (float64, float64, error) {
	utc1 := UNIX_JD_EPOCH
	utc2 := unixtime / 86400
	dut1 := 0.0
	xp := 0.0
	yp := 0.0

	freq_GHz := 100.0
	wavelength_um := (SPEED_OF_LIGHT_METERS_PER_SEC * 1e-3) / freq_GHz
	pressure_mbar := 500.0
	temperature_C := 0.0
	relative_humidity := 0.2

	var aob, zob, hob, dob, rob, eo C.double

	// ICRS RA,Dec to observed place.
	stat := C.eraAtco13(
		C.double(deg2rad(ra)),       // ICRS right ascension at J2000.0 (radians, Note 1)
		C.double(deg2rad(dec)),      // ICRS declination at J2000.0 (radians, Note 1)
		0,                           // RA proper motion (radians/year; Note 2)
		0,                           // Dec proper motion (radians/year)
		0,                           // parallax (arcsec)
		0,                           // radial velocity (km/s, +ve if receding)
		C.double(utc1),              // UTC as a 2-part...
		C.double(utc2),              // ...quasi Julian Date (Notes 3-4)
		C.double(dut1),              // UT1-UTC (seconds, Note 5)
		CCATP_LONGITUDE_EAST_RAD,    // longitude (radians, east +ve, Note 6)
		CCATP_LATITUDE_RAD,          // latitude (geodetic, radians, Note 6)
		CCATP_ELEVATION_METERS,      // height above ellipsoid (m, geodetic, Notes 6,8)
		C.double(xp),                // polar motion coordinates (radians, Note 7)
		C.double(yp),                // polar motion coordinates (radians, Note 7)
		C.double(pressure_mbar),     // pressure at the observer (hPa = mB, Note 8)
		C.double(temperature_C),     // ambient temperature at the observer (deg C)
		C.double(relative_humidity), // relative humidity at the observer (range 0-1)
		C.double(wavelength_um),     // wavelength (micrometers, Note 9)
		&aob,                        // observed azimuth (radians: N=0,E=90)
		&zob,                        // observed zenith distance (radians)
		&hob,                        // observed hour angle (radians)
		&dob,                        // observed declination (radians)
		&rob,                        // observed right ascension (CIO-based, radians)
		&eo)                         // equation of the origins (ERA-GST)

	var err error
	switch stat {
	case 0:
		// ok
	case 1:
		err = fmt.Errorf("eraAtco13: dubious year")
	case -1:
		err = fmt.Errorf("eraAtco13: unacceptable date")
	default:
		err = fmt.Errorf("eraAtco13: unknow error")
	}

	return rad2deg(float64(aob)), 90 - rad2deg(float64(zob)), err
}
