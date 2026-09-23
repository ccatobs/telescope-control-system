package main

import (
	"math"
	"time"
)

func programTrackTime(t time.Time) (int32, float64) {
	_, doy, tod := acuTimeParts(t)
	return doy, tod
}

// acuTime converts a (year, day-of-year, seconds-of-day) ACU time to a time.Time.
// Day-of-year is 1-based, i.e., January 1 is day 1.
func acuTime(year int, doy int32, tod float64) time.Time {
	return time.Date(year, time.January, int(doy), 0, 0, 0, 0, time.UTC).Add(Seconds2Duration(tod))
}

// acuTimeParts is the inverse of acuTime.
func acuTimeParts(t time.Time) (int, int32, float64) {
	utc := t.UTC()
	doy := utc.YearDay()
	h, m, s := utc.Clock()
	ns := utc.Nanosecond()
	return utc.Year(), int32(doy), float64(60*(60*h+m)+s) + float64(ns)*1e-9
}

func Unixtime2Time(unixtime float64) time.Time {
	a, b := math.Modf(unixtime)
	s := int64(a)
	ns := int64(1e9 * b)
	return time.Unix(s, ns).UTC()
}

func Time2Unixtime(t time.Time) float64 {
	return float64(t.UnixNano()) * 1e-9
}

// Convert float64 seconds to a time.Duration.
func Seconds2Duration(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}
