package dynamic

import (
	"bytes"
)

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
// Buffer, unlike bytes.Buffer, attempts to shrink the capacity of the Buffer when calling Reset.
type Buffer struct {
	*bytes.Buffer
	c uint16
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *Buffer) Reset() {
	const shrinkAfter = 100
	if b.Len()*4 < b.Cap() {
		b.c++
		if b.c > shrinkAfter {
			b.c = 0
			b.Buffer = bytes.NewBuffer(make([]byte, b.Len()/2))
		}
	} else {
		b.c = 0
	}
	b.Buffer.Reset()
}

// NewBuffer creates and initializes a new Buffer using buf as its
// initial contents. The new Buffer takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a Buffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a Buffer.
func NewBuffer(buf []byte) *Buffer { return &Buffer{Buffer: bytes.NewBuffer(buf)} }

// NewBufferString creates and initializes a new Buffer using string s as its
// initial contents. It is intended to prepare a buffer to read an existing
// string.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a Buffer.
func NewBufferString(s string) *Buffer {
	return &Buffer{Buffer: bytes.NewBuffer([]byte(s))}
}
