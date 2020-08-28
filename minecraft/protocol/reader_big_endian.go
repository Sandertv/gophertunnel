// +build armbe arm64be ppc64 mips mips64 mips64p32 ppc sparc sparc64 s390 s390x

package protocol

import (
	"encoding/binary"
	"io"
	"math"
)

// Uint16 reads a little endian uint16 from the underlying buffer.
func (r *Reader) Uint16(x *uint16) {
	if r.Len() < 2 {
		r.panic(io.EOF)
	}
	*x = binary.LittleEndian.Uint16(r.buf[r.off:])
	r.off += 2
}

// Int16 reads a little endian int16 from the underlying buffer.
func (r *Reader) Int16(x *int16) {
	if r.Len() < 2 {
		r.panic(io.EOF)
	}
	*x = int16(binary.LittleEndian.Uint16(r.buf[r.off:]))
	r.off += 2
}

// Uint32 reads a little endian uint32 from the underlying buffer.
func (r *Reader) Uint32(x *uint32) {
	if r.Len() < 4 {
		r.panic(io.EOF)
	}
	*x = binary.LittleEndian.Uint32(r.buf[r.off:])
	r.off += 4
}

// Int32 reads a little endian int32 from the underlying buffer.
func (r *Reader) Int32(x *int32) {
	if r.Len() < 4 {
		r.panic(io.EOF)
	}
	*x = int32(binary.LittleEndian.Uint32(r.buf[r.off:]))
	r.off += 4
}

// BEInt32 reads a big endian int32 from the underlying buffer.
func (r *Reader) BEInt32(x *int32) {
	if r.Len() < 4 {
		r.panic(io.EOF)
	}
	*x = *(*int32)(unsafe.Pointer(&r.buf[r.off]))
	r.off += 4
}

// Uint64 reads a little endian uint64 from the underlying buffer.
func (r *Reader) Uint64(x *uint64) {
	if r.Len() < 8 {
		r.panic(io.EOF)
	}
	*x = binary.LittleEndian.Uint64(r.buf[r.off:])
	r.off += 8
}

// Int64 reads a little endian int64 from the underlying buffer.
func (r *Reader) Int64(x *int64) {
	if r.Len() < 8 {
		r.panic(io.EOF)
	}
	*x = int64(binary.LittleEndian.Uint64(r.buf[r.off:]))
	r.off += 8
}

// Float32 reads a little endian float32 from the underlying buffer.
func (r *Reader) Float32(x *float32) {
	if r.Len() < 4 {
		r.panic(io.EOF)
	}
	*x = math.Float32frombits(binary.LittleEndian.Uint32(r.buf[r.off:]))
	r.off += 4
}
