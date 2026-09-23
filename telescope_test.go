package main

import (
	"testing"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

// statusRec builds an ACU status record for time t, as the ACU reports it.
func statusRec(t time.Time) datasets.StatusGeneral8100 {
	year, doy, tod := acuTimeParts(t)
	return datasets.StatusGeneral8100{
		Year: uint32(year),
		Time: float64(doy) + tod/(24*60*60),
	}
}

func TestStatusTime(t *testing.T) {
	t0 := time.Date(2026, 12, 31, 23, 59, 59, 500_000_000, time.UTC)
	rec := statusRec(t0)
	got := statusTime(&rec)
	if d := got.Sub(t0).Abs(); d > time.Microsecond {
		t.Errorf("statusTime: got %v, expected %v", got, t0)
	}
}

func TestCalibrateTime(t *testing.T) {
	for _, offset := range []time.Duration{-3 * time.Second, 0, 3 * time.Second} {
		tel := NewTelescope(nil)
		tel.rec = statusRec(time.Now().Add(offset))
		if err := tel.CalibrateTime(); err != nil {
			t.Fatalf("CalibrateTime(%v): %v", offset, err)
		}
		if d := (tel.pointing.tOffset - offset).Abs(); d > 100*time.Millisecond {
			t.Errorf("CalibrateTime(%v): got offset %v", offset, tel.pointing.tOffset)
		}
	}
}

func TestCalibrateTimeNoStatus(t *testing.T) {
	tel := NewTelescope(nil)
	if err := tel.CalibrateTime(); err == nil {
		t.Error("CalibrateTime with no ACU status: expected error")
	}
}
