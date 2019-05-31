package protocol

import (
	"fmt"
	"runtime"
)

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
	return fmt.Errorf("%v: %v", callFrame(), e)
}
