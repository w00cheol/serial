package serial

import (
	"fmt"
	"log"
)

// set custom option value as false
var option bool = false
var port string = "/dev/ttyACM0"

func usage() {
	fmt.Printf("\n===============================================================\n")
	fmt.Printf("USAGE: run with COMMAND\n")
	fmt.Printf("all:\t\t\t\tRead All Data\n")
	fmt.Printf("(COMBINE BELOW COMMANDS):\tRead Costomized Data but takes 30 secs\n")
	fmt.Printf("t:\t\t\t\tRead Temperature Data Only\n")
	fmt.Printf("h:\t\t\t\tRead Humidity Data Only\n")
	fmt.Printf("p:\t\t\t\tRead Pressure Data Only\n")
	fmt.Printf("a:\t\t\t\tRead Tilt Data Only\n")
	fmt.Printf("x:\t\t\t\tRead Vibration (X Axis) Data Only\n")
	fmt.Printf("v:\t\t\t\tRead Vibration (Y Axis) Data Only\n")
	fmt.Printf("w:\t\t\t\tRead Vibration (Z Axis) Data Only\n")
	fmt.Printf("l:\t\t\t\tRead Light Level Data Only\n")
	fmt.Printf("f:\t\t\t\tRead Sound Data Only\n")
	fmt.Printf("b:\t\t\t\tRead Broadband Data Only\n")
	fmt.Printf("===============================================================\n\n")
}

func InitPort(initPort string) {
	port = initPort
}

func RunWithCommand(cmd string) {
	// portName must be according to your environment.
	// use "ll /dev/tty*" to see all the serial port.
	d := NewDLPTH1C(port)

	// Make sure to close it later.
	defer d.vcp.Close()

	// change the option to call these functions below.
	// d.set2G()
	// d.set4G()
	// d.set8G()
	// d.set16G()

	// the type of chan must be same as the return type of the function what you call.
	var in chan *TimeSeriesData = make(chan *TimeSeriesData)
	defer close(in)

	if len(cmd) == 0 {
		usage()
		return

	} else if len(cmd) > 10 {
		log.Fatal("TOO MANY ARGUMENTS...")

	} else {
		if cmd == "all" {
			// Read all
			go d.readAllAsync(in)

		} else if len(cmd) == 1 {
			// Select one function by the command that had been selected by user.
			switch cmd {
			case string(TemperatureASCIICmd):
				go d.readTemperatureAsync(in)

			case string(HumidityASCIICmd):
				go d.readHumidityAsync(in)

			case string(PressureASCIICmd):
				go d.readPressureAsync(in)

			case string(TiltASCIICmd):
				go d.readTiltAsync(in)

			case string(VibrationXASCIICmd):
				go d.readVibrationAsync(VibrationXASCIICmd, in)

			case string(VibrationYASCIICmd):
				go d.readVibrationAsync(VibrationYASCIICmd, in)

			case string(VibrationZASCIICmd):
				go d.readVibrationAsync(VibrationZASCIICmd, in)

			case string(LightASCIICmd):
				go d.readLightAsync(in)

			case string(SoundASCIICmd):
				go d.readSoundAsync(in)

			case string(BroadbandASCIICmd):
				go d.readBroadbandAsync(in)

			default:
				usage()
				return
			}

		} else {
			// Select function by the command that had been selected by user.
			// It needs to be updated to use onther parsing function depending on each option.
			// But imagine how the custom option could be,
			// The value of the sensor responds is too variable to expect every kind of format, exception, data loss as well.
			// So it is decided to call readAllAsync(chan) function and just extract only the kind of data that user wants.
			option = true
			go d.readAllAsync(in)
		}
	}

	// Recieve data from channel continously
	// If option flag is true, which means user wants recieve only certain kind of data
	if option {
		for timeSeriesData := range in {
			for _, c := range []byte(cmd) {
				// data extracting
				if _, exist := timeSeriesData.Data[c]; !exist {
					log.Fatalf("%+v IS A WRONG ARGUMENT...", string(c))
				}

				timeSeriesData.Data[c].print()
			}
			fmt.Printf("Time: %+v\n\n", timeSeriesData.Time)
		}
	} else {
		// It will be called when user executes "all" command or only 1 kind of data command
		for timeSeriesData := range in {
			for _, data := range timeSeriesData.Data {
				data.print()
			}

			fmt.Printf("Time: %+v\n\n", timeSeriesData.Time)
		}
	}
}
