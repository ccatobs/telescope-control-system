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
	pts := make([]datasets.TimePositionTransfer, maxFreeProgramTrackStack)
	for {
		// XXX:TBD find out how many stack slots are free
		nmax := 2000

		iter := pattern.Iterator()

		// upload batch
		n := 0
		for !pattern.Done(iter) {
			err := pattern.Next(iter, &pts[n])
			if err != nil {
				log.Printf("pattern error: %v", err)
				break
			}

			// XXX:TBD correct velocity
			rawAz, rawEl := t.pointing.Sky2Raw(pts[n].AzPosition, pts[n].ElPosition)
			err = checkAzEl(rawAz, rawEl)
			if err != nil {
				return err
			}

			pts[n].AzPosition = rawAz
			pts[n].ElPosition = rawEl

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
		waitSecs := 86400*float64(pts[n-1].Day-pts[0].Day) + (pts[n-1].TimeOfDay - pts[0].TimeOfDay)
		wait := time.Duration(waitSecs) * time.Second
		log.Printf("upload: sleeping for %.3g minutes", wait.Minutes())

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			log.Print("upload: cancelled")
			return nil
		}
	}
}
