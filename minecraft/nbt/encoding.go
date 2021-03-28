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

	WriteInt16(w *offsetWriter, x int16) error
	WriteInt32(w *offsetWriter, x int32) error
	WriteInt64(w *offsetWriter, x int64) error
	WriteFloat32(w *offsetWriter, x float32) error
	WriteFloat64(w *offsetWriter, x float64) error
	WriteString(w *offsetWriter, x string) error
}

// NetworkLittleEndian is the variable sized integer implementation of NBT. It is otherwise the same as the
// normal little endian NBT. The NetworkLittleEndian format limits the total bytes of NBT that may be read. If
// the limit is hit, the reading operation will fail immediately.
var NetworkLittleEndian networkLittleEndian

// LittleEndian is the fixed size little endian implementation of NBT. It is the format typically used for
// writing Minecraft (Bedrock Edition) world saves.
var LittleEndian littleEndian

// BigEndian is the fixed size big endian implementation of NBT. It is the original implementation, and is
// used only on Minecraft Java Edition.
var BigEndian bigEndian

var _ = BigEndian
var _ = LittleEndian
var _ = NetworkLittleEndian

type networkLittleEndian struct{}

// WriteInt16 ...
func (networkLittleEndian) WriteInt16(w *offsetWriter, x int16) error {
	b := make([]byte, 2)
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

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

// WriteFloat32 ...
func (networkLittleEndian) WriteFloat32(w *offsetWriter, x float32) error {
	bits := math.Float32bits(x)
	b := make([]byte, 4)
	b[0] = byte(bits)
	b[1] = byte(bits >> 8)
	b[2] = byte(bits >> 16)
	b[3] = byte(bits >> 24)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

// WriteFloat64 ...
func (networkLittleEndian) WriteFloat64(w *offsetWriter, x float64) error {
	bits := math.Float64bits(x)
	b := make([]byte, 8)
	b[0] = byte(bits)
	b[1] = byte(bits >> 8)
	b[2] = byte(bits >> 16)
	b[3] = byte(bits >> 24)
	b[4] = byte(bits >> 32)
	b[5] = byte(bits >> 40)
	b[6] = byte(bits >> 48)
	b[7] = byte(bits >> 56)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
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

// Int16 ...
func (networkLittleEndian) Int16(r *offsetReader) (int16, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(uint16(b[0]) | uint16(b[1])<<8), nil
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

// Float32 ...
func (networkLittleEndian) Float32(r *offsetReader) (float32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

// Float64 ...
func (networkLittleEndian) Float64(r *offsetReader) (float64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
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
	data, err := consumeN(int(length), r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return string(data), nil
}

type littleEndian struct{}

// WriteInt16 ...
func (littleEndian) WriteInt16(w *offsetWriter, x int16) error {
	b := make([]byte, 2)
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

// WriteInt32 ...
func (littleEndian) WriteInt32(w *offsetWriter, x int32) error {
	b := make([]byte, 4)
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	b[2] = byte(x >> 16)
	b[3] = byte(x >> 24)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

// WriteInt64 ...
func (littleEndian) WriteInt64(w *offsetWriter, x int64) error {
	b := make([]byte, 8)
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	b[2] = byte(x >> 16)
	b[3] = byte(x >> 24)
	b[4] = byte(x >> 32)
	b[5] = byte(x >> 40)
	b[6] = byte(x >> 48)
	b[7] = byte(x >> 56)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

// WriteFloat32 ...
func (littleEndian) WriteFloat32(w *offsetWriter, x float32) error {
	bits := math.Float32bits(x)
	b := make([]byte, 4)
	b[0] = byte(bits)
	b[1] = byte(bits >> 8)
	b[2] = byte(bits >> 16)
	b[3] = byte(bits >> 24)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

// WriteFloat64 ...
func (littleEndian) WriteFloat64(w *offsetWriter, x float64) error {
	bits := math.Float64bits(x)
	b := make([]byte, 8)
	b[0] = byte(bits)
	b[1] = byte(bits >> 8)
	b[2] = byte(bits >> 16)
	b[3] = byte(bits >> 24)
	b[4] = byte(bits >> 32)
	b[5] = byte(bits >> 40)
	b[6] = byte(bits >> 48)
	b[7] = byte(bits >> 56)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
	}
	return nil
}

// WriteString ...
func (littleEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > math.MaxInt16 {
		return InvalidStringError{Off: w.off, String: x, Err: errors.New("string length exceeds maximum length prefix")}
	}
	length := int16(len(x))
	b := make([]byte, 2)
	b[0], b[1] = byte(length), byte(length>>8)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

// Int16 ...
func (littleEndian) Int16(r *offsetReader) (int16, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(uint16(b[0]) | uint16(b[1])<<8), nil
}

// Int32 ...
func (littleEndian) Int32(r *offsetReader) (int32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

// Int64 ...
func (littleEndian) Int64(r *offsetReader) (int64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return int64(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

// Float32 ...
func (littleEndian) Float32(r *offsetReader) (float32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

// Float64 ...
func (littleEndian) Float64(r *offsetReader) (float64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

// String ...
func (littleEndian) String(r *offsetReader) (string, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	stringLength := int(uint16(b[0]) | uint16(b[1])<<8)
	data, err := consumeN(stringLength, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return string(data), nil
}

type bigEndian struct{}

// WriteInt16 ...
func (bigEndian) WriteInt16(w *offsetWriter, x int16) error {
	b := make([]byte, 2)
	b[0] = byte(x >> 8)
	b[1] = byte(x)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

// WriteInt32 ...
func (bigEndian) WriteInt32(w *offsetWriter, x int32) error {
	b := make([]byte, 4)
	b[0] = byte(x >> 24)
	b[1] = byte(x >> 16)
	b[2] = byte(x >> 8)
	b[3] = byte(x)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

// WriteInt64 ...
func (bigEndian) WriteInt64(w *offsetWriter, x int64) error {
	b := make([]byte, 8)
	b[0] = byte(x >> 56)
	b[1] = byte(x >> 48)
	b[2] = byte(x >> 40)
	b[3] = byte(x >> 32)
	b[4] = byte(x >> 24)
	b[5] = byte(x >> 16)
	b[6] = byte(x >> 8)
	b[7] = byte(x)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

// WriteFloat32 ...
func (bigEndian) WriteFloat32(w *offsetWriter, x float32) error {
	bits := math.Float32bits(x)
	b := make([]byte, 8)
	b[0] = byte(bits >> 24)
	b[1] = byte(bits >> 16)
	b[2] = byte(bits >> 8)
	b[3] = byte(bits)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

// WriteFloat64 ...
func (bigEndian) WriteFloat64(w *offsetWriter, x float64) error {
	bits := math.Float64bits(x)
	b := make([]byte, 8)
	b[0] = byte(bits >> 56)
	b[1] = byte(bits >> 48)
	b[2] = byte(bits >> 40)
	b[3] = byte(bits >> 32)
	b[4] = byte(bits >> 24)
	b[5] = byte(bits >> 16)
	b[6] = byte(bits >> 8)
	b[7] = byte(bits)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
	}
	return nil
}

// WriteString ...
func (bigEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > math.MaxInt16 {
		return InvalidStringError{Off: w.off, String: x, Err: errors.New("string length exceeds maximum length prefix")}
	}
	length := int16(len(x))
	b := make([]byte, 2)
	b[0], b[1] = byte(length>>8), byte(length)
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

// Int16 ...
func (bigEndian) Int16(r *offsetReader) (int16, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(uint16(b[0])<<8 | uint16(b[1])), nil
}

// Int32 ...
func (bigEndian) Int32(r *offsetReader) (int32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])), nil
}

// Int64 ...
func (bigEndian) Int64(r *offsetReader) (int64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return int64(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])), nil
}

// Float32 ...
func (bigEndian) Float32(r *offsetReader) (float32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])), nil
}

// Float64 ...
func (bigEndian) Float64(r *offsetReader) (float64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])), nil
}

// String ...
func (bigEndian) String(r *offsetReader) (string, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	stringLength := int(uint16(b[0])<<8 | uint16(b[1]))
	data, err := consumeN(stringLength, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return string(data), nil
}

// consumeN consumes n bytes from the offset reader and returns them. It returns an error if the reader does
// not have that many bytes available.
func consumeN(n int, r *offsetReader) ([]byte, error) {
	if n < 0 {
		return nil, InvalidArraySizeError{Off: r.off, Op: "Consume", NBTLength: n}
	}
	data := r.Next(n)
	if len(data) != n {
		return nil, BufferOverrunError{Op: "Consume"}
	}
	return data, nil
}
