// +build !armbe,!arm64be,!ppc64,!mips,!mips64,!mips64p32,!ppc,!sparc,!sparc64,!s390,!s390x

package protocol

import (
	"encoding/binary"
	"unsafe"
)

// Uint16 writes a little endian uint16 to the underlying buffer.
func (w *Writer) Uint16(x *uint16) {
	data := *(*[2]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}

// Int16 writes a little endian int16 to the underlying buffer.
func (w *Writer) Int16(x *int16) {
	data := *(*[2]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}

// Uint32 writes a little endian uint32 to the underlying buffer.
func (w *Writer) Uint32(x *uint32) {
	data := *(*[4]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}

// Int32 writes a little endian int32 to the underlying buffer.
func (w *Writer) Int32(x *int32) {
	data := *(*[4]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}

// BEInt32 writes a big endian int32 to the underlying buffer.
func (w *Writer) BEInt32(x *int32) {
	w.buf = append(w.buf, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(w.buf[len(w.buf)-4:], uint32(*x))
}

// Uint64 writes a little endian uint64 to the underlying buffer.
func (w *Writer) Uint64(x *uint64) {
	data := *(*[8]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}

// Int64 writes a little endian int64 to the underlying buffer.
func (w *Writer) Int64(x *int64) {
	data := *(*[8]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}

// Float32 writes a little endian float32 to the underlying buffer.
func (w *Writer) Float32(x *float32) {
	data := *(*[4]byte)(unsafe.Pointer(x))
	w.buf = append(w.buf, data[:]...)
}
