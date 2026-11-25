package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

// largest ACU - TCS clock difference CalibrateTime will correct for
const maxClockOffset = 24 * time.Hour

// Telescope provides a higher-level interface to the ACU.
// Responsible for pointing corrections and coordinate transformations.
type Telescope struct {
	acu      *ACU
	pointing Pointing
	// latest ACU status, replaced (never modified) by UpdateStatus,
	// so it can be read while the main loop updates it
	status  *atomic.Pointer[datasets.StatusGeneral8100]
	uploads *sync.WaitGroup // background program track uploads
	posted  *atomic.Bool    // program track points were sent to housekeeping

	// housekeeping URL program track points are posted to; empty disables posting
	trackUploadURL string
}

func NewTelescope(acu *ACU) *Telescope {
	return &Telescope{
		acu:      acu,
		pointing: NewPointing(),
		status:   new(atomic.Pointer[datasets.StatusGeneral8100]),
		uploads:  new(sync.WaitGroup),
		posted:   new(atomic.Bool),
	}
}

// WaitForUploads blocks until all background program track uploads have exited.
// Uploads only exit early when their context is cancelled.
func (t Telescope) WaitForUploads() {
	t.uploads.Wait()
}

func (t *Telescope) UpdateStatus() error {
	rec := new(datasets.StatusGeneral8100)
	err := t.acu.StatusGeneral8100Get(rec)
	if err != nil {
		*rec = datasets.StatusGeneral8100{} // Year 0 marks the status invalid
	}
	t.status.Store(rec)
	return err
}

// Status returns the latest ACU status, which callers must not modify.
// Before the first UpdateStatus it's empty, with Year 0.
func (t Telescope) Status() *datasets.StatusGeneral8100 {
	if rec := t.status.Load(); rec != nil {
		return rec
	}
	return new(datasets.StatusGeneral8100)
}

// statusTime returns the time of an ACU status record,
// whose Time field is the fractional day-of-year.
func statusTime(rec *datasets.StatusGeneral8100) time.Time {
	doy, frac := math.Modf(rec.Time)
	return acuTime(int(rec.Year), int32(doy), frac*24*60*60)
}

func (t *Telescope) CalibrateTime() error {
	rec := t.Status()
	if rec.Year == 0 {
		return fmt.Errorf("can't calibrate time: can't contact ACU")
	}
	offset := statusTime(rec).Sub(time.Now())
	if offset.Abs() > maxClockOffset {
		return fmt.Errorf("ACU/TCS clock mismatch too large to correct: %v", offset)
	}
	slog.Info("ACU - TCS clock difference", "offset", offset)
	t.pointing.tOffset = offset
	return nil
}

func (t Telescope) Ready() error {
	rec := t.Status()
	if rec.Year == 0 {
		return fmt.Errorf("can't contact ACU")
	}
	if rec.Year > 2024 {
		mismatch := statusTime(rec).Sub(time.Now().Add(t.pointing.tOffset))
		if mismatch.Abs() > 2*time.Second {
			return fmt.Errorf("ACU & TCS clock mismatch: %v", mismatch)
		}
	}
	if !rec.Remote {
		return fmt.Errorf("ACU not in remote mode")
	}
	if rec.AzimuthComputerDisabled != 0 {
		return fmt.Errorf("ACU azimuth computer disabled")
	}
	if rec.ElevationComputerDisabled != 0 {
		return fmt.Errorf("ACU elevation computer disabled")
	}
	if rec.AzimuthAxisDisabled != 0 {
		return fmt.Errorf("ACU azimuth axis disabled")
	}
	if rec.ElevationAxisDisabled != 0 {
		return fmt.Errorf("ACU elevation axis disabled")
	}
	var extra datasets.StatusExtra8100
	err := t.acu.DatasetGet("StatusExtra8100", &extra)
	if err != nil {
		return err
	}
	if !extra.AzimuthProfilerActive {
		return fmt.Errorf("ACU azimuth profiler not active")
	}
	if !extra.ElevationProfilerActive {
		return fmt.Errorf("ACU elevation profiler not active")
	}
	return nil
}

func (t Telescope) EnablePositionBroadcast(host string, port int) error {
	return t.acu.PositionBroadcastEnable(host, port)
}

func (t Telescope) Stop() error {
	return t.acu.ModeSet("Stop")
}

// SetMode switches both axes to mode, stopping them first.
func (t Telescope) SetMode(mode string) error {
	// ICD Section 9.1: "Before commanding or setting up a new mode,
	// it is best practice to set the antenna to Stop mode first."
	err := t.acu.ModeSet("Stop")
	if err != nil {
		return err
	}
	return t.acu.ModeSet(mode)
}

func (t Telescope) MoveTo(az, el float64) error {
	// ICD Section 9.1: "Before commanding or setting up a new mode,
	// it is best practice to set the antenna to Stop mode first."
	err := t.acu.ModeSet("Stop")
	if err != nil {
		return err
	}

	// set preset position and go
	rawAz, rawEl := t.pointing.Sky2Raw(az, el)
	err = t.acu.PresetPositionSet(rawAz, rawEl)
	if err != nil {
		return err
	}
	return t.acu.ModeSet("Preset")
}

// postTrackPoints sends program track points to housekeeping.
// An empty list means the program track was cleared, aborted, or failed.
// Errors are logged and otherwise ignored.
func (t Telescope) postTrackPoints(points []ScanPatternSample) {
	url := t.trackUploadURL
	if url == "" {
		return
	}
	t.posted.Store(len(points) > 0)
	err := postJSON(url, &struct {
		Points []ScanPatternSample `json:"points"`
	}{
		Points: points,
	})
	if err != nil {
		slog.Warn("housekeeping post failed", "err", err)
	}
}

// ClearProgramTrack clears the ACU program track stack, and tells housekeeping.
func (t Telescope) ClearProgramTrack() error {
	err := t.acu.ProgramTrackClear()
	if err != nil {
		return err
	}
	t.postTrackPoints(nil)
	return nil
}

// EndProgramTrack is called when a command ends, after its uploads have exited.
// If the command sent program track points to housekeeping but didn't finish
// normally (it was aborted or failed), housekeeping is sent an empty list.
func (t Telescope) EndProgramTrack(finished bool) {
	if t.posted.Swap(false) && !finished {
		t.postTrackPoints(nil)
	}
}

// UploadScanPattern uploads a program track in batches.
// firstBatch is closed once the first batch has been uploaded.
func (t Telescope) UploadScanPattern(ctx context.Context, pattern ScanPattern, firstBatch chan<- struct{}) error {
	iter := pattern.Iterator()
	total := 0
	samples := make([]ScanPatternSample, maxFreeProgramTrackStack)
	pts := make([]datasets.TimePositionTransfer, maxFreeProgramTrackStack)
	var status datasets.StatusGeneral8100

	for {
		err := t.acu.StatusGeneral8100Get(&status)
		if err != nil {
			slog.Error("failed to get ACU status", "err", err)
			return err
		}
		nmax := int(status.QtyOfFreeProgramTrackStackPositions)
		if nmax == 0 {
			return fmt.Errorf("upload: ACU program track stack is full")
		}

		// upload batch
		n := 0
		for !pattern.Done(iter) {
			x := &samples[n]
			err := pattern.Next(iter, x)
			if err != nil {
				slog.Error("pattern error", "err", err)
				break
			}

			rawT, rawAz, rawEl, rawVaz, rawVel := t.pointing.Track2Raw(
				x.T,
				x.Az,
				x.El,
				x.AzVel,
				x.ElVel,
			)
			err = checkAzEl(rawAz, rawEl, rawVaz, rawVel)
			if err != nil {
				return err
			}

			pt := &pts[n]
			pt.Day, pt.TimeOfDay = programTrackTime(rawT)
			pt.AzPosition = rawAz
			pt.ElPosition = rawEl
			pt.AzVelocity = rawVaz
			pt.ElVelocity = rawVel
			pt.AzFlag = x.AzFlag
			pt.ElFlag = x.ElFlag

			n++
			if n == nmax {
				break
			}
		}

		if n <= 0 {
			return fmt.Errorf("upload: no points")
		}

		// don't upload a batch computed after the command ended
		if ctx.Err() != nil {
			slog.Info("upload: cancelled")
			return nil
		}

		total += n
		slog.Info("upload: adding points", "n", n)
		err = t.acu.ProgramTrackAdd(pts[:n])
		if err != nil {
			return err
		}
		if total == n {
			close(firstBatch)
		}

		t.postTrackPoints(samples[:n])

		if pattern.Done(iter) {
			slog.Info("upload: done", "total", total)
			return nil
		}

		// sleep until we can upload the next batch
		lastT := samples[n-1].T
		wait := time.Until(lastT) / 2
		slog.Info("upload: waiting for next batch", "wait", wait.Round(time.Second))
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			slog.Info("upload: cancelled")
			return nil
		}
	}
}
