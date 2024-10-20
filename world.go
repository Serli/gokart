package gokart

import (
	"embed"
	"encoding/json"
	"log"
	"math"
)

// World, all available tracks
type World struct {
	Tracks []*Track `json:"tracks"`
}

// GetTrack from given GPS positions find the closest known track
func (w World) GetTrack(points []Timely) (t *Track) {
	minD := -1.0
	for _, tr := range w.Tracks {
		for _, pt := range points {
			gps := pt.Value.(GPS5)
			if gps.Accuracy >= 10000 {
				// not precise enough
				continue
			}
			if minD < 0 || math.Abs(tr.To(gps)) < minD {
				minD = math.Abs(tr.To(gps))
				t = tr
			}
		}
	}
	return
}

//go:embed data/theworld.json
var content embed.FS

// TheWorld world from ../data/theworld.json
var TheWorld World

func init() {
	// Read the embedded JSON file
	data, err := content.ReadFile("data/theworld.json")
	if err != nil {
		// should never happen as embed
		log.Println("unable to open embed theworld.json", err)
		return
	}
	// Unmarshal the JSON data into a struct
	if err := json.Unmarshal(data, &TheWorld); err != nil {
		log.Println("unable to unmarshal TheWorld", err)
		return
	}
}

func ExtractLimits(gps []Timely) (limits Line) {
	for _, p := range gps {
		v := p.Value.(GPS5)
		if limits.P1.Latitude == 0. || v.Latitude < limits.P1.Latitude {
			limits.P1.Latitude = v.Latitude
		}
		if limits.P1.Longitude == 0. || v.Longitude < limits.P1.Longitude {
			limits.P1.Longitude = v.Longitude
		}
		if limits.P2.Latitude == 0. || v.Latitude > limits.P2.Latitude {
			limits.P2.Latitude = v.Latitude
		}
		if limits.P2.Longitude == 0. || v.Longitude > limits.P2.Longitude {
			limits.P2.Longitude = v.Longitude
		}
	}
	return
}
