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
