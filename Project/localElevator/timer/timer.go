package timer

import (
	"time"
)

var endTime float64
var active bool



func getWallTime() float64 {
	currentTime := time.Now()
	seconds := float64(currentTime.Unix())
	nanoseconds := float64(currentTime.Nanosecond())
	return seconds + nanoseconds*1e-9
}

func Start(duration float64) {
	endTime = getWallTime() + duration
	active = true
}

func Stop() {
	active = false
}

func TimedOut() bool {
	return (active && getWallTime() > endTime)
}

