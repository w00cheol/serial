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
		sep := strings.Split(string(b), "= ")
		if len(sep) < 2 {
			continue
		}
		temperatureStr := strings.Split(sep[1], "\xb0C")[0]
		if len(temperatureStr) > 5 {
			continue
		}
		temperature, err := strconv.ParseFloat(temperatureStr, 64)
		if err != nil {
			continue
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

		b := make([]byte, 128)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		sep := strings.Split(string(b), " = ")
		if len(sep) < 2 {
			return DataMissingError()
		}
		humidityStr := strings.Split(sep[1], "%")[0]
		if len(humidityStr) > 5 {
			return DataMissingError()
		}
		humidity, err := strconv.ParseFloat(humidityStr, 64)
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
		sep := strings.Split(string(b), "= ")
		if len(sep) < 2 {
			return DataMissingError()
		}
		pressureStr := strings.TrimSpace(strings.Split(strings.Split(sep[1], "\r")[0], "\x00")[0])
		pressure, err := strconv.ParseFloat(pressureStr, 64)
		if err != nil {
			return err
		}

		out <- &(TimeSeriesData{Time: t, Data: pressure})
	}
}

func (d *DLPTH1C) readTiltAsync(out chan<- *TimeSeriesData) error {
	for {
		// make new tilt for return
		tilt := new(TiltData)

		// request tilt value in ascii code
		d.vcp.Write([]byte{TiltASCIICmd})
		t := time.Now()
		time.Sleep(2 * time.Second)

		b := make([]byte, 64)
		if _, err := d.vcp.Read(b); err != nil {
			return err
		}

		// string parsing
		sep := strings.Split(string(b), ":")
		if len(sep) < 4 {
			return DataMissingError()
		}

		xAxisStr := strings.TrimSpace(strings.Split(strings.Split(sep[1], " ")[0], "\r")[0])
		yAxisStr := strings.TrimSpace(strings.Split(strings.Split(sep[2], " ")[0], "\r")[0])
		zAxisStr := strings.TrimSpace(strings.Split(strings.Split(sep[3], " ")[0], "\r")[0])

		xAxis, err := strconv.ParseInt(xAxisStr, 10, 64)
		if err != nil {
			return err
		}
		yAxis, err := strconv.ParseInt(yAxisStr, 10, 64)
		if err != nil {
			return err
		}
		zAxis, err := strconv.ParseInt(zAxisStr, 10, 64)
		if err != nil {
			return err
		}

		// assign into struct's member value
		tilt.XAxis = xAxis
		tilt.YAxis = yAxis
		tilt.ZAxis = zAxis

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
		lines := strings.Split(string(b), "\n")
		if len(lines) < 8 {
			return DataMissingError()
		}

		vibration := new(VibrationData)

		// ignore two lines('\n') at first and the last from after line 8.
		// (DLP-TH1C returns meaningless byte '\n' within a response of vibration request)
		for i, line := range lines[2:8] {
			// string parsing
			sep := strings.Split(line, ":")
			if len(sep) != 3 {
				return DataMissingError()
			}

			peakStr := strings.Split(sep[1], "Hz")[0]
			peakStr = strings.TrimLeft(peakStr, " ")
			peak, err := strconv.ParseInt(peakStr, 10, 64)
			if err != nil {
				return err
			}

			ampStr := strings.Split(strings.Split(sep[2], "\r")[0], "\x00")[0]
			amp, err := strconv.ParseFloat(ampStr, 64)
			if err != nil {
				return err
			}

			// set value into response struct
			vibration.Peak[i] = peak
			vibration.Amp[i] = amp

			// it goes out to the channel
			out <- &(TimeSeriesData{Time: t, Data: vibration})
		}
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
		sep := strings.Split(string(b), ": ")
		if len(sep) < 2 {
			return DataMissingError()
		}

		lightStr := strings.Split(strings.Split(strings.Split(sep[1], "\r")[0], "\n")[0], "\x00")[0]

		light64, err := strconv.ParseInt(lightStr, 10, 8)
		if err != nil {
			return err
		}

		// light value consists of 8 bits (according to dlpdesing.com that made DLP-TH1C)
		light := int8(light64)

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
		lines := strings.Split(string(b), "\n")
		if len(lines) < 9 {
			return DataMissingError()
		}

		sound := new(SoundData)

		// ignore two lines at first and the last after line 8.
		// (DLP-TH1C returns meaningless byte "\n" within a result of sound value)
		for i, line := range lines[2:8] {
			sep := strings.Split(line, ":")
			if len(sep) != 3 {
				return DataMissingError()
			}

			peakStr := strings.Split(sep[1], "Hz")[0]
			peakStr = strings.TrimLeft(peakStr, " ")
			peak, err := strconv.ParseInt(peakStr, 10, 64)
			if err != nil {
				return err
			}

			ampStr := strings.Split(strings.Split(sep[2], "\r")[0], "\x00")[0]
			amp, err := strconv.ParseFloat(ampStr, 64)
			if err != nil {
				return err
			}

			// set value into the response struct
			sound.Peak[i] = peak
			sound.Amp[i] = amp

			// it goes out to the channel
			out <- &(TimeSeriesData{Time: t, Data: sound})
		}
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
		sep := strings.Split(string(b), ": ")
		if len(sep) < 2 {
			return DataMissingError()
		}

		broadbandStr := strings.Split(strings.Split(strings.Split(sep[1], "\r")[0], "\n")[0], "\x00")[0]

		broadband, err := strconv.ParseFloat(broadbandStr, 64)
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
