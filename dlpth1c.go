// Provides functions communicating with the DLPTH1C sensor.
// Provides only functions that communicate using ascii.
// In my environment, using byte, the dlp-th1c sensor loses some data for some reason (but I couldn't find).
// The "//string parsing" parts in various parts of the function were also written considering data loss.
package serial

import (
	"io"
	"log"
	"strings"
	"sync"
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
		PortName:              portName,
		BaudRate:              115200,
		DataBits:              8,
		StopBits:              1,
		InterCharacterTimeout: 1000,
		MinimumReadSize:       0,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	return &DLPTH1C{portName: portName, vcp: port}
}

func (d *DLPTH1C) readAllAsync(out chan<- *TimeSeriesData) error {
	const numParsingFuntion int = 10
	const numMapWriteFunction int = 1
	var wg sync.WaitGroup

	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// Assign return value
		allData := map[byte]SensorData{}

		// Get time
		t := time.Now()

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

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}
			b = append(b, buff...)
		}

		sep := strings.Split(string(b), "\n")
		if len(sep) < 35 {
			log.Print("Data Missing.")
			log.Print("Try again.")
			continue
		}

		// discard few lines at first for parsing
		i := 0
		for sep[i] == "" {
			i++
		}

		wg.Add(numParsingFuntion)
		wg.Add(numMapWriteFunction)

		// https://go.dev/doc/faq#atomic_maps
		// Map write fuction to avoid "concurrent map writes error"
		// "Map operations" not defined to be atomic,
		// so it is decided that use channel to store data.
		var mapWriteChan chan *MapWriteData = make(chan *MapWriteData)
		go func() {
			defer wg.Done()
			defer close(mapWriteChan)

			for i := 0; i < numParsingFuntion; i++ {
				writeData := <-mapWriteChan
				allData[writeData.Key] = writeData.Value
			}
		}()

		// string parsing (temperature)
		go func() {
			defer wg.Done()

			temperature, err := parseTemperature(sep[i])
			if err != nil {
				log.Print(err)
			}
			mapWriteChan <- &MapWriteData{TemperatureASCIICmd, temperature}
		}()

		// string parsing (humidity)
		go func() {
			defer wg.Done()

			humidity, err := parseHumidity(sep[i+1])
			if err != nil {
				log.Print(err)
			}
			mapWriteChan <- &MapWriteData{HumidityASCIICmd, humidity}
		}()

		// string parsing (pressure)
		go func() {
			defer wg.Done()

			pressure, err := parsePressure(sep[i+2])
			if err != nil {
				log.Print(err)
			}
			mapWriteChan <- &MapWriteData{PressureASCIICmd, pressure}
		}()

		// string parsing (tilt)
		go func() {
			defer wg.Done()

			tilt, err := parseTilt(sep[i+3])
			if err != nil {
				log.Print(err)
			}
			mapWriteChan <- &MapWriteData{TiltASCIICmd, tilt}
		}()

		// string parsing (vibration X, Y, Z in order)
		go func() {
			defer wg.Done()

			vibrationX, err := parseVibration(VibrationXASCIICmd, strings.Join(sep[i+4:i+11], "\n"))
			if err != nil {
				log.Print(err)
			}

			mapWriteChan <- &MapWriteData{VibrationXASCIICmd, vibrationX}
		}()

		go func() {
			defer wg.Done()

			vibrationY, err := parseVibration(VibrationYASCIICmd, strings.Join(sep[i+11:i+18], "\n"))
			if err != nil {
				log.Print(err)
			}

			mapWriteChan <- &MapWriteData{VibrationYASCIICmd, vibrationY}
		}()

		go func() {
			defer wg.Done()

			vibrationZ, err := parseVibration(VibrationZASCIICmd, strings.Join(sep[i+18:i+25], "\n"))
			if err != nil {
				log.Print(err)
			}

			mapWriteChan <- &MapWriteData{VibrationZASCIICmd, vibrationZ}
		}()

		// string parsing (light)
		go func() {
			defer wg.Done()

			light, err := parseLight(sep[i+25])
			if err != nil {
				log.Print(err)
			}

			mapWriteChan <- &MapWriteData{LightASCIICmd, light}
		}()

		// string parsing (sound)
		go func() {
			defer wg.Done()

			sound, err := parseSound(strings.Join(sep[i+26:i+33], "\n"))
			if err != nil {
				log.Print(err)
			}

			mapWriteChan <- &MapWriteData{SoundASCIICmd, sound}
		}()

		// string parsing (broadband)
		go func() {
			defer wg.Done()

			broadband, err := parseBroadband(sep[i+33])
			if err != nil {
				log.Print(err)
			}

			mapWriteChan <- &MapWriteData{BroadbandASCIICmd, broadband}
		}()

		wg.Wait()
		out <- &(TimeSeriesData{Time: t, Data: allData})
	}
}

func (d *DLPTH1C) readTemperatureAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{TemperatureASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		temperature, err := parseTemperature(string(b))
		if err != nil {
			log.Fatal(err)
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[TemperatureASCIICmd] = temperature
		out <- result
	}
}

func (d *DLPTH1C) readHumidityAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{HumidityASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		humidity, err := parseHumidity(string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[HumidityASCIICmd] = humidity
		out <- result
	}
}

func (d *DLPTH1C) readPressureAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{PressureASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		pressure, err := parsePressure(string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[PressureASCIICmd] = pressure
		out <- result
	}
}

func (d *DLPTH1C) readTiltAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{TiltASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		tilt, err := parseTilt(string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[TiltASCIICmd] = tilt
		out <- result
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

		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{cmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		vibration, err := parseVibration(cmd, string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel

		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[cmd] = vibration
		out <- result
	}
}

func (d *DLPTH1C) readLightAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{LightASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		light, err := parseLight(string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[LightASCIICmd] = light
		out <- result
	}
}

func (d *DLPTH1C) readSoundAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{SoundASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		sound, err := parseSound(string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[SoundASCIICmd] = sound
		out <- result
	}
}

func (d *DLPTH1C) readBroadbandAsync(out chan<- *TimeSeriesData) error {
	for {
		// Assign byte array that will be given the sensor data
		b := make([]byte, 0)
		// Assigns 1 byte array where bytes from the port will be stored
		buff := make([]byte, 1)

		// request temperature value in ascii code
		d.vcp.Write([]byte{BroadbandASCIICmd})
		t := time.Now()

		// Read from response
		for {
			n, err := d.vcp.Read(buff)
			if err != nil {
				if err == io.EOF || n == 0 {
					break
				}
				log.Fatal(err)
			}

			b = append(b, buff...)
		}

		// string parsing
		broadband, err := parseBroadband(string(b))
		if err != nil {
			return err
		}

		// it goes out to the channel
		result := new(TimeSeriesData)
		result.Time = t
		result.Data = make(map[byte]SensorData)
		result.Data[BroadbandASCIICmd] = broadband
		out <- result
	}
}

func (d *DLPTH1C) set2G() error {
	d.vcp.Write([]byte{Set2GASCIICmd})

	// Assigns 1 byte array where bytes from the port will be stored
	buff := make([]byte, 1)

	// Clear buffer
	for {
		n, err := d.vcp.Read(buff)
		if err != nil {
			if err == io.EOF || n == 0 {
				break
			}
			log.Fatal(err)
		}
	}

	return nil
}

func (d *DLPTH1C) set4G() error {
	d.vcp.Write([]byte{Set4GASCIICmd})

	// Assigns 1 byte array where bytes from the port will be stored
	buff := make([]byte, 1)

	// Clear buffer
	for {
		n, err := d.vcp.Read(buff)
		if err != nil {
			if err == io.EOF || n == 0 {
				break
			}
			log.Fatal(err)
		}
	}

	return nil
}

func (d *DLPTH1C) set8G() error {
	d.vcp.Write([]byte{Set8GASCIICmd})

	// Assigns 1 byte array where bytes from the port will be stored
	buff := make([]byte, 1)

	// Clear buffer
	for {
		n, err := d.vcp.Read(buff)
		if err != nil {
			if err == io.EOF || n == 0 {
				break
			}
			log.Fatal(err)
		}
	}

	return nil
}

func (d *DLPTH1C) set16G() error {
	d.vcp.Write([]byte{Set16GASCIICmd})

	// Assigns 1 byte array where bytes from the port will be stored
	buff := make([]byte, 1)

	// Clear buffer
	for {
		n, err := d.vcp.Read(buff)
		if err != nil {
			if err == io.EOF || n == 0 {
				break
			}
			log.Fatal(err)
		}
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
