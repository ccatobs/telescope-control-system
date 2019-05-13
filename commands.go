package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/ccatp/antenna-control-unit/datasets"
)

const (
	positionTol              = 1e-4
	speedTol                 = 1e-4
	maxFreeProgramTrackStack = 10000
	azimuthMin               = 1.0
	azimuthMax               = 358.8
	elevationMin             = 5.0
	elevationMax             = 90.0
)

var (
	errAzimuthOutOfRange   = fmt.Errorf("azimuth out of range [%g,%g]", azimuthMin, azimuthMax)
	errElevationOutOfRange = fmt.Errorf("elevation out of range [%g,%g]", elevationMin, elevationMax)
)

func checkAzEl(az, el float64) error {
	if az < azimuthMin || az > azimuthMax {
                error := fmt.Sprintf("commanded azimuth (%g) out of range [%g,%g]", az, azimuthMin, azimuthMax)
                log.Print(error)
		return fmt.Errorf(error)
	}
	if el < elevationMin || el > elevationMax {
                error := fmt.Sprintf("commanded elevation (%g) out of range [%g,%g]", el, elevationMin, elevationMax)
                log.Print(error)
		return fmt.Errorf(error)
	}
	return nil
}

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
	return checkAzEl(cmd.Azimuth, cmd.Elevation)
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

func startPattern(ctx context.Context, tel *Telescope, pattern ScanPattern) (IsDoneFunc, error) {
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

func (cmd azScanCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	pattern := NewAzimuthScanPattern(cmd.StartTime, cmd.NumScans, cmd.Elevation, cmd.AzimuthRange, cmd.Speed, time.Duration(cmd.TurnaroundTime*1e9)*time.Nanosecond)
	return startPattern(ctx, tel, pattern)
}

type trackCmd struct {
	StartTime time.Time `json:"start_time"`
	StopTime  time.Time `json:"stop_time"`
	RA        float64
	Dec       float64
}

func (cmd trackCmd) Check() error {
	if !cmd.StartTime.Before(cmd.StopTime) {
		return fmt.Errorf("bad times: start=%v, stop=%v", cmd.StartTime, cmd.StopTime)
	}
	return nil
}

func (cmd trackCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	pattern, err := NewTrackScanPattern(cmd.StartTime, cmd.StopTime, cmd.RA, cmd.Dec)
	if err != nil {
		return nil, err
	}
	return startPattern(ctx, tel, pattern)
}

type pathCmd struct {
	Coordsys string
	Points   [][3]float64
}

func (cmd pathCmd) Check() error {
	switch cmd.Coordsys {
	case "Horizon":
	case "ICRS":
	default:
		return fmt.Errorf("bad coordinate system: %s", cmd.Coordsys)
	}

	if len(cmd.Points) == 0 {
		return fmt.Errorf("no points in path")
	}

	// sanity check the first 100 points
	pattern := NewPathScanPattern(cmd.Coordsys, cmd.Points)
	iter := pattern.Iterator()
	for i := 0; i < 100; i++ {
		if pattern.Done(iter) {
			break
		}
		var pt datasets.TimePositionTransfer
		err := pattern.Next(iter, &pt)
		if err != nil {
			return err
		}
		err = checkAzEl(pt.AzPosition, pt.ElPosition)
		if err != nil {
			return errors.Wrapf(err, "point %d", i)
		}
	}

	return nil
}

func (cmd pathCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	pattern := NewPathScanPattern(cmd.Coordsys, cmd.Points)
	return startPattern(ctx, tel, pattern)
}
