package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gokart"
)

// Ensure that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

func main() {
	inName := flag.String("in", "", "Required: GoPro MP4 file to read")
	outName := flag.String("out", "best_lap.png", "Output lap image name")
	lap := flag.Int("lap", 0, "Lap number to draw (0 for best)")
	mode := flag.String("mode", "acc", "Info to graph, speed or acc")
	path := flag.String("path", filepath.Join("..", "..", "data"), "Path for aerial images storage")
	flag.Parse()

	if *inName == "" {
		flag.Usage()
		return
	}
	tele, err := gokart.ReadGoProTelemetry(*inName)
	if err != nil {
		log.Fatalln("Unable to get GoPro telemetry:", err)
		return
	}
	gps := gokart.GpsWithTime(tele)
	track := gokart.TheWorld.GetTrack(gps)
	lapCounter := gokart.NewLapCounter(track)
	fmt.Println("Track:", track.Name)
	for gps_index := range gps {
		if gps_index == 0 {
			// need at least 2 points
			continue
		}
		lapCounter.Update(gps[gps_index-1].Time, gps[gps_index-1], gps[gps_index])
	}
	lapnbr := lapCounter.Best()
	if *lap != 0 {
		lapnbr = *lap
	}
	fmt.Println("drawning lap", lapnbr)
	rgba, err := lapCounter.DrawLap(*path, *mode, gps, lapnbr)
	if err != nil {
		log.Fatal(err)
	}
	imgFile, err := os.Create(*outName)
	if err != nil {
		log.Fatal(err)
	}
	if err = png.Encode(imgFile, rgba); err != nil {
		return
	}
}
