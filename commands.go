package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

const (
	positionTol              = 1e-4
	speedTol                 = 1e-4
	maxFreeProgramTrackStack = 10000

	azimuthMin      = -180.0
	azimuthMax      = 360.0
	azimuthSpeedMax = 3.0  // [deg/sec]
	azimuthAccelMax = 6.0  // [deg/sec^2]
	azimuthJerkMax  = 12.0 // [deg/sec^3]

	elevationMin      = -90.0
	elevationMax      = 180.0
	elevationSpeedMax = 1.5 // [deg/sec]
	elevationAccelMax = 1.5 // [deg/sec^2]
	elevationJerkMax  = 6.0 // [deg/sec^3]
)

func checkAzEl(az, el, vaz, vel float64) error {
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
	if math.Abs(vaz) > azimuthSpeedMax {
		error := fmt.Sprintf("commanded azimuth vel (%g) out of range [%g,%g]", vaz, -azimuthSpeedMax, azimuthSpeedMax)
		log.Print(error)
		return fmt.Errorf(error)
	}
	if math.Abs(vel) > elevationSpeedMax {
		error := fmt.Sprintf("commanded elevation vel (%g) out of range [%g,%g]", vel, -elevationSpeedMax, elevationSpeedMax)
		log.Print(error)
		return fmt.Errorf(error)
	}
	return nil
}

type IsDoneFunc func(*Telescope) (bool, error)

type Command interface {
	Check() error
	Start(context.Context, *Telescope) (IsDoneFunc, error)
}

// JSON times are float64 unixtime in seconds,
// except that small values are relative to now.
func jsontime(x float64) time.Time {
	if x < 100000 {
		x += Time2Unixtime(time.Now())
	}
	return Unixtime2Time(x)
}

/*
 */

type enablePositionBroadcastCmd struct {
	Host string `json:"destination_host"`
	Port int    `json:"destination_port"`
}

func (cmd enablePositionBroadcastCmd) Check() error {
	if cmd.Port < 1024 || cmd.Port > 65535 {
		return fmt.Errorf("invalid port number %d", cmd.Port)
	}
	return nil
}

func (cmd enablePositionBroadcastCmd) Start(ctx context.Context, t *Telescope) (IsDoneFunc, error) {
	err := t.EnablePositionBroadcast(cmd.Host, cmd.Port)
	return func(*Telescope) (bool, error) { return true, nil }, err
}

/*
 */

type moveToCmd struct {
	Azimuth   float64
	Elevation float64
}

func (cmd moveToCmd) Check() error {
	return checkAzEl(cmd.Azimuth, cmd.Elevation, 0, 0)
}

func estimateMoveTime(az0, az1, el0, el1 float64) time.Duration {
	s, c := math.Sincos((az1 - az0) * math.Pi / 180)
	daz := math.Abs(math.Atan2(s, c) * 180 / math.Pi)
	daz = math.Max(daz, 360-daz) // might have to take the long way around (XXX:refine this estimate)
	taz := daz/azimuthSpeedMax + azimuthSpeedMax/azimuthAccelMax + azimuthAccelMax/azimuthJerkMax
	tel := math.Abs(el1-el0)/elevationSpeedMax + elevationSpeedMax/elevationAccelMax + elevationAccelMax/elevationJerkMax
	return Seconds2Duration(1.1 * math.Max(taz, tel))
}

func (cmd moveToCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	t0 := time.Now()
	rec := tel.Status()
	timeout := estimateMoveTime(cmd.Azimuth, rec.AzimuthCurrentPosition, cmd.Elevation, rec.ElevationCurrentPosition)
	log.Printf("estimated move time: %g secs", timeout.Seconds())
	err := tel.MoveTo(cmd.Azimuth, cmd.Elevation)
	isDone := func(tel *Telescope) (bool, error) {
		rec := tel.Status()
		done := (rec.AzimuthMode == datasets.AzimuthModePreset) &&
			(rec.ElevationMode == datasets.ElevationModePreset) &&
			(math.Abs(rec.AzimuthCurrentPosition-rec.AzimuthCommandedPosition) < positionTol) &&
			(math.Abs(rec.ElevationCurrentPosition-rec.ElevationCommandedPosition) < positionTol) &&
			(math.Abs(rec.AzimuthCurrentVelocity) < speedTol) &&
			(math.Abs(rec.ElevationCurrentVelocity) < speedTol)
		if !done && time.Since(t0) > timeout {
			return false, fmt.Errorf("move command timed out")
		}
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
	StartTime      float64    `json:"start_time"`
	TurnaroundTime float64    `json:"turnaround_time"`
	Speed          float64    `json:"speed"`
}

func (cmd azScanCmd) Check() error {
	// XXX:TBD
	return nil
}

func startPattern(ctx context.Context, tel *Telescope, pattern ScanPattern) (IsDoneFunc, error) {
	// ICD Section 9.1: "Before commanding or setting up a new mode,
	// it is best practice to set the antenna to Stop mode first."
	err := tel.acu.ModeSet("Stop")
	if err != nil {
		return nil, err
	}

	err = tel.acu.ProgramTrackClear()
	if err != nil {
		return nil, err
	}
	time.Sleep(3 * time.Millisecond) // wait for ProgramTrackClear to take effect

	uploadErrChan := make(chan error)
	go func() {
		uploadErrChan <- tel.UploadScanPattern(ctx, pattern)
	}()

	isDone := func(tel *Telescope) (bool, error) {
		// check for upload errors
		select {
		default:
		case err := <-uploadErrChan:
			if err != nil {
				return true, err
			}
		}

		// XXX:racy
		rec := tel.Status()
		done := (rec.QtyOfFreeProgramTrackStackPositions == maxFreeProgramTrackStack-1) && // last point remains on the stack
			(math.Abs(rec.AzimuthCurrentVelocity) < speedTol) &&
			(math.Abs(rec.ElevationCurrentVelocity) < speedTol) &&
			(rec.AzimuthMode == datasets.AzimuthModeProgramTrack) &&
			(rec.ElevationMode == datasets.ElevationModeProgramTrack)
		return done, nil
	}
	return isDone, tel.acu.ModeSet("ProgramTrack")
}

func (cmd azScanCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	t0 := jsontime(cmd.StartTime)
	pattern := NewAzimuthScanPattern(t0, cmd.NumScans, cmd.Elevation, cmd.AzimuthRange, cmd.Speed, Seconds2Duration(cmd.TurnaroundTime))
	return startPattern(ctx, tel, pattern)
}

type trackCmd struct {
	StartTime float64 `json:"start_time"`
	StopTime  float64 `json:"stop_time"`
	RA        float64
	Dec       float64
	Coordsys  string
}

func (cmd trackCmd) Check() error {
	switch cmd.Coordsys {
	case "Horizon":
	case "ICRS":
	default:
		return fmt.Errorf("bad coordinate system: %s", cmd.Coordsys)
	}
	if cmd.StopTime < cmd.StartTime {
		return fmt.Errorf("bad times: start=%f, stop=%f", cmd.StartTime, cmd.StopTime)
	}
	return nil
}

func (cmd trackCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	pattern, err := NewTrackScanPattern(jsontime(cmd.StartTime), jsontime(cmd.StopTime), cmd.RA, cmd.Dec, cmd.Coordsys)
	if err != nil {
		return nil, err
	}
	return startPattern(ctx, tel, pattern)
}

type pathCmd struct {
	Coordsys  string
	Points    [][5]float64
	StartTime float64 `json:"start_time"`
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

	// check the times
	for i := 1; i < len(cmd.Points); i++ {
		// ACU ICD 2.0, section 8.9.3:
		// "The minimum time interval between two samples is 0.05 s."
		if cmd.Points[i][0]-cmd.Points[i-1][0] < 0.05 {
			return fmt.Errorf("points are separated by less than 50 ms")
		}
	}

	// check the first 100 coordinates
	pattern := NewPathScanPattern(jsontime(cmd.StartTime), cmd.Points, cmd.Coordsys)
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
		err = checkAzEl(pt.AzPosition, pt.ElPosition, pt.AzVelocity, pt.ElVelocity)
		if err != nil {
			return fmt.Errorf("point %d: %w", i, err)
		}
	}

	return nil
}

func (cmd pathCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	pattern := NewPathScanPattern(jsontime(cmd.StartTime), cmd.Points, cmd.Coordsys)
	return startPattern(ctx, tel, pattern)
}
