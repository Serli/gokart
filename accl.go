package gokart

import (
	"math"
	"time"

	"github.com/stilldavid/gopro-utils/telemetry"
)

type ACCL telemetry.ACCL

// From https://www.nxp.com/docs/en/application-note/AN3461.pdf

// Pitch in degrees
func (a ACCL) Pitch() float64 {
	tanpitch := -a.X / math.Sqrt(a.Y*a.Y+a.Z*a.Z)
	return math.Atan(tanpitch) * 180. / math.Pi
}

// Roll in degrees
func (a ACCL) Roll() float64 {
	sign := 1.0
	if a.Z < 0. {
		sign = -1
	}
	tanroll := a.Y / (sign * math.Sqrt(0.01*a.X*a.X+a.Z*a.Z))
	return math.Atan(tanroll) * 180. / math.Pi
}

// AcclWithTime TODO find a way to generic it, []any is impossible
func AcclWithTime(values []*telemetry.TELEM) (all []Timely) {
	all = make([]Timely, 0)
	for i, v := range values {
		if v.Time.Time.IsZero() {
			continue
		}
		available := v.Accl
		if len(available) == 0 {
			// nothing ?
			continue
		}
		delta := time.Second.Nanoseconds()
		if i+1 < len(values) {
			// we have a real delta
			delta = time.Duration(values[i+1].Time.Time.Sub(v.Time.Time)).Nanoseconds()
			if delta > 2*time.Second.Nanoseconds() {
				// too far in time, adjust previous
				delta -= time.Second.Nanoseconds()
				for i := range all {
					all[i].Time = all[i].Time.Add(time.Duration(delta))
				}
				delta = time.Second.Nanoseconds()
			}
		}
		for j, value := range available {
			all = append(all, Timely{
				Time:  v.Time.Time.Add(time.Duration((int64(j) * delta) / int64(len(available)))),
				Value: ACCL(value),
			})
		}
	}
	return
}
