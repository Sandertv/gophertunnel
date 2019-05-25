package nbt

import "io"

// offsetWriter is a wrapper around an io.Writer which keeps track of the amount of bytes written, so that it
// may be used in errors.
type offsetWriter struct {
	io.Writer
	off int64

	// WriteByte is a function implemented by offsetWriter if the io.Writer does not implement it itself.
	WriteByte func(byte) error
}

// Write writes a byte slice to the underlying io.Writer. It increases the byte offset by exactly n.
func (w *offsetWriter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	w.off += int64(n)
	return
}
