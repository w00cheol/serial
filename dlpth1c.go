// Provides functions communicating with the DLPTH1C sensor.
// Provides only functions that communicate using ascii.
// In my environment, using byte, the dlp-th1c sensor loses some data for some reason (but I couldn't find).
// The "//string parsing" parts in various parts of the function were also written considering data loss.
package main

import (
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type DLPTH1C struct {
	portName string
	vcp      io.ReadWriteCloser
}

func NewDLPTH1C(portName string) *DLPTH1C {
	// Set up options.
	options := serial.OpenOptions{
		PortName:        portName,
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 8,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	return &DLPTH1C{portName: portName, vcp: port}
}

func (d *DLPTH1C) readAllAsync(out chan<- *TimeSeriesData) error {
	allData := new(AllData)
	for {
		// request all value in ascii code
		// ([]byte{'t','h','p','a','x','v','w','l','f','b',})
		d.vcp.Write(
			[]byte{
				TemperatureASCIICmd,
				HumidityASCIICmd,
				PressureASCIICmd,
				TiltASCIICmd,
				VibrationXASCIICmd,
				VibrationYASCIICmd,
				VibrationZASCIICmd,
				LightASCIICmd,
				SoundASCIICmd,
				BroadbandASCIICmd,
			})

		// need to wait response for 30 seconds because of data loss issue
		t := time.Now()
		time.Sleep(30 * time.Second)

		// read from response
		b := make([]byte, 2048)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing (temperature)
		sep := strings.Split(string(b), "= ")
		if len(sep) < 3 {
			DataMissingError()
		}
		temperatureStr := strings.Split(sep[1], "\xb0C")[0]
		temperature, err := strconv.ParseFloat(temperatureStr, 64)
		if err != nil {
			DataMissingError()
		}
		allData.Temperature = temperature

		// string parsing (humidity)
		humidityStr := strings.Split(sep[2], "%")[0]
		humidity, err := strconv.ParseFloat(humidityStr, 64)
		if err != nil {
			return err
		}
		allData.Humidity = humidity

		out <- &(TimeSeriesData{Time: t, Data: allData})
	}
}

func (d *DLPTH1C) readTemperatureAsync(out chan<- *TimeSeriesData) error {
	for {
		// request temperature value in ascii code
		d.vcp.Write([]byte{TemperatureASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		// read from response
		b := make([]byte, 64)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		temperature, err := parseTemperature(b)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: temperature})
	}
}

func (d *DLPTH1C) readHumidityAsync(out chan<- *TimeSeriesData) error {
	for {
		// request humidity value in ascii code
		d.vcp.Write([]byte{HumidityASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 2048)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		humidity, err := parseHumidity(b)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: humidity})
	}
}

func (d *DLPTH1C) readPressureAsync(out chan<- *TimeSeriesData) error {
	for {
		// request pressure value in ascii code
		d.vcp.Write([]byte{PressureASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 64)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		pressure, err := parsePressure(b)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: pressure})
	}
}

func (d *DLPTH1C) readTiltAsync(out chan<- *TimeSeriesData) error {
	for {
		// request tilt value in ascii code
		d.vcp.Write([]byte{TiltASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 64)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		tilt, err := parseTilt(b)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: tilt})
	}
}

// this funcion requires certain command for specify axis
// please check ./cmd.go
func (d *DLPTH1C) readVibrationAsync(cmd byte, out chan<- *TimeSeriesData) error {
	for {
		// request vibration value in ascii code
		// the command must be the one of 3 axis command
		if cmd != VibrationXASCIICmd && cmd != VibrationYASCIICmd && cmd != VibrationZASCIICmd {
			return InvalidCommandError()
		}

		d.vcp.Write([]byte{cmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 256)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		vibration, err := parseVibration(b)
		if err != nil {
			return err
		}

		// it goes out to the channel
		out <- &(TimeSeriesData{Time: t, Data: vibration})
	}
}

func (d *DLPTH1C) readLightAsync(out chan<- *TimeSeriesData) error {
	for {
		// request light value in ascii code
		d.vcp.Write([]byte{LightASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 32)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		light, err := parseLight(b)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: light})
	}
}

func (d *DLPTH1C) readSoundAsync(out chan<- *TimeSeriesData) error {
	for {
		// request sound value in ascii code
		d.vcp.Write([]byte{SoundASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 256)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		sound, err := parseSound(b)
		if err != nil {
			return err
		}

		// it goes out to the channel
		out <- &TimeSeriesData{Time: t, Data: sound}
	}
}

func (d *DLPTH1C) readBroadbandAsync(out chan<- *TimeSeriesData) error {
	for {
		// request sound value in ascii code
		d.vcp.Write([]byte{BroadbandASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 32)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		broadband, err := parseBroadband(b)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: broadband})
	}
}

func (d *DLPTH1C) set2G() error {
	d.vcp.Write([]byte{Set2GASCIICmd})
	time.Sleep(2 * time.Second)

	b := make([]byte, 4)
	if _, err := d.vcp.Read(b); err != nil {
		return err
	}

	return nil
}

func (d *DLPTH1C) set4G() error {
	d.vcp.Write([]byte{Set4GASCIICmd})
	time.Sleep(2 * time.Second)

	b := make([]byte, 4)
	if _, err := d.vcp.Read(b); err != nil {
		return err
	}

	return nil
}

func (d *DLPTH1C) set8G() error {
	d.vcp.Write([]byte{Set8GASCIICmd})
	time.Sleep(2 * time.Second)

	b := make([]byte, 4)
	if _, err := d.vcp.Read(b); err != nil {
		return err
	}

	return nil
}

func (d *DLPTH1C) set16G() error {
	d.vcp.Write([]byte{Set16GASCIICmd})
	time.Sleep(2 * time.Second)

	b := make([]byte, 4)
	if _, err := d.vcp.Read(b); err != nil {
		return err
	}

	return nil
}

func bitwiseOR2Bytes(b []byte) (uint16, error) {
	if len(b) != 2 {
		return 0, InvalidByteLengthError()
	}

	return uint16(b[1])<<8 | uint16(b[0]), nil
}

func bitwiseOR3Bytes(b []byte) (uint32, error) {
	if len(b) != 3 {
		return 0, InvalidByteLengthError()
	}

	return uint32(b[2])<<16 | uint32(b[1])<<8 | uint32(b[0]), nil
}

func bitwiseOR4Bytes(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, InvalidByteLengthError()
	}

	return uint32(b[3])<<24 | uint32(b[2])<<16 | uint32(b[1])<<8 | uint32(b[0]), nil
}
