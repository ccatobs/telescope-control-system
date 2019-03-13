package main

import (
	"time"
)

func DaySeconds(t time.Time) float64 {
	h, m, s := t.Clock()
	return float64(60*60*h+60*m+s) + 1e-9*float64(t.Nanosecond())
}
