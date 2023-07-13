// Define custom error in this file
package main

import "errors"

func InvalidByteLengthError() error {
	return errors.New("Invalid byte length error")
}

func DataMissingError() error {
	return errors.New("Data missing error")
}

func InvalidCommandError() error {
	return errors.New("Invalid command error")
}
