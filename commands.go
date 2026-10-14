package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sync/atomic"
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

func checkCoordsys(s string) error {
	switch s {
	case "Horizon", "ICRS", "Galactic":
		return nil
	default:
		return fmt.Errorf("bad coordinate system: %s", s)
	}
}

func checkAzEl(az, el, vaz, vel float64) error {
	if az < azimuthMin || az > azimuthMax {
		return fmt.Errorf("commanded azimuth (%g) out of range [%g,%g]", az, azimuthMin, azimuthMax)
	}
	if el < elevationMin || el > elevationMax {
		return fmt.Errorf("commanded elevation (%g) out of range [%g,%g]", el, elevationMin, elevationMax)
	}
	if math.Abs(vaz) > azimuthSpeedMax {
		return fmt.Errorf("commanded azimuth vel (%g) out of range [%g,%g]", vaz, -azimuthSpeedMax, azimuthSpeedMax)
	}
	if math.Abs(vel) > elevationSpeedMax {
		return fmt.Errorf("commanded elevation vel (%g) out of range [%g,%g]", vel, -elevationSpeedMax, elevationSpeedMax)
	}
	return nil
}

// ACU Extension ICD v1.3, section 3:
// Program track should start at least 5 seconds in future;
// we'll pad it to ~10 seconds for now
func checkProgramTrackStartTime(startTime float64) error {
	if time.Until(jsontime(startTime)) < 9800*time.Millisecond {
		return fmt.Errorf("program track starts too soon (< 10 seconds from now)")
	}
	return nil
}

// activeUploads counts the scan pattern uploads still running in the background.
var activeUploads atomic.Int32

// uploadFailed is set when the upload of the running scan pattern failed.
var uploadFailed atomic.Bool

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
	Azimuth   float64 `json:"azimuth"`
	Elevation float64 `json:"elevation"`
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
	slog.Info("estimated move time", "timeout", timeout.Round(time.Millisecond))
	err := tel.MoveTo(cmd.Azimuth, cmd.Elevation)
	isDone := func(tel *Telescope) (bool, error) {
		rec := tel.Status()
		if rec.AzimuthMode != datasets.AzimuthModePreset {
			return true, fmt.Errorf("azimuth not in preset mode: %d", rec.AzimuthMode)
		}
		if rec.ElevationMode != datasets.ElevationModePreset {
			return true, fmt.Errorf("elevation not in preset mode: %d", rec.ElevationMode)
		}
		done := (math.Abs(rec.AzimuthCurrentPosition-rec.AzimuthCommandedPosition) < positionTol) &&
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
	if err := checkProgramTrackStartTime(cmd.StartTime); err != nil {
		return err
	}

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
	uploadFailed.Store(false)
	activeUploads.Add(1)
	go func() {
		err := tel.UploadScanPattern(ctx, pattern)
		if err != nil {
			uploadFailed.Store(true)
		}
		activeUploads.Add(-1)
		uploadErrChan <- err
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

		rec := tel.Status()

		if rec.AzimuthMode != datasets.AzimuthModeProgramTrack {
			return true, fmt.Errorf("azimuth not in program track mode: %d", rec.AzimuthMode)
		}
		if rec.ElevationMode != datasets.ElevationModeProgramTrack {
			return true, fmt.Errorf("elevation not in program track mode: %d", rec.ElevationMode)
		}

		// XXX:racy
		done := (rec.QtyOfFreeProgramTrackStackPositions == maxFreeProgramTrackStack-1) && // last point remains on the stack
			(math.Abs(rec.AzimuthCurrentVelocity) < speedTol) &&
			(math.Abs(rec.ElevationCurrentVelocity) < speedTol)
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
	RA        float64 `json:"ra"`
	Dec       float64 `json:"dec"`
	Coordsys  string  `json:"coordsys"`
}

func (cmd trackCmd) Check() error {
	if err := checkCoordsys(cmd.Coordsys); err != nil {
		return err
	}

	if err := checkProgramTrackStartTime(cmd.StartTime); err != nil {
		return err
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
	Coordsys  string       `json:"coordsys"`
	Points    [][5]float64 `json:"points"`
	StartTime float64      `json:"start_time"`
}

func (cmd pathCmd) Check() error {
	if err := checkCoordsys(cmd.Coordsys); err != nil {
		return err
	}

	if err := checkProgramTrackStartTime(cmd.StartTime); err != nil {
		return err
	}

	if len(cmd.Points) == 0 {
		return fmt.Errorf("no points in path")
	}

	// check the times
	for i := 1; i < len(cmd.Points); i++ {
		// ACU ICD v2.0, section 8.9.3:
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
		var pt ScanPatternSample
		err := pattern.Next(iter, &pt)
		if err != nil {
			return err
		}
		err = checkAzEl(pt.Az, pt.El, pt.AzVel, pt.ElVel)
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

// runningPath is the state of a running /path command that /path/append extends.
type runningPath struct {
	id       uint64
	t0       time.Time
	coordsys string
	lastT    float64 // time of the last queued point [seconds after t0]
}

// appendResult is the answer to a /path/append request.
type appendResult struct {
	id     uint64 // ID of the running path, or 0 if none runs
	status int    // HTTP status code
	err    error
}

// appendPath appends points to the program track of the running path p, or
// rejects them. p is nil when no path runs.
func appendPath(tel *Telescope, p *runningPath, points [][5]float64) appendResult {
	if p == nil {
		return appendResult{0, http.StatusConflict, fmt.Errorf("no path running")}
	}
	// the path ends at the next status check; the stack lacks some of its points
	if uploadFailed.Load() {
		return appendResult{p.id, http.StatusConflict, fmt.Errorf("path upload failed")}
	}
	if len(points) == 0 {
		return appendResult{p.id, http.StatusBadRequest, fmt.Errorf("no points to append")}
	}
	pattern := NewPathScanPattern(p.t0, points, p.coordsys)
	iter := pattern.Iterator()
	last := p.lastT
	for i, x := range points {
		// ACU ICD v2.0, section 8.9.3:
		// "The minimum time interval between two samples is 0.05 s."
		if x[0]-last < 0.05 {
			return appendResult{p.id, http.StatusBadRequest, fmt.Errorf("point %d: time %g is not at least 50 ms after the last queued point (%g)", i, x[0], last)}
		}
		last = x[0]

		var pt ScanPatternSample
		err := pattern.Next(iter, &pt)
		if err == nil {
			err = checkAzEl(pt.Az, pt.El, pt.AzVel, pt.ElVel)
		}
		if err != nil {
			return appendResult{p.id, http.StatusBadRequest, fmt.Errorf("point %d: %w", i, err)}
		}
	}
	// appended points must follow the path's own points on the stack
	if activeUploads.Load() > 0 {
		return appendResult{p.id, http.StatusServiceUnavailable, fmt.Errorf("path upload in progress")}
	}

	// UploadScanPattern waits between batches when the points don't fit the
	// stack, so only append what fits in one batch.
	var status datasets.StatusGeneral8100
	if err := tel.acu.StatusGeneral8100Get(&status); err != nil {
		return appendResult{p.id, http.StatusInternalServerError, err}
	}
	if free := int(status.QtyOfFreeProgramTrackStackPositions); len(points) > free {
		return appendResult{p.id, http.StatusServiceUnavailable, fmt.Errorf("%d points don't fit the %d free ACU program track stack positions", len(points), free)}
	}
	if err := tel.UploadScanPattern(context.Background(), pattern); err != nil {
		return appendResult{p.id, http.StatusInternalServerError, err}
	}
	p.lastT = last
	return appendResult{p.id, http.StatusOK, nil}
}
