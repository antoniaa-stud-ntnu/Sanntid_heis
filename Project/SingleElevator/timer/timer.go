package timer

import (
	
	"time"
)

var timerEndTime float64
var timerActive bool

func getWallTime() float64 {
	currentTime := time.Now()
	seconds := float64(currentTime.Unix())
	nanoseconds := float64(currentTime.Nanosecond())
	return seconds + nanoseconds*1e-9
}

func TimerStart(duration float64) {
	timerEndTime = getWallTime() + duration
	timerActive = true
}

func TimerStop() {
	timerActive = false
}

func TimerTimedOut() bool {
	return (timerActive && getWallTime() > timerEndTime)
}

