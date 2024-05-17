package datasets

import (
	"fmt"
	"io"
	"math"
)

func (p TimePositionTransfer) WriteSSV(w io.Writer) error {
	h := math.Floor(p.TimeOfDay / (60 * 60))
	m := math.Floor(p.TimeOfDay/60 - 60*h)
	s := p.TimeOfDay - 60*60*h - 60*m
	// XXX: print time without rounding, perhaps with big.NewFloat(s).Text('f',-53)?
	_, err := fmt.Fprintf(w, "%03d, %02d:%02d:%012.9f;%g;%g;%g;%g;%d;%d\r\n", p.Day, int(h), int(m), s, p.AzPosition, p.ElPosition, p.AzVelocity, p.ElVelocity, p.AzFlag, p.ElFlag)
	return err
}

func (p *TimePositionTransfer) ReadSSV(r io.Reader) error {
	var h, m int
	var s float64
	_, err := fmt.Fscanf(r, "%d,%d:%d:%f;%f;%f;%f;%f;%d;%d\r\n", &p.Day, &h, &m, &s, &p.AzPosition, &p.ElPosition, &p.AzVelocity, &p.ElVelocity, &p.AzFlag, &p.ElFlag)
	p.TimeOfDay = 60*60*float64(h) + 60*float64(m) + s
	return err
}
