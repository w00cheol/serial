// Define Data type in this file
package main

import "time"

type TimeSeriesData struct {
	Time time.Time
	Data any
}

type TiltData struct {
	XAxis int64
	YAxis int64
	ZAxis int64
}

// It could be VibrationX, VibrationY, and also VibratoinY
type VibrationData struct {
	Peak [6]int64
	Amp  [6]float64
}

type SoundData struct {
	Peak [6]int64
	Amp  [6]float64
}
