package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gocv.io/x/gocv"

	"github.com/Serli/gokart"
)

// Ensure that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

func main() {
	inName := flag.String("in", filepath.Join("..", "..", "data", "20240914T1112_Ancenis.mp4"), "Required: GoPro MP4 file to read")
	outPath := flag.String("path", ".", "path to store generated files")
	start := flag.Int("start", 0, "Start frame number")
	stop := flag.Int("stop", -1, "Stop frame number")
	export := flag.Bool("export", false, "Export each image as frame_{number}.png WARNING can fill your drive!!!!")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}
	var (
		webcam *gocv.VideoCapture
		data   []byte
		err    error
	)
	// open webcam
	webcam, err = gocv.OpenVideoCapture(*inName)
	if err != nil {
		log.Fatalf("error opening video capture device: %v\n", *inName)
	}
	defer webcam.Close()
	tele, err := gokart.ReadGoProTelemetry(*inName)
	if err != nil {
		log.Fatalf("Unable to get GoPro telemetry: %s\n", err)
	}
	gps := gokart.GpsWithTime(tele)
	// "best" video starting point time
	startTime := gps[0].Time
	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()
	count := 0
	for {
		if ok := webcam.Read(&img); !ok {
			return
		}
		t := startTime.Add(time.Duration(int64(webcam.Get(gocv.VideoCapturePosMsec)+0.5)) * time.Millisecond)
		count++
		if img.Empty() {
			continue
		}
		if count < *start {
			continue
		}
		if *stop != -1 && count > *stop {
			break
		}
		gps_index := gokart.FindIndex(t, gps)
		if gps_index != -1 && gps_index < (len(gps)-1) {
			inter, err := gokart.Interpolate(t, gps[gps_index], gps[gps_index+1])
			if err != nil {
				log.Fatal(err)
			}
			pos := inter.Value.(gokart.GPS5)
			fmt.Printf(
				"frame %d latitude:%f longitude:%f accuracy (in cm):%d\n",
				count, pos.Latitude, pos.Longitude, pos.Accuracy)
			if *export {
				path := filepath.Join(*outPath, fmt.Sprintf("part%03d", count/1000))
				os.MkdirAll(path, os.ModePerm)
				// write image
				gocv.IMWrite(
					filepath.Join(path, fmt.Sprintf("frame_%d.png", count)),
					img)
				if data, err = json.MarshalIndent(pos, "", "  "); err != nil {
					return
				}
				// write json !
				if err = os.WriteFile(
					filepath.Join(path, fmt.Sprintf("frame_%d.json", count)),
					data, os.ModePerm); err != nil {
					return
				}
			}
		}
	}
}
