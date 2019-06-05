package protocol

import "fmt"

const lowerLimit = 64
const mediumLimit = 256
const higherLimit = 1024

// LimitHitError is returned by a reading operation if it hits the limit of the maximum amount of elements in
// an array.
type LimitHitError struct {
	Limit int
	Type  string
}

// Error ...
func (err LimitHitError) Error() string {
	return wrap(fmt.Errorf("maximum element count %v hit for type '%v'", err.Limit, err.Type)).Error()
}

// NegativeCountError is returned when a count prefix of an array-type structure is a negative number. Most
// types only have unsigned count prefixes, but some do not and may return this error.
type NegativeCountError struct {
	Type string
}

// Error ...
func (err NegativeCountError) Error() string {
	return wrap(fmt.Errorf("invalid negative count prefix for type '%v'", err.Type)).Error()
}
