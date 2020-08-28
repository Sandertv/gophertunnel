package protocol

import (
	"bytes"
	"fmt"
)

// WriteString writes a string x to Buffer dst. The string is a slice of bytes prefixed by a varuint32
// specifying its length.
func WriteString(dst *bytes.Buffer, x string) error {
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return fmt.Errorf("%v: error writing string length: %v", callFrame(), err)
	}
	if _, err := dst.WriteString(x); err != nil {
		return fmt.Errorf("%v: error writing string: %v", callFrame(), err)
	}
	return nil
}

// WriteByteSlice writes a []byte x to Buffer dst. The []byte is prefixed by a varuint32 holding its length.
func WriteByteSlice(dst *bytes.Buffer, x []byte) error {
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return fmt.Errorf("%v: error writing []byte length: %v", callFrame(), err)
	}
	if _, err := dst.Write(x); err != nil {
		return fmt.Errorf("%v: error writing []byte: %v", callFrame(), err)
	}
	return nil
}
