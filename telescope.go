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
	iter := pattern.Iterator()
	total := 0
	pts := make([]datasets.TimePositionTransfer, maxFreeProgramTrackStack)
	var status datasets.StatusGeneral8100

	err := t.acu.ProgramTrackClear()
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Millisecond) // wait for ProgramTrackClear to take effect

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
		now := time.Now().UTC()
		waitSecs := 86400*float64(pts[n-1].Day-int32(now.YearDay())) + (pts[n-1].TimeOfDay - DaySeconds(now))
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
