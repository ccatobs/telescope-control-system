package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
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
	rec      datasets.StatusGeneral8100
}

func NewTelescope(acu *ACU) *Telescope {
	return &Telescope{
		acu:      acu,
		pointing: NewPointing(),
	}
}

func (t *Telescope) UpdateStatus() error {
	t.rec.Year = 0 // invalidate current status
	return t.acu.StatusGeneral8100Get(&t.rec)
}

func (t Telescope) Status() *datasets.StatusGeneral8100 {
	return &t.rec
}

// statusTime returns the time of an ACU status record,
// whose Time field is the fractional day-of-year.
func statusTime(rec *datasets.StatusGeneral8100) time.Time {
	doy, frac := math.Modf(rec.Time)
	return acuTime(int(rec.Year), int32(doy), frac*24*60*60)
}

func (t *Telescope) CalibrateTime() error {
	if t.rec.Year == 0 {
		return fmt.Errorf("can't calibrate time: can't contact ACU")
	}
	offset := statusTime(&t.rec).Sub(time.Now())
	if offset.Abs() > maxClockOffset {
		return fmt.Errorf("ACU/TCS clock mismatch too large to correct: %v", offset)
	}
	slog.Info("ACU - TCS clock difference", "offset", offset)
	t.pointing.tOffset = offset
	return nil
}

func (t Telescope) Ready() error {
	if t.rec.Year == 0 {
		return fmt.Errorf("can't contact ACU")
	}
	if t.rec.Year > 2024 {
		mismatch := statusTime(&t.rec).Sub(time.Now().Add(t.pointing.tOffset))
		if mismatch.Abs() > 2*time.Second {
			return fmt.Errorf("ACU & TCS clock mismatch: %v", mismatch)
		}
	}
	if !t.rec.Remote {
		return fmt.Errorf("ACU not in remote mode")
	}
	var extra datasets.StatusExtra8100
	err := t.acu.DatasetGet("StatusExtra8100", &extra)
	if err != nil {
		return err
	}
	if !extra.AzimuthProfilerActive {
		return fmt.Errorf("azimuth profiler not active")
	}
	if !extra.ElevationProfilerActive {
		return fmt.Errorf("elevation profiler not active")
	}
	return nil
}

func (t Telescope) EnablePositionBroadcast(host string, port int) error {
	return t.acu.PositionBroadcastEnable(host, port)
}

func (t Telescope) Stop() error {
	return t.acu.ModeSet("Stop")
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

// UploadScanPattern uploads a program track in batches.
func (t Telescope) UploadScanPattern(ctx context.Context, pattern ScanPattern) error {
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

		total += n
		slog.Info("upload: adding points", "n", n)
		err = t.acu.ProgramTrackAdd(pts[:n])
		if err != nil {
			return err
		}

		// send points to housekeeping
		// XXX:FIXME temporary hack
		url := os.Getenv("XXX_PROGRAM_TRACK_UPLOAD_URL")
		if url != "" {
			err = postJSON(url, &struct {
				Points []ScanPatternSample `json:"points"`
			}{
				Points: samples[:n],
			})
			if err != nil {
				slog.Warn("upload: housekeeping post failed", "err", err)
				// ignore error
			}
		}

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
