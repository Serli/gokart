package gokart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stilldavid/gopro-utils/telemetry"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// GetStreamsCodecTag, retreive streams index and codec
func GetStreamsCodecTag(filename string) (m map[int]string, err error) {
	m = make(map[int]string)
	probejson, perr := ffmpeg.Probe(filename)
	if perr != nil {
		err = fmt.Errorf("probe access:%s", perr)
		return
	}
	var probe map[string]any
	if err = json.Unmarshal([]byte(probejson), &probe); err != nil {
		err = fmt.Errorf("cannot unmarshal json probe:%s", err)
		return
	}
	raw, ok := probe["streams"]
	if !ok {
		err = fmt.Errorf("unable to find streams")
		return
	}
	streams, ok := raw.([]any)
	if !ok {
		err = fmt.Errorf("wrong type for streams:%T", raw)
		return
	}
	for _, raw := range streams {
		stream, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if raw, ok = stream["codec_tag_string"]; !ok {
			continue
		}
		tag, ok := raw.(string)
		if !ok {
			continue
		}
		if raw, ok = stream["index"]; !ok {
			continue
		}
		// default type for number ?
		index, ok := raw.(float64)
		if !ok {
			continue
		}
		m[int(index)] = tag
	}
	return
}

func ReadTelemetry(filename string, index int) (values []*telemetry.TELEM, err error) {
	gpmd := bytes.NewBuffer(nil)
	if err = ffmpeg.Input(filename).
		Get(strconv.Itoa(index)).
		Output("pipe:", ffmpeg.KwArgs{"codec": "copy", "format": "rawvideo"}).
		WithOutput(gpmd, os.Stdout).
		Run(); err != nil {
		return
	}
	values = make([]*telemetry.TELEM, 0)
	for {
		t, terr := telemetry.Read(gpmd)
		if terr != nil {
			if terr == io.EOF {
				break
			}
			err = terr
			return
		}
		if t == nil {
			break
		}
		values = append(values, t)
	}
	return
}

type Timely struct {
	Time  time.Time
	Value any
}

// GetVideoStartTime, clearly wrong, use first telemery time instead
func GetVideoStartTime(filename string) (start time.Time, err error) {
	data, err := ffmpeg.Probe(filename, nil)
	if err != nil {
		err = fmt.Errorf("error probing: %v", filename)
		return
	}
	infos := make(map[string]any)
	if err = json.Unmarshal([]byte(data), &infos); err != nil {
		err = fmt.Errorf("unable to unmarshall video json probe:%s", err)
		return
	}
	raw, ok := infos["streams"]
	if !ok {
		err = fmt.Errorf("no streams key in probe infos")
		return
	}
	streams, ok := raw.([]any)
	if !ok {
		err = fmt.Errorf("result should be []any is %T", raw)
		return
	}
	// loop over stream to find first video
	err = fmt.Errorf("unfound")
	for _, raw := range streams {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		// video ?
		c, ok := m["codec_type"]
		if !ok {
			continue
		}
		s, ok := c.(string)
		if !ok || s != "video" {
			continue
		}
		// we have a video stream look for tags
		t, ok := m["tags"]
		if !ok {
			continue
		}
		tags, ok := t.(map[string]any)
		if !ok {
			continue
		}
		c, ok = tags["creation_time"]
		if !ok {
			continue
		}
		s, ok = c.(string)
		if !ok {
			continue
		}
		// TODO find location in Meta data
		var location *time.Location
		if location, err = time.LoadLocation("Europe/Paris"); err != nil {
			return
		}
		if start, err = time.Parse(time.RFC3339, s); err != nil {
			return
		}
		// Hard change as string is wrong on my GoPro
		start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), location)
		c, ok = tags["timecode"]
		if !ok {
			continue
		}
		s, ok = c.(string)
		if !ok {
			continue
		}
		if parts := strings.Split(s, ":"); len(parts) == 4 {
			// image position in seconde ?
			pos, perr := strconv.Atoi(parts[3])
			if perr != nil {
				err = fmt.Errorf("unable to parse image position in timecode %s:%s", s, perr)
				return
			}
			c, ok = tags["r_frame_rate"]
			if !ok {
				continue
			}
			s, ok = c.(string)
			if !ok {
				continue
			}
			switch s {
			case "25/1":
				if pos < 0 || pos > 24 {
					err = fmt.Errorf("image pos for r_frame_rate %s should be in [0 24] we have:%d", s, pos)
					return
				}
				start = start.Add((time.Duration(pos) * time.Second) / time.Duration(25))
				err = nil
				return
			default:
				err = fmt.Errorf("unknown r_frame_rate %s", s)
				return
			}
		}
	}
	return
}

func ReadGoProTelemetry(filename string) (values []*telemetry.TELEM, err error) {
	var m map[int]string
	if m, err = GetStreamsCodecTag(filename); err != nil {
		return
	}
	// find gpmd
	gpmdIndex := -1
	for k, v := range m {
		if v == "gpmd" {
			gpmdIndex = k
		}
	}
	if gpmdIndex == -1 {
		err = fmt.Errorf("unable to find gpmd stream")
		return
	}
	values, err = ReadTelemetry(filename, gpmdIndex)
	return
}

func dichotomy(t time.Time, all []Timely, start, stop int) int {
	if stop == start+1 {
		return start
	}
	mid := (start + stop) / 2
	if mid == start {
		mid = start + 1
	}
	if t.Before(all[mid].Time) {
		return dichotomy(t, all, start, mid)
	} else {
		return dichotomy(t, all, mid, stop)
	}
}

// FindIndex return index just before t (index+1 will be after t)
func FindIndex(t time.Time, all []Timely) (index int) {
	index = -1
	if len(all) < 2 {
		// ??
		return
	}
	if t.Before(all[0].Time) {
		// not in range
		return
	}
	if t.After(all[len(all)-1].Time) {
		// not in range
		return
	}
	index = dichotomy(t, all, 0, len(all)-1)
	return
}
