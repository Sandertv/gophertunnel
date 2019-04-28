package protocol

import (
	"bytes"
	"fmt"
	"unsafe"
)

// WriteString writes a string x to Buffer dst. The string is a slice of bytes prefixed by a varuint32
// specifying its length.
func WriteString(dst *bytes.Buffer, x string) error {
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return fmt.Errorf("error writing string length: %v", err)
	}
	if _, err := dst.WriteString(x); err != nil {
		return fmt.Errorf("error writing string: %v", err)
	}
	return nil
}

// String reads a string from Buffer src, setting the result to the pointer to a string passed. The string
// read is prefixed by a varuint32.
func String(src *bytes.Buffer, x *string) error {
	var length uint32
	if err := Varuint32(src, &length); err != nil {
		return fmt.Errorf("error reading string length: %v", err)
	}
	data := src.Next(int(length))

	// Use the unsafe package to convert the byte slice to a string without copying.
	*x = *(*string)(unsafe.Pointer(&data))
	return nil
}
