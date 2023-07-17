// Define command constant for request in this file
package main

// It will be wrapped as return type when parse error occurs
const ParseErrorCodeDLPTH1C = -1

type BinaryCmd byte
type ASCIICmd byte

// Binary request command
const (
	PingBinaryCmd        byte = 0x22
	TemperatureBinaryCmd byte = 0x54
	HumidityBinaryCmd    byte = 0x48
	PressureBinaryCmd    byte = 0x50
	TiltBinaryCmd        byte = 0x41
	VibrationXBinaryCmd  byte = 0x58
	VibrationYBinaryCmd  byte = 0x56
	VibrationZBinaryCmd  byte = 0x57
	LightBinaryCmd       byte = 0x4C
	SoundBinaryCmd       byte = 0x46
	BroadBinaryCmd       byte = 0x42
	Set2GBinaryCmd       byte = 0x6d
	Set4GBinaryCmd       byte = 0x6E
	Set8GBinaryCmd       byte = 0x2C
	Set16GBinaryCmd      byte = 0x2E
)

// Divisor for calculating from Binary value (, which is response from sensor)
const (
	TemperatureDivisor float64 = 100.0
	HumidityDivisor    float64 = 1024.0
	PressureDivisor    float64 = 25600.0
)

// ASCII request command
const (
	PingASCIICmd        byte = 0x27 // '''
	HelpASCIICmd        byte = 0x3f // '?'
	TemperatureASCIICmd byte = 0x74 // 't'
	HumidityASCIICmd    byte = 0x68 // 'h'
	PressureASCIICmd    byte = 0x70 // 'p'
	TiltASCIICmd        byte = 0x61 // 'a'
	VibrationXASCIICmd  byte = 0x78 // 'x'
	VibrationYASCIICmd  byte = 0x76 // 'v'
	VibrationZASCIICmd  byte = 0x77 // 'w'
	LightASCIICmd       byte = 0x6C // 'l'
	SoundASCIICmd       byte = 0x66 // 'f'
	BroadbandASCIICmd   byte = 0x62 // 'b'
	Set2GASCIICmd       byte = 0x6D // 'm'
	Set4GASCIICmd       byte = 0x6E // 'n'
	Set8GASCIICmd       byte = 0x2C // ','
	Set16GASCIICmd      byte = 0x2E // '.'
)
