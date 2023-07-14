// Define Data type in this file
package main

import (
	"fmt"
	"time"
)

// All response data from sensor has to follow(implement) this interface.
type SensorData interface {
	print()
}

// Define new data type as itself to implement interface
type TemperatureData float64
type HumidityData float64
type PressureData float64
type LightData int8
type BroadbandData float64

type TimeSeriesData struct {
	Time time.Time
	Data map[byte]SensorData
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

func (temperatureData TemperatureData) print() {
	fmt.Printf("Temperature: %+v(℃)\n", temperatureData)
}

func (humidity HumidityData) print() {
	fmt.Printf("Humidity: %+v(%%)\n", humidity)
}

func (pressure PressureData) print() {
	fmt.Printf("Pressure: %+v(hPa)\n", pressure)
}

func (tiltData *TiltData) print() {
	fmt.Println("Tilt data below")

	fmt.Printf("XAxis: %+v\n", tiltData.XAxis)
	fmt.Printf("YAxis: %+v\n", tiltData.YAxis)
	fmt.Printf("ZAxis: %+v\n", tiltData.ZAxis)
}

func (vibrationData *VibrationData) print() {
	fmt.Println("Vibration data below")

	fmt.Printf("Fund: %+v(Hz)\t", vibrationData.Peak[0])
	fmt.Printf("Amp: %+v\n", vibrationData.Amp[0])

	for i := 1; i < 6; i++ {
		fmt.Printf("Peak%d: %+v(Hz)\t", i+1, vibrationData.Peak[i])
		fmt.Printf("Amp: %+v\n", vibrationData.Amp[i])
	}
}

func (lightData LightData) print() {
	fmt.Printf("Light: %d\n", lightData)
}

func (soundData *SoundData) print() {
	fmt.Println("Sound data below")

	fmt.Printf("Fund: %+v(Hz)\t", soundData.Peak[0])
	fmt.Printf("Amp: %+v\n", soundData.Amp[0])

	for i := 1; i < 6; i++ {
		fmt.Printf("Peak%d: %+v(Hz)\t", i+1, soundData.Peak[i])
		fmt.Printf("Amp: %+v\n", soundData.Amp[i])
	}
}

func (broadbanddata BroadbandData) print() {
	fmt.Printf("Broadband: %+v\n", broadbanddata)
}

// do not use anymore since https://github.com/w00cheol/serial/commit/15d0f2690c37e121818a7f6ab7a93cb38d895186
// type AllData struct {
// 	Temperature TemperatureData
// 	Humidity    HumidityData
// 	Pressure    PressureData
// 	Tilt        *TiltData
// 	VibrationX  *VibrationData
// 	VibrationY  *VibrationData
// 	VibrationZ  *VibrationData
// 	Light       LightData
// 	Sound       *SoundData
// 	Broadband   BroadbandData
// }

// func (allData *AllData) print() {
// 	allData.Temperature.print()
// 	allData.Humidity.print()
// 	allData.Pressure.print()
// 	allData.Tilt.print()
// 	allData.VibrationX.print()
// 	allData.VibrationY.print()
// 	allData.VibrationZ.print()
// 	allData.Light.print()
// 	allData.Sound.print()
// 	allData.Broadband.print()
// }
