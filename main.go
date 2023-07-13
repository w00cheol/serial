package main

import (
	"fmt"
)

func main() {
	// portName must be according to your environment.
	// use "ll /dev/tty*" to see all the serial port.
	d := NewDLPTH1C("/dev/ttyACM0")

	// Make sure to close it later.
	defer d.vcp.Close()

	// change the option to call these functions below.
	// d.set2G()
	// d.set4G()
	// d.set8G()
	// d.set16G()

	// the type of chan must be same as the return type of the function what you call.
	var in chan *TimeSeriesData = make(chan *TimeSeriesData, 1)

	// call and go your function below.
	// GO!!!
	go d.readHumidityAsync(in)

	// recieve data from channel continously
	for timeSeriesData := range in {
		fmt.Printf("Data: %+v\n", timeSeriesData.Data)
		fmt.Printf("Time: %+v\n", timeSeriesData.Time)
		fmt.Println()
	}
}
