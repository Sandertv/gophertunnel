package protocol

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
)

// UUID reads a little endian UUID from buffer src into UUID id.
func UUID(src *bytes.Buffer, id *uuid.UUID) error {
	b := make([]byte, 16)
	if _, err := src.Read(b); err != nil {
		return fmt.Errorf("%v: need exactly 16 bytes to decode a UUID", callFrame())
	}
	*id = uuid.UUID(reverseUUIDBytes(b))
	return nil
}

// WriteUUID writes a little endian UUID id to buffer dst.
func WriteUUID(dst *bytes.Buffer, id uuid.UUID) error {
	b := reverseUUIDBytes(id[:])
	_, err := dst.Write(b[:])
	if err != nil {
		return wrap(err)
	}
	return nil
}

// reverseUUIDBytes reverses the 16 bytes that a UUID exists out of, so that it is written in little endian.
// This means first swapping the order of the two int64s, and after that reversing all bytes.
func reverseUUIDBytes(b []byte) [16]byte {
	b = append(b[8:], b[:8]...)
	var arr [16]byte
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = b[j], b[i]
	}
	return arr
}
