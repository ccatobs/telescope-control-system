package main

import (
	"math"
	"time"
)

func DaySeconds(t time.Time) float64 {
	h, m, s := t.Clock()
	return float64(60*60*h+60*m+s) + 1e-9*float64(t.Nanosecond())
}

func Unixtime2Time(unixtime float64) time.Time {
	a, b := math.Modf(unixtime)
	s := int64(a)
	ns := int64(1e9 * b)
	return time.Unix(s, ns).UTC()
}
