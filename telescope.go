package main

import (
	"context"
	"log"
	"time"

	"github.com/ccatp/antenna-control-unit/datasets"
)

// Telescope provides a higher-level interface to the ACU.
// Responsible for pointing corrections and coordinate transformations.
type Telescope struct {
	acu      *ACU
	pointing Pointing
}

func NewTelescope(acu *ACU) *Telescope {
	return &Telescope{
		acu:      acu,
		pointing: NewPointing(),
	}
}

func (t Telescope) Stop() error {
	return t.acu.ModeSet("Stop")
}

func (t Telescope) MoveTo(az, el float64) error {
	rawAz, rawEl := t.pointing.Sky2Raw(az, el)
	err := t.acu.PresetPositionSet(rawAz, rawEl)
	if err != nil {
		return err
	}
	return t.acu.ModeSet("Preset")
}

// UploadScanPattern uploads a program track in batches.
func (t Telescope) UploadScanPattern(ctx context.Context, pattern ScanPattern) error {
	total := 0
	done := false
	pts := make([]datasets.TimePositionTransfer, maxFreeProgramTrackStack)
	for {
		// XXX:TBD find out how many stack slots are free
		nmax := 2000
		t0 := pattern.Time()

		// upload batch
		n := 0
		for ; n < nmax; n++ {
			done = !pattern.Next(&pts[n])
			if done {
				break
			}
			// XXX:TBD correct velocity
			rawAz, rawEl := t.pointing.Sky2Raw(pts[n].AzPosition, pts[n].ElPosition)
			pts[n].AzPosition = rawAz
			pts[n].ElPosition = rawEl
		}
		total += n
		if n > 0 {
			log.Printf("upload: adding %d points", n)
			err := t.acu.ProgramTrackAdd(pts[:n])
			if err != nil {
				return err
			}
		}
		if done {
			log.Printf("upload: done, %d points total", total)
			return nil
		}

		// sleep
		t := pattern.Time()
		wait := t.Sub(time.Now()) - t.Sub(t0)/4
		log.Printf("upload: sleeping for %.3g minutes", wait.Minutes())

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			log.Print("upload: cancelled")
			return nil
		}
	}
}
