package protocol

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// chainErr chains together a variadic amount of errors into a single error and returns it. If all errors
// passed are nil, the error returned will also be nil.
func chainErr(err ...error) error {
	var msg string
	hasEOF := true
	for _, e := range err {
		if e == nil {
			continue
		}
		if strings.Contains(msg, "EOF") {
			if hasEOF {
				// No need to log multiple EOFs.
				continue
			}
			hasEOF = true
		}
		msg += wrap(e).Error() + "\n"
	}
	if msg == "" {
		return nil
	}
	return errors.New(strings.TrimRight(msg, "\n"))
}

// callFrame obtains a call frame and formats it so that it includes the file, function and line.
func callFrame() string {
	targetFrameIndex := 3

	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}
	return fmt.Sprintf("%v/%v", frame.Function, frame.Line)
}

// wrap wraps a callframe around an error and returns the new error.
func wrap(e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("%v: %v", callFrame(), e)
}
