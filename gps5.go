package gokart

import (
	"time"

	"github.com/cedricjoulain/gopro-utils/telemetry"
)

// GPS5 enriched GPS5 from telemetry
type GPS5 struct {
	telemetry.GPS5
	Accuracy uint16 // gps accuracy in cm
}

// NewGPS5 simple constructor from Latitude and Longitude
func NewGPS5(lat, lon float64) (g GPS5) {
	g.Latitude = lat
	g.Longitude = lon
	return
}

// Line two points defining a "line" like start line
// can also be used for segment
type Line struct {
	P1 GPS5
	P2 GPS5
}

func (l Line) IsZero() bool {
	if l.P1.Latitude != 0 {
		return false
	}
	if l.P1.Longitude != 0 {
		return false
	}
	if l.P2.Latitude != 0 {
		return false
	}
	if l.P2.Longitude != 0 {
		return false
	}
	return true
}

// NewLine simple constructor from  2 Latitudes and Longitudes
func NewLine(lat1, lon1, lat2, lon2 float64) (l Line) {
	l.P1 = NewGPS5(lat1, lon1)
	l.P2 = NewGPS5(lat2, lon2)
	return
}

func (l Line) To(g GPS5) float64 {
	return l.Side(g) * Distance(NewGPS5(
		(l.P1.Latitude+l.P2.Latitude)/2,
		(l.P1.Longitude+l.P2.Longitude)/2),
		g)
}

func (l Line) Side(g GPS5) float64 {
	s := (g.Latitude-l.P1.Latitude)*(l.P2.Longitude-l.P1.Longitude) -
		(g.Longitude-l.P1.Longitude)*(l.P2.Latitude-l.P1.Latitude)
	if s >= 0 {
		return 1
	}
	return -1
}

// GpsWithTime retrieve GPS5 list, enriched time and accuracy
func GpsWithTime(values []*telemetry.TELEM) (all []Timely) {
	all = make([]Timely, 0)
	for i, v := range values {
		if v.Time.Time.IsZero() {
			continue
		}
		available := v.Gps
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
				Value: GPS5{GPS5: value, Accuracy: v.GpsAccuracy.Accuracy},
			})
		}
	}
	return
}
