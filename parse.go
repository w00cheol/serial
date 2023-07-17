package main

import (
	"log"
	"strconv"
	"strings"
)

func parseTemperature(b string) (TemperatureData, error) {
	sep := strings.Split(b, "= ")
	if len(sep) < 2 {
		return TemperatureData(ParseErrorCodeDLPTH1C), DataMissingError()
	}

	temperatureStr := strings.Split(sep[1], "\xb0C")[0]
	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil {
		log.Print(b)
		return TemperatureData(ParseErrorCodeDLPTH1C), err
	}

	return TemperatureData(temperature), nil
}

func parseHumidity(b string) (HumidityData, error) {
	sep := strings.Split(b, "= ")
	if len(sep) < 2 {
		log.Print(b)
		return HumidityData(ParseErrorCodeDLPTH1C), DataMissingError()
	}

	humidityStr := strings.Split(sep[1], "%")[0]
	humidity, err := strconv.ParseFloat(humidityStr, 64)
	if err != nil {
		log.Print(b)
		return HumidityData(ParseErrorCodeDLPTH1C), err
	}

	return HumidityData(humidity), nil
}

func parsePressure(b string) (PressureData, error) {
	sep := strings.Split(b, "= ")
	if len(sep) < 2 {
		log.Print(b)
		return PressureData(ParseErrorCodeDLPTH1C), DataMissingError()
	}

	pressureStr := strings.TrimSpace(strings.Split(strings.Split(sep[1], "\r")[0], "\x00")[0])
	pressure, err := strconv.ParseFloat(pressureStr, 64)
	if err != nil {
		log.Print(b)
		return PressureData(ParseErrorCodeDLPTH1C), err
	}

	return PressureData(pressure), nil
}

func parseTilt(b string) (tilt *TiltData, err error) {
	// make new tilt for return
	tilt = new(TiltData)

	sep := strings.Split(b, ":")
	if len(sep) < 4 {
		log.Print(b)
		return nil, DataMissingError()
	}

	xAxisStr := strings.TrimSpace(strings.Split(strings.Split(sep[1], " ")[0], "\r")[0])
	yAxisStr := strings.TrimSpace(strings.Split(strings.Split(sep[2], " ")[0], "\r")[0])
	zAxisStr := strings.TrimSpace(strings.Split(strings.Split(sep[3], " ")[0], "\r")[0])

	xAxis, err := strconv.ParseInt(xAxisStr, 10, 64)
	if err != nil {
		log.Print(b)
		return nil, err
	}
	yAxis, err := strconv.ParseInt(yAxisStr, 10, 64)
	if err != nil {
		log.Print(b)
		return nil, err
	}
	zAxis, err := strconv.ParseInt(zAxisStr, 10, 64)
	if err != nil {
		log.Print(b)
		return nil, err
	}

	// assign into struct's member value
	tilt.XAxis = xAxis
	tilt.YAxis = yAxis
	tilt.ZAxis = zAxis

	return tilt, nil
}

func parseVibration(cmd byte, b string) (vibration *VibrationData, err error) {
	// string parsing
	lines := strings.Split(b, "\n")
	if len(lines) < 7 {
		log.Print(b)
		return nil, DataMissingError()
	}

	// create new VibrationData pointer type variable
	vibration = new(VibrationData)

	// Ignore few lines('\n') by filter (explained below).
	// If the length is not 3 when split based on the letter ':', it will be considered a meaningless line.
	// (DLP-TH1C returns meaningless byte '\n' within a response of vibration request)
	i := 0
	for _, line := range lines {
		// string parsing
		sep := strings.Split(line, ":")
		if len(sep) != 3 {
			continue
		}

		peakStr := strings.Split(sep[1], "Hz")[0]
		peakStr = strings.TrimLeft(peakStr, " ")
		peak, err := strconv.ParseInt(peakStr, 10, 64)
		if err != nil {
			log.Print(b)
			return nil, err
		}

		ampStr := strings.Split(strings.Split(sep[2], "\r")[0], "\x00")[0]
		amp, err := strconv.ParseFloat(ampStr, 64)
		if err != nil {
			log.Print(b)
			return nil, err
		}

		// set value into response struct
		vibration.Axis = cmd
		vibration.Peak[i] = peak
		vibration.Amp[i] = amp

		i++
		if i > 5 {
			return vibration, nil
		}
	}

	log.Print(b)
	return nil, DataMissingError()
}

func parseLight(b string) (LightData, error) {
	sep := strings.Split(b, ": ")
	if len(sep) < 2 {
		log.Print(b)
		return LightData(ParseErrorCodeDLPTH1C), DataMissingError()
	}

	lightStr := strings.Split(strings.Split(strings.Split(sep[1], "\r")[0], "\n")[0], "\x00")[0]
	light64, err := strconv.ParseInt(lightStr, 10, 8)
	if err != nil {
		log.Print(b)
		return LightData(ParseErrorCodeDLPTH1C), err
	}

	// light value consists of 8 bits (according to dlpdesing.com that made DLP-TH1C)
	return LightData(light64), nil
}

func parseSound(b string) (sound *SoundData, err error) {
	sound = new(SoundData)

	lines := strings.Split(b, "\n")
	if len(lines) < 7 {
		log.Print(b)
		return nil, DataMissingError()
	}

	// Ignore few lines('\n') by filter (explained below).
	// If the length is not 3 when split based on the letter ':', it will be considered a meaningless line.
	// (DLP-TH1C returns meaningless byte '\n' within a response of vibration request)
	i := 0
	for _, line := range lines {
		sep := strings.Split(line, ":")
		if len(sep) != 3 {
			continue
		}

		peakStr := strings.Split(sep[1], "Hz")[0]
		peakStr = strings.TrimLeft(peakStr, " ")
		peak, err := strconv.ParseInt(peakStr, 10, 64)
		if err != nil {
			log.Print(b)
			return nil, err
		}

		ampStr := strings.Split(strings.Split(sep[2], "\r")[0], "\x00")[0]
		amp, err := strconv.ParseFloat(ampStr, 64)
		if err != nil {
			log.Print(b)
			return nil, err
		}

		// set value into the response struct
		sound.Peak[i] = peak
		sound.Amp[i] = amp

		i++
		if i > 5 {
			return sound, nil
		}
	}

	log.Print(b)
	return nil, DataMissingError()
}

func parseBroadband(b string) (BroadbandData, error) {
	sep := strings.Split(b, ": ")
	if len(sep) < 2 {
		return BroadbandData(ParseErrorCodeDLPTH1C), DataMissingError()
	}

	broadbandStr := strings.Split(strings.Split(strings.Split(sep[1], "\r")[0], "\n")[0], "\x00")[0]
	broadband, err := strconv.ParseFloat(broadbandStr, 64)
	if err != nil {
		log.Print(b)
		return BroadbandData(ParseErrorCodeDLPTH1C), err
	}

	return BroadbandData(broadband), nil
}
