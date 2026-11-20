package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

const (
	positionTol              = 1e-4
	speedTol                 = 1e-4
	maxFreeProgramTrackStack = 10000

	// max time for stow, unstow, & survival mode to complete
	modeChangeTimeout = 5 * time.Minute

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

// spemParams are the SPEM parameter names, in ACU dataset order.
// ACU Extension ICD v1.4, section 2.5.
var spemParams = []string{
	"IA", "IE", "TFC", "TF", "TFS", "TFEC", "TFE", "TFES",
	"AN", "AW", "AN2", "AW2", "CA", "NPAE",
	"AES", "AEC", "AES2", "AEC2", "EES", "EEC", "EES2", "EEC2",
	"NRX", "NRY", "LELA", "LEVA", "LELE", "LEVE",
	"LEL_velocity",
}

// setSPEMCmd sets a subset of the SPEM parameters.
type setSPEMCmd map[string]float64

func (cmd setSPEMCmd) Check() error {
	if len(cmd) == 0 {
		return fmt.Errorf("no SPEM parameters")
	}
	for name, value := range cmd {
		if !slices.Contains(spemParams, name) {
			return fmt.Errorf("unknown SPEM parameter: %s", name)
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("SPEM parameter %s is not finite", name)
		}
	}
	return nil
}

func (cmd setSPEMCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	for _, name := range spemParams {
		value, ok := cmd[name]
		if !ok {
			continue
		}
		err := tel.acu.SPEMParameterSet(name, value)
		if err != nil {
			return nil, fmt.Errorf("setting SPEM parameter %s: %w", name, err)
		}
	}
	return func(*Telescope) (bool, error) { return true, nil }, nil
}

/*
 */

// startMode switches the telescope to mode, and is done once
// atTarget is true and both axes have stopped moving.
// The axes must remain in one of the allowed modes
// (azimuth and elevation share the same mode codes).
func startMode(tel *Telescope, mode string, allowed []uint8, atTarget func(*datasets.StatusGeneral8100) bool) (IsDoneFunc, error) {
	t0 := time.Now()
	err := tel.SetMode(mode)
	isDone := func(tel *Telescope) (bool, error) {
		rec := tel.Status()
		if !slices.Contains(allowed, rec.AzimuthMode) {
			return true, fmt.Errorf("azimuth not in %s mode: %d", mode, rec.AzimuthMode)
		}
		if !slices.Contains(allowed, rec.ElevationMode) {
			return true, fmt.Errorf("elevation not in %s mode: %d", mode, rec.ElevationMode)
		}
		done := atTarget(rec) &&
			(math.Abs(rec.AzimuthCurrentVelocity) < speedTol) &&
			(math.Abs(rec.ElevationCurrentVelocity) < speedTol)
		if !done && time.Since(t0) > modeChangeTimeout {
			return false, fmt.Errorf("%s timed out", mode)
		}
		return done, nil
	}
	return isDone, err
}

// onCommandedPosition reports whether both axes have reached their commanded positions.
func onCommandedPosition(rec *datasets.StatusGeneral8100) bool {
	return (math.Abs(rec.AzimuthCurrentPosition-rec.AzimuthCommandedPosition) < positionTol) &&
		(math.Abs(rec.ElevationCurrentPosition-rec.ElevationCommandedPosition) < positionTol)
}

// stowPinsInserted reports whether the elevation stow pins are inserted.
func stowPinsInserted(rec *datasets.StatusGeneral8100) bool {
	return rec.ElevationStowPinsStatus == datasets.ElevationStowPinsStatusAllInserted ||
		rec.ElevationStowPinsStatus == datasets.ElevationStowPinsStatusNotAvailable
}

// stowCmd inserts the stow pins. It does not move the telescope.
type stowCmd struct{}

func (cmd stowCmd) Check() error { return nil }

func (cmd stowCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	return startMode(tel, "Stow", []uint8{datasets.AzimuthModeStow}, func(rec *datasets.StatusGeneral8100) bool {
		return rec.ElevationAxisInStowPosition && stowPinsInserted(rec)
	})
}

// unstowCmd retracts the stow pins.
type unstowCmd struct{}

func (cmd unstowCmd) Check() error { return nil }

func (cmd unstowCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	// the ACU may return to Stop mode once the pins are out
	allowed := []uint8{datasets.AzimuthModeUnStow, datasets.AzimuthModeStop}
	return startMode(tel, "UnStow", allowed, func(rec *datasets.StatusGeneral8100) bool {
		return rec.ElevationStowPinsStatus == datasets.ElevationStowPinsStatusAllRetracted ||
			rec.ElevationStowPinsStatus == datasets.ElevationStowPinsStatusNotAvailable
	})
}

// survivalModeCmd moves to the survival position & inserts the stow pins.
type survivalModeCmd struct{}

func (cmd survivalModeCmd) Check() error { return nil }

func (cmd survivalModeCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	return startMode(tel, "SurvivalMode", []uint8{datasets.AzimuthModeSurvivalMode}, func(rec *datasets.StatusGeneral8100) bool {
		return onCommandedPosition(rec) && stowPinsInserted(rec)
	})
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
		done := onCommandedPosition(rec) &&
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
	if err := checkScanStartTime(cmd.StartTime); err != nil {
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

	// clear any latched failures, e.g. from a previous program track
	err = tel.acu.FailureReset()
	if err != nil {
		return nil, err
	}

	// a previous upload must not add points after we clear the stack
	tel.WaitForUploads()

	err = tel.ClearProgramTrack()
	if err != nil {
		return nil, err
	}
	time.Sleep(3 * time.Millisecond) // wait for ProgramTrackClear to take effect

	// buffered, so the upload goroutine can exit if nobody's listening
	uploadErrChan := make(chan error, 1)
	firstBatch := make(chan struct{})
	tel.uploads.Go(func() {
		uploadErrChan <- tel.UploadScanPattern(ctx, pattern, firstBatch)
	})

	// the ACU won't enter program track mode with an empty stack
	select {
	case <-firstBatch:
	case err := <-uploadErrChan:
		// nil means the whole pattern was uploaded in one batch
		if err != nil {
			return nil, err
		}
	}

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
	return moveThenStartPattern(ctx, tel, cmd.AzimuthRange[0], cmd.Elevation, cmd.StartTime, func(t0 time.Time) (ScanPattern, error) {
		return NewAzimuthScanPattern(t0, cmd.NumScans, cmd.Elevation, cmd.AzimuthRange, cmd.Speed, Seconds2Duration(cmd.TurnaroundTime)), nil
	})
}

// Scan start times are optional; 0 means start as soon as possible.
func checkScanStartTime(startTime float64) error {
	if startTime == 0 {
		return nil
	}
	return checkProgramTrackStartTime(startTime)
}

// moveThenStartPattern moves to (az,el), the start of the pattern, waits there
// until shortly before startTime, then starts the pattern made by newPattern.
// If startTime is 0, the pattern starts as soon as possible after the move.
func moveThenStartPattern(ctx context.Context, tel *Telescope, az, el float64, startTime float64, newPattern func(t0 time.Time) (ScanPattern, error)) (IsDoneFunc, error) {
	var t0 time.Time
	if startTime != 0 {
		t0 = jsontime(startTime)
	}
	moveDone, err := moveToCmd{Azimuth: az, Elevation: el}.Start(ctx, tel)
	if err != nil {
		return nil, err
	}
	var patternDone IsDoneFunc
	isDone := func(tel *Telescope) (bool, error) {
		if patternDone != nil {
			return patternDone(tel)
		}
		done, err := moveDone(tel)
		if err != nil {
			return true, err
		}
		if !done {
			return false, nil
		}
		// ACU Extension ICD v1.4, section 3:
		// "mode activation should precede first track time by 5 s"
		if t0.IsZero() {
			t0 = time.Now().Add(10 * time.Second)
		}
		lead := time.Until(t0)
		if lead < 5*time.Second {
			return true, fmt.Errorf("reached start of scan too late to start at %v", t0)
		}
		// wait until shortly before the scan
		if lead > 10*time.Second {
			return false, nil
		}
		slog.Info("at start of scan, starting scan", "start_time", t0)
		pattern, err := newPattern(t0)
		if err != nil {
			return true, err
		}
		patternDone, err = startPattern(ctx, tel, pattern)
		return err != nil, err
	}
	return isDone, nil
}

type holographyScanCmd struct {
	Azimuth        float64 `json:"azimuth"`
	Elevation      float64 `json:"elevation"`
	Radius         float64 `json:"radius"`
	NumLines       int     `json:"num_lines"`
	Speed          float64 `json:"speed"`
	StartTime      float64 `json:"start_time"`
	TurnaroundTime float64 `json:"turnaround_time"`
}

func (cmd holographyScanCmd) Check() error {
	if err := checkScanStartTime(cmd.StartTime); err != nil {
		return err
	}
	if cmd.NumLines <= 0 || cmd.NumLines%2 != 0 {
		return fmt.Errorf("num_lines (%d) must be positive and even", cmd.NumLines)
	}
	if !(cmd.Radius > 0) {
		return fmt.Errorf("radius (%g) must be positive", cmd.Radius)
	}
	if !(cmd.Speed > 0) {
		return fmt.Errorf("speed (%g) must be positive", cmd.Speed)
	}
	// "The minimum time interval between two samples is 0.05 s."
	if !(cmd.TurnaroundTime >= 0.05) {
		return fmt.Errorf("turnaround_time (%g) must be at least 0.05 s", cmd.TurnaroundTime)
	}
	if 2*cmd.Radius/cmd.Speed < 4*0.05 {
		return fmt.Errorf("lines too short for speed: points would be < 0.05 s apart")
	}
	// check every point, including the lead-in & lead-out
	pattern, err := cmd.newPattern(time.Now())
	if err != nil {
		return err
	}
	iter := pattern.Iterator()
	for !pattern.Done(iter) {
		var pt ScanPatternSample
		if err := pattern.Next(iter, &pt); err != nil {
			return err
		}
		if err := checkAzEl(pt.Az, pt.El, pt.AzVel, pt.ElVel); err != nil {
			return err
		}
	}
	return nil
}

func (cmd holographyScanCmd) newPattern(t0 time.Time) (ScanPattern, error) {
	return NewHolographyScanPattern(t0, cmd.NumLines, cmd.Azimuth, cmd.Elevation, cmd.Radius, cmd.Speed, Seconds2Duration(cmd.TurnaroundTime))
}

func (cmd holographyScanCmd) Start(ctx context.Context, tel *Telescope) (IsDoneFunc, error) {
	// the pattern starts at rest at the center
	return moveThenStartPattern(ctx, tel, cmd.Azimuth, cmd.Elevation, cmd.StartTime, cmd.newPattern)
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
