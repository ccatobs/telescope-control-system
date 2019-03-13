package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ccatp/antenna-control-unit/datasets"
)

const (
	positionTol              = 1e-4
	speedTol                 = 1e-4
	maxFreeProgramTrackStack = 10000
	azimuthMin               = 90.0
	azimuthMax               = 190.0
	elevationMin             = 0.0
	elevationMax             = 90.0
)

var (
	errAzimuthOutOfRange   = fmt.Errorf("azimuth out of range [%g,%g]", azimuthMin, azimuthMax)
	errElevationOutOfRange = fmt.Errorf("elevation out of range [%g,%g]", elevationMin, elevationMax)
)

type IsDoneFunc func(*datasets.MonitoringRecord) (bool, error)

type Command interface {
	Check() error
	Start(context.Context, *Telescope) (IsDoneFunc, error)
}

/*
 */

type moveToCmd struct {
	Azimuth   float64
	Elevation float64
}

func (cmd moveToCmd) Check() error {
	if cmd.Azimuth < azimuthMin || cmd.Azimuth > azimuthMax {
		return errAzimuthOutOfRange
	}
	if cmd.Elevation < elevationMin || cmd.Elevation > elevationMax {
		return errElevationOutOfRange
	}
	return nil
}

func (cmd moveToCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	err := tel.MoveTo(cmd.Azimuth, cmd.Elevation)
	isDone := func(rec *datasets.MonitoringRecord) (bool, error) {
		done := (rec.AzimuthMode == datasets.AzimuthModePreset) &&
			(rec.ElevationMode == datasets.ElevationModePreset) &&
			(math.Abs(rec.AzimuthCurrentPosition-rec.AzimuthDesiredPosition) < positionTol) &&
			(math.Abs(rec.ElevationCurrentPosition-rec.ElevationDesiredPosition) < positionTol) &&
			(math.Abs(rec.AzimuthCurrentVelocity) < speedTol) &&
			(math.Abs(rec.ElevationCurrentVelocity) < speedTol)
		return done, nil
	}
	return isDone, err
}

/*
 */

type azScanCmd struct {
	AzimuthRange   [2]float64 `json:"azimuth_range"`
	Elevation      float64    `json:"elevation"`
	NumScans       int        `json:"num_scans"`
	StartTime      time.Time  `json:"start_time"`
	TurnaroundTime float64    `json:"turnaround_time"`
	Speed          float64    `json:"speed"`
}

func (cmd azScanCmd) Check() error {
	// XXX:TBD
	return nil
}

func (cmd azScanCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	pattern := AzimuthScanPattern(cmd.StartTime, cmd.NumScans, cmd.Elevation, cmd.AzimuthRange, cmd.Speed, time.Duration(cmd.TurnaroundTime*1e9)*time.Nanosecond)
	go func() {
		err := tel.UploadScanPattern(ctx, pattern)
		if err != nil {
			log.Print(err)
		}
	}()
	isDone := func(rec *datasets.MonitoringRecord) (bool, error) {
		// XXX:racy
		done := rec.FreeProgramTrackStack == maxFreeProgramTrackStack-1 // last point remains on the stack
		return done, nil
	}
	return isDone, tel.acu.ModeSet("ProgramTrack")
}