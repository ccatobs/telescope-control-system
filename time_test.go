package main

import (
	"testing"
	"time"
)

func TestSeconds2Duration(t *testing.T) {
	got := Seconds2Duration(0.123)
	expected := 123 * time.Millisecond
	if got != expected {
		t.Errorf("Seconds2Duration: got %v, expected %v", got, expected)
	}
}

func TestConversions(t *testing.T) {
	t0 := time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC)
	ut := 1234567890.0
	{
		got := Time2Unixtime(t0)
		if got != ut {
			t.Errorf("Time2Unixtime: got %v, expected %v", got, ut)
		}
	}
	{
		got := Unixtime2Time(ut)
		if got != t0 {
			t.Errorf("Unixtime2Time: got %v, expected %v", got, t0)
		}
	}
	{
		doy, tod := programTrackTime(t0)
		expectedDoy := int32(31 + 13)
		expectedTod := 60*(60*23+31) + 30.0
		if doy != expectedDoy {
			t.Errorf("programTrackTime: got %v, expected %v", doy, expectedDoy)
		}
		if tod != expectedTod {
			t.Errorf("programTrackTime: got %v, expected %v", tod, expectedTod)
		}
	}
	{
		year, doy, tod := acuTimeParts(t0)
		if year != 2009 || doy != 31+13 || tod != 60*(60*23+31)+30.0 {
			t.Errorf("acuTimeParts: got (%v, %v, %v)", year, doy, tod)
		}
		got := acuTime(year, doy, tod)
		if got != t0 {
			t.Errorf("acuTime: got %v, expected %v", got, t0)
		}
	}
}
