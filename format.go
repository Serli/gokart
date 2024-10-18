package gokart

import (
	"fmt"
	"time"
)

func DurationToChrono(d time.Duration) string {
	ms := d.Milliseconds()
	m := ms / (60 * 1000)
	s := (ms - m*60*1000) / 1000
	cs := (ms - ((m*60)+s)*1000) / 10
	return fmt.Sprintf("%d:%02d.%02d", m, s, cs)
}
