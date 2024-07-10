package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ccatobs/antenna-control-unit/datasets"
)

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

func statusTime(t time.Time) (uint32, float64) {
	doy, tod := VertexTime(t)
	return uint32(t.UTC().Year()), float64(doy) + tod/(24*60*60)
}

func (t Telescope) Ready() error {
	if t.rec.Year == 0 {
		return fmt.Errorf("can't contact ACU")
	}
	if t.rec.Year > 2024 {
		y, d := statusTime(time.Now())
		dy := t.rec.Year - y
		dt := math.Abs(t.rec.Time-d) * 24 * 60 * 60
		if dy != 0 || dt > 2 {
			return fmt.Errorf("ACU & TCS clock mismatch: %d years, %g seconds", dy, dt)
		}
	}
	if !t.rec.Remote {
		return fmt.Errorf("ACU not in remote mode")
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
	rawAz, rawEl, _, _ := t.pointing.Sky2Raw(az, el, 0, 0)
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
	pts := make([]datasets.TimePositionTransfer, maxFreeProgramTrackStack)
	var status datasets.StatusGeneral8100

	for {
		err := t.acu.StatusGeneral8100Get(&status)
		if err != nil {
			log.Print("failed to get ACU status: ", err)
			return err
		}
		nmax := int(status.QtyOfFreeProgramTrackStackPositions)

		// upload batch
		n := 0
		for !pattern.Done(iter) {
			pt := &pts[n]
			err := pattern.Next(iter, pt)
			if err != nil {
				log.Printf("pattern error: %v", err)
				break
			}

			rawAz, rawEl, rawVaz, rawVel := t.pointing.Sky2Raw(
				pt.AzPosition,
				pt.ElPosition,
				pt.AzVelocity,
				pt.ElVelocity,
			)
			err = checkAzEl(rawAz, rawEl, rawVaz, rawVel)
			if err != nil {
				return err
			}

			pt.AzPosition = rawAz
			pt.ElPosition = rawEl
			pt.AzVelocity = rawVaz
			pt.ElVelocity = rawVel

			n++
			if n == nmax {
				break
			}
		}
		total += n
		if n > 0 {
			log.Printf("upload: adding %d points", n)
			err := t.acu.ProgramTrackAdd(pts[:n])
			if err != nil {
				return err
			}
		}

		if pattern.Done(iter) {
			log.Printf("upload: done, %d points total", total)
			return nil
		}

		// sleep until we can upload the next batch
		doy, tod := VertexTime(time.Now())
		waitSecs := 86400*float64(pts[n-1].Day-doy) + (pts[n-1].TimeOfDay - tod)
		waitSecs = waitSecs / 2
		wait := time.Duration(waitSecs) * time.Second
		log.Printf("upload: next batch in %.3g minutes", wait.Minutes())

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			log.Print("upload: cancelled")
			return nil
		}
	}
}
