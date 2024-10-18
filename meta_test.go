package gokart

import (
	"log"
	"path/filepath"
	"testing"
)

var (
	Ancenis20240914 = filepath.Join(".", "data", "20240914T1112_Ancenis.mp4")
)

func TestProbeAncenis20240914(t *testing.T) {
	testProbe(t, Ancenis20240914, 113437, 10460)
}

func testProbe(t *testing.T, filename string, refAccl, refGps5 int) {
	m, err := GetStreamsCodecTag(filename)
	if err != nil {
		t.Error(err)
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
		t.Errorf("unable to find gpmd stream")
		return
	}
	values, err := ReadTelemetry(filename, gpmdIndex)
	if err != nil {
		t.Error(err)
		return
	}
	accl := AcclWithTime(values)
	if refAccl != len(accl) {
		t.Errorf("accl len is %d should be %d", len(accl), refAccl)
	}
	gps5 := GpsWithTime(values)
	if refGps5 != len(gps5) {
		t.Errorf("gps5 len is %d should be %d", len(gps5), refGps5)
	}
	start, err := GetVideoStartTime(filename)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("ACCL start", accl[0].Time)
	log.Println("ACCL stop", accl[len(accl)-1].Time)
	log.Println("GPS start", gps5[0].Time)
	log.Println("GPS stop", gps5[len(gps5)-1].Time)
	log.Println("Start from video stream", start)
}
