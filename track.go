package gokart

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TILE_SIZE size of tile used for map
const TILE_SIZE = 512

var (
	// color for the rect when faces detected
	// blue := color.RGBA{0, 0, 255, 0}
	white  = color.RGBA{255, 255, 255, 255}
	red    = color.RGBA{255, 0, 0, 255}
	green  = color.RGBA{0, 255, 0, 255}
	blue   = color.RGBA{0, 0, 255, 255}
	orange = color.RGBA{255, 180, 0, 255}
)

// Track informations like start lines sectors...
// Also some logos for overlay option
type Track struct {
	Name     string      `json:"name"`
	LogoFile string      `json:"logofile,omitempty"`
	LogoMask string      `json:"logomask,omitempty"`
	Map      *image.RGBA `json:"map,omitempty"`
	Start    Line        `json:"start"`
	Sectors  []Line      `json:"sectors"`
	Limits   Line        `json:"limits"`
}

// SetLimits, update track bounding box
// used for map overlay
func (t *Track) SetLimits(limits Line) {
	t.Limits = limits
	// for debug
	log.Println(t.Name, "New Limits", limits)
}

// To distance to start line
func (t Track) To(g GPS5) float64 {
	return t.Start.To(g)
}

// NewLapStart do we cross start line and when
func (t Track) NewLapStart(g1, g2 Timely) (start time.Time) {
	return Crossed(t.Start, g1, g2)
}

// ShortName track name usable by OS (filename...)
func (t Track) ShortName() string {
	return strings.ReplaceAll(t.Name, " ", "")
}

// ImageFileName name of aerial map file
func (t Track) ImageFileName(path string) string {
	return filepath.Join(path, fmt.Sprintf("%s.png", t.ShortName()))
}

// UpdateMap read given image will be used as aerial image
func (t *Track) UpdateMap(path string) (err error) {
	if t.Map != nil {
		// map already loaded
	}
	var imgFile *os.File
	if imgFile, err = os.Open(t.ImageFileName(path)); err != nil {
		return
	}
	// iwidth x iheight files to close
	defer imgFile.Close()
	var (
		img any
		ok  bool
	)
	if img, _, err = image.Decode(imgFile); err != nil {
		return
	}
	if t.Map, ok = img.(*image.RGBA); !ok {
		err = fmt.Errorf("map image is not *image.RGBA but %T", img)
	}
	return
}

// PosToXY given map boundaries and lat lon return x y int map
func (t Track) PosToXY(r image.Rectangle, lat, lon float64) (x, y int) {
	xratio := (lon - t.Limits.P1.Longitude) / (t.Limits.P2.Longitude - t.Limits.P1.Longitude)
	yratio := 1.0 - ((lat - t.Limits.P1.Latitude) / (t.Limits.P2.Latitude - t.Limits.P1.Latitude))
	x = int(float64(r.Max.X-r.Min.X-TILE_SIZE)*xratio+0.5) + (TILE_SIZE / 2) + 64
	y = int(float64(r.Max.Y-r.Min.Y-TILE_SIZE)*yratio+0.5) + (TILE_SIZE / 2) + 64
	return
}

// Crossed do we cross line and when
func Crossed(line Line, g1, g2 Timely) (t time.Time) {
	g1Side := line.To(g1.Value.(GPS5))
	g2Side := line.To(g2.Value.(GPS5))
	if g1Side*g2Side > 0. {
		// we don't cross line
		return
	}
	if math.Abs(g1Side) > 8.0 || math.Abs(g2Side) > 8.0 {
		// too far (> 8 m) from line
		return
	}
	from1 := math.Abs(g1Side) / (math.Abs(g1Side) + math.Abs(g2Side))
	frominNano := float64(g2.Time.Sub(g1.Time)) * from1
	t = g1.Time.Add(time.Duration(frominNano + 0.5))
	return
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//
//	a given longitude and latitude relatively accurately (using a spherical
//	approximation of the Earth) through the Haversin Distance Formula for
//	great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(g1, g2 GPS5) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = g1.Latitude * math.Pi / 180
	lo1 = g1.Longitude * math.Pi / 180
	la2 = g2.Latitude * math.Pi / 180
	lo2 = g2.Longitude * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// LapCounter keep all trak lap
type LapCounter struct {
	track       *Track
	laps        [][]time.Time
	best        int
	bestD       time.Duration
	bestSectors []time.Duration
	current     int
	status      []int

	prevstatus []int
}

// NewLapCounter create a LapCounter from given track
func NewLapCounter(track *Track) (l LapCounter) {
	l.track = track
	l.laps = make([][]time.Time, 0)
	// No best
	l.best = -1
	l.appendEmptyLap()
	l.current = 0
	l.bestSectors = make([]time.Duration, len(l.track.Sectors)+1)
	l.status = make([]int, len(l.track.Sectors)+1)
	l.prevstatus = make([]int, len(l.track.Sectors)+1)
	return
}

// Track associated track
func (l LapCounter) Track() *Track {
	return l.track
}

// SectorStatus status of each sector
func (l LapCounter) SectorStatus() []int {
	return l.status
}

// PrevStatus status of previsou sectors
func (l LapCounter) PrevStatus() []int {
	return l.prevstatus
}

func (l LapCounter) PrevTime() time.Duration {
	return l.laps[l.current][0].Sub(l.laps[l.current-1][0])
}

func (l *LapCounter) TheroreticalBest() (best time.Duration) {
	for _, d := range l.bestSectors {
		if d == 0 {
			// missing sector
			return 0
		}
		best += d
	}
	return
}

// appendEmptyLap, sector times
func (l *LapCounter) appendEmptyLap() {
	l.laps = append(l.laps, make([]time.Time, 1+len(l.track.Sectors)))
}

// Update lapcounter with information at t
func (l *LapCounter) Update(t time.Time, prev, current Timely) {
	// New lap ?
	if newStart := l.track.NewLapStart(prev, current); !newStart.IsZero() {
		var lapD time.Duration
		if !l.laps[l.current][0].IsZero() {
			// one more full lap
			lapD = newStart.Sub(l.laps[l.current][0])
			if l.bestD == 0 || lapD < l.bestD {
				l.bestD = lapD
				l.best = l.current
			}
		}
		if len(l.track.Sectors) > 0 && !l.laps[l.current][len(l.laps[l.current])-1].IsZero() {
			// last sector
			l.updateSector(len(l.track.Sectors), newStart.Sub(l.laps[l.current][len(l.laps[l.current])-1]), newStart)
		}
		l.appendEmptyLap()
		l.current++
		l.laps[l.current][0] = newStart
		// cleanup status
		for i := range l.status {
			l.prevstatus[i] = l.status[i]
			l.status[i] = 0
		}
		l.PrintLap(l.current - 1)
	}
	// look at sectors
	for i, sector := range l.track.Sectors {
		if cross := Crossed(sector, prev, current); !cross.IsZero() {
			// TODO check for right order!
			l.laps[l.current][i+1] = cross
			// valid sector?
			if !l.laps[l.current][i].IsZero() {
				// it's not last sector so no need for nextStart
				l.updateSector(i, cross.Sub(l.laps[l.current][i]), time.Time{})
			}
		}
	}
}

func (l *LapCounter) updateSector(i int, d time.Duration, nextStart time.Time) {
	if l.bestSectors[i] == 0 || d < l.bestSectors[i] {
		// new best sector
		l.bestSectors[i] = d
		l.status[i] = 2
	} else {
		// Already a best lap ?
		if l.best >= 0 {
			if !l.laps[l.best][i].IsZero() {
				var before time.Duration
				if i+1 == len(l.laps[l.best]) {
					// last sector, use next start
					before = nextStart.Sub(l.laps[l.best][i])
				} else {
					if !l.laps[l.best][i+1].IsZero() {
						// something to compare
						before = l.laps[l.best][i+1].Sub(l.laps[l.best][i])
					}
				}
				if before != 0 {
					if before < d {
						l.status[i] = -1
					} else {
						l.status[i] = 1
					}
				}
			}
		} else {
			l.status[i] = 1
		}
	}
}

func (l LapCounter) PrintLap(lap int) {
	var (
		b strings.Builder
		d time.Duration
	)
	if !l.laps[lap][0].IsZero() {
		d = l.laps[lap+1][0].Sub(l.laps[lap][0])
	}
	fmt.Fprintf(&b, "Lap %02d %s", lap, DurationToChrono(d))
	for i, t := range l.laps[lap] {
		if i == 0 {
			// start line
			continue
		}
		d = 0
		if !l.laps[lap][i-1].IsZero() && !t.IsZero() {
			d = t.Sub(l.laps[lap][i-1])
		}
		fmt.Fprintf(&b, " S%02d %s", i, DurationToChrono(d))
	}
	// last sector
	last := len(l.laps[lap])
	if last > 1 {
		d = 0
		if !l.laps[lap][last-1].IsZero() {
			d = l.laps[lap+1][0].Sub(l.laps[lap][last-1])
		}
		fmt.Fprintf(&b, " S%02d %s", last, DurationToChrono(d))
	}
	fmt.Println(b.String())
}

func (l LapCounter) Current() int {
	return l.current
}

func (l LapCounter) CurrentTime(t time.Time) (d time.Duration) {
	if l.current < 1 {
		return
	}
	d = t.Sub(l.laps[l.current][0])
	return
}

func (l LapCounter) Best() int {
	if l.best > -1 {
		return l.best
	}
	return 0
}

func (l LapCounter) BestTime() (d time.Duration) {
	if l.best > 0 {
		return l.laps[l.best+1][0].Sub(l.laps[l.best][0])
	}
	return
}

func GetSpeed(gps []Timely, index int) (value float64) {
	return gps[index].Value.(GPS5).Speed3D
}

// ACC_DELTA how many step in past and future to compute acceleration
const ACC_DELTA = 12

func GetAcc(gps []Timely, index int) (value float64) {
	// "average" on
	istart := index - ACC_DELTA
	if istart < 0 {
		istart = 0
	}
	istop := index
	if istop == 0 {
		istop = 1
	}
	return (gps[istop].Value.(GPS5).Speed3D - gps[istart].Value.(GPS5).Speed3D) / gps[istop].Time.Sub(gps[istart].Time).Seconds()
}

func GetAccColor(acc, min, max float64) (color color.RGBA) {
	color.A = 255
	if acc < 0.0 {
		ratio := acc / min
		color.R = uint8(ratio*float64(red.R) + (1-ratio)*float64(white.R))
		color.G = uint8(ratio*float64(red.G) + (1-ratio)*float64(white.G))
		color.B = uint8(ratio*float64(red.B) + (1-ratio)*float64(white.B))
	} else {
		ratio := acc / max
		color.R = uint8(ratio*float64(green.R) + (1-ratio)*float64(white.R))
		color.G = uint8(ratio*float64(green.G) + (1-ratio)*float64(white.G))
		color.B = uint8(ratio*float64(green.B) + (1-ratio)*float64(white.B))
	}
	return
}

func (l LapCounter) DrawLap(path, mode string, gps []Timely, index int) (rgba image.Image, err error) {
	getValue := GetSpeed
	switch mode {
	case "acc":
		getValue = GetAcc
	}
	gpsStart := FindIndex(l.laps[index][0], gps)
	gpsStop := FindIndex(l.laps[index+1][0], gps)
	log.Println("GPS start", gpsStart, gps[gpsStart])
	log.Println("GPS stop", gpsStop, gps[gpsStop])
	minMode := 300000000.0
	maxMode := 0.0
	for i := gpsStart; i <= gpsStop; i++ {
		if getValue(gps, i) > maxMode {
			maxMode = getValue(gps, i)
		}
		if getValue(gps, i) < minMode {
			minMode = getValue(gps, i)
		}
	}
	log.Println(mode, "between", minMode, "and", maxMode)
	// ensure map loaded
	if err = l.track.UpdateMap(path); err != nil {
		return
	}
	// draw all gps lines
	rgba = l.track.Map
	r := rgba.Bounds()
	for i := gpsStart; i < gpsStop; i++ {
		x1, y1 := l.track.PosToXY(r, gps[i].Value.(GPS5).Latitude, gps[i].Value.(GPS5).Longitude)
		x2, y2 := l.track.PosToXY(r, gps[i+1].Value.(GPS5).Latitude, gps[i+1].Value.(GPS5).Longitude)
		color := white
		switch mode {
		case "speed", "res":
			ratio := (getValue(gps, i) - minMode) / (maxMode - minMode)
			color = green
			if ratio < 0.5 {
				color.R = uint8((2*ratio)*float64(green.R) + (1-2*ratio)*float64(blue.R))
				color.G = uint8((2*ratio)*float64(green.G) + (1-2*ratio)*float64(blue.G))
				color.B = uint8((2*ratio)*float64(green.B) + (1-2*ratio)*float64(blue.B))
			}
			if ratio > 0.5 {
				ratio -= 0.5
				color.R = uint8((2*ratio)*float64(red.R) + (1-2*ratio)*float64(green.R))
				color.G = uint8((2*ratio)*float64(red.G) + (1-2*ratio)*float64(green.G))
				color.B = uint8((2*ratio)*float64(red.B) + (1-2*ratio)*float64(green.B))
			}
			if mode == "res" {
				// accuracy in meters
				accuracy := float64(gps[i].Value.(GPS5).Accuracy) / 100.
				// radius, accuracy in pixels
				// accuracy seems not to be so accurate...
				radius := 3 * accuracy / MeterPerPixel(gps[i].Value.(GPS5).Latitude)
				DrawCircle(rgba.(*image.RGBA), x1, y1, int(radius+0.5), color)
			} else {
				DrawCircleLine(rgba.(*image.RGBA), x1, y1, x2, y2, 3, color)
			}
		case "acc":
			DrawCircleLine(
				rgba.(*image.RGBA), x1, y1, x2, y2, 3,
				GetAccColor(getValue(gps, i), minMode, maxMode))
		}
	}
	return
}

// MeterPerPixel for zoom = 20
func MeterPerPixel(lat float64) float64 {
	return 40075016.686 * math.Abs(math.Cos(lat*math.Pi/180)) / math.Pow(2, 20+8)
}
