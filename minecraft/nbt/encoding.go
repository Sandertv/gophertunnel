package nbt

import (
	"errors"
	"math"
	"unsafe"
)

// Encoding is an encoding variant of NBT. In general, there are three different encodings of NBT, which are
// all the same except for the way basic types are written.
type Encoding interface {
	Int16(r *offsetReader) (int16, error)
	Int32(r *offsetReader) (int32, error)
	Int64(r *offsetReader) (int64, error)
	Float32(r *offsetReader) (float32, error)
	Float64(r *offsetReader) (float64, error)
	String(r *offsetReader) (string, error)
	Int32Slice(r *offsetReader) ([]int32, error)
	Int64Slice(r *offsetReader) ([]int64, error)

	WriteInt16(w *offsetWriter, x int16) error
	WriteInt32(w *offsetWriter, x int32) error
	WriteInt64(w *offsetWriter, x int64) error
	WriteFloat32(w *offsetWriter, x float32) error
	WriteFloat64(w *offsetWriter, x float64) error
	WriteString(w *offsetWriter, x string) error
}

var (
	// NetworkLittleEndian is the variable sized integer implementation of NBT. It is otherwise the same as the
	// normal little endian NBT. The NetworkLittleEndian format limits the total bytes of NBT that may be read. If
	// the limit is hit, the reading operation will fail immediately. NetworkLittleEndian is generally used for NBT
	// sent over network in the Bedrock Edition protocol.
	NetworkLittleEndian networkLittleEndian

	// LittleEndian is the fixed size little endian implementation of NBT. It is the format typically used for
	// writing Minecraft (Bedrock Edition) world saves.
	LittleEndian littleEndian

	// BigEndian is the fixed size big endian implementation of NBT. It is the original implementation, and is
	// used only on Minecraft Java Edition.
	BigEndian bigEndian

	_ Encoding = NetworkLittleEndian
	_ Encoding = LittleEndian
	_ Encoding = BigEndian
)

type networkLittleEndian struct{ littleEndian }

// WriteInt32 ...
func (networkLittleEndian) WriteInt32(w *offsetWriter, x int32) error {
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt32", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

// WriteInt64 ...
func (networkLittleEndian) WriteInt64(w *offsetWriter, x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt64", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

// WriteString ...
func (networkLittleEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > math.MaxInt16 {
		return InvalidStringError{Off: w.off, String: x, Err: errors.New("string length exceeds maximum length prefix")}
	}
	ux := uint32(len(x))
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteString", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

// Int32 ...
func (networkLittleEndian) Int32(r *offsetReader) (int32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int32"}
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	x := int32(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}

// Int64 ...
func (networkLittleEndian) Int64(r *offsetReader) (int64, error) {
	var ux uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int64"}
		}
		ux |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}

// String ...
func (e networkLittleEndian) String(r *offsetReader) (string, error) {
	var length uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return "", BufferOverrunError{Op: "String"}
		}
		length |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	if length > math.MaxInt16 {
		return "", InvalidStringError{Off: r.off, Err: errors.New("string length exceeds maximum length prefix")}
	}
	data := make([]byte, length)
	if _, err := r.Read(data); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&data)), nil
}

// Int32Slice ...
func (e networkLittleEndian) Int32Slice(r *offsetReader) ([]int32, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int32Slice"}
	}
	m := make([]int32, n)
	for i := int32(0); i < n; i++ {
		m[i], err = e.Int32(r)
		if err != nil {
			return nil, BufferOverrunError{Op: "Int32Slice"}
		}
	}
	return m, nil
}

// Int64Slice ...
func (e networkLittleEndian) Int64Slice(r *offsetReader) ([]int64, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int64Slice"}
	}
	m := make([]int64, n)
	for i := int32(0); i < n; i++ {
		m[i], err = e.Int64(r)
		if err != nil {
			return nil, BufferOverrunError{Op: "Int64Slice"}
		}
	}
	return m, nil
}
