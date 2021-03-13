package main

import (
	"math"
	"time"
)

func VertexTime(t time.Time) (int32, float64) {
	utc := t.UTC()
	doy := utc.YearDay()
	h, m, s := utc.Clock()
	ns := utc.Nanosecond()
	return int32(doy), float64(60*(60*h+m)+s) + float64(ns)*1e-9
}

func Unixtime2Time(unixtime float64) time.Time {
	a, b := math.Modf(unixtime)
	s := int64(a)
	ns := int64(1e9 * b)
	return time.Unix(s, ns).UTC()
}

func Time2Unixtime(t time.Time) float64 {
	now := t.UnixNano()
	s := now / 1e9
	ns := now % 1e9
	return float64(s) + float64(ns)*1e-9
}

// Convert float64 seconds to a time.Duration.
func Seconds2Duration(s float64) time.Duration {
	return time.Duration(s*1e9) * time.Nanosecond
}
