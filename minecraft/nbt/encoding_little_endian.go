//go:build !armbe && !arm64be && !ppc64 && !mips && !mips64 && !mips64p32 && !ppc && !sparc && !sparc64 && !s390 && !s390x

package nbt

import (
	"encoding/binary"
	"math"
	"unsafe"
)

type littleEndian struct{}

// WriteInt16 ...
func (littleEndian) WriteInt16(w *offsetWriter, x int16) error {
	b := *(*[2]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

// WriteInt32 ...
func (littleEndian) WriteInt32(w *offsetWriter, x int32) error {
	b := *(*[4]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

// WriteInt64 ...
func (littleEndian) WriteInt64(w *offsetWriter, x int64) error {
	b := *(*[8]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

// WriteFloat32 ...
func (littleEndian) WriteFloat32(w *offsetWriter, x float32) error {
	b := *(*[4]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

// WriteFloat64 ...
func (littleEndian) WriteFloat64(w *offsetWriter, x float64) error {
	b := *(*[8]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
	}
	return nil
}

// WriteString ...
func (e littleEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > maxStringSize {
		return InvalidStringError{Off: w.off, N: uint(len(x)), Err: errStringTooLong}
	}
	if err := e.WriteInt16(w, int16(uint16(len(x)))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	b := *(*[]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

// Int16 ...
func (littleEndian) Int16(r *offsetReader) (int16, error) {
	b := make([]byte, 2)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return *(*int16)(unsafe.Pointer(&b[0])), nil
}

// Int32 ...
func (littleEndian) Int32(r *offsetReader) (int32, error) {
	b := make([]byte, 4)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return *(*int32)(unsafe.Pointer(&b[0])), nil
}

// Int64 ...
func (littleEndian) Int64(r *offsetReader) (int64, error) {
	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return *(*int64)(unsafe.Pointer(&b[0])), nil
}

// Float32 ...
func (littleEndian) Float32(r *offsetReader) (float32, error) {
	b := make([]byte, 4)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return *(*float32)(unsafe.Pointer(&b[0])), nil
}

// Float64 ...
func (littleEndian) Float64(r *offsetReader) (float64, error) {
	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return *(*float64)(unsafe.Pointer(&b[0])), nil
}

// String ...
func (e littleEndian) String(r *offsetReader) (string, error) {
	strLen, err := e.Int16(r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	b := make([]byte, uint16(strLen))
	if _, err := r.Read(b); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&b)), nil
}

// Int32Slice ...
func (e littleEndian) Int32Slice(r *offsetReader) ([]int32, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int32Slice"}
	}
	b := make([]byte, n*4)
	if _, err := r.Read(b); err != nil {
		return nil, BufferOverrunError{Op: "Int32Slice"}
	}
	if n == 0 {
		return []int32{}, nil
	}
	return unsafe.Slice((*int32)(unsafe.Pointer(&b[0])), n), nil
}

// Int64Slice ...
func (e littleEndian) Int64Slice(r *offsetReader) ([]int64, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int64Slice"}
	}
	b := make([]byte, n*8)
	if _, err := r.Read(b); err != nil {
		return nil, BufferOverrunError{Op: "Int64Slice"}
	}
	if n == 0 {
		return []int64{}, nil
	}
	return unsafe.Slice((*int64)(unsafe.Pointer(&b[0])), n), nil
}

type bigEndian struct{}

// WriteInt16 ...
func (bigEndian) WriteInt16(w *offsetWriter, x int16) error {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(x))
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

// WriteInt32 ...
func (bigEndian) WriteInt32(w *offsetWriter, x int32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

// WriteInt64 ...
func (bigEndian) WriteInt64(w *offsetWriter, x int64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

// WriteFloat32 ...
func (bigEndian) WriteFloat32(w *offsetWriter, x float32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, math.Float32bits(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

// WriteFloat64 ...
func (bigEndian) WriteFloat64(w *offsetWriter, x float64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(x))
	if _, err := w.Write(b[:]); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
	}
	return nil
}

// WriteString ...
func (e bigEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > maxStringSize {
		return InvalidStringError{Off: w.off, N: uint(len(x)), Err: errStringTooLong}
	}
	if err := e.WriteInt16(w, int16(uint16(len(x)))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	b := *(*[]byte)(unsafe.Pointer(&x))
	if _, err := w.Write(b); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

// Int16 ...
func (bigEndian) Int16(r *offsetReader) (int16, error) {
	b := make([]byte, 2)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}

// Int32 ...
func (bigEndian) Int32(r *offsetReader) (int32, error) {
	b := make([]byte, 4)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

// Int64 ...
func (bigEndian) Int64(r *offsetReader) (int64, error) {
	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

// Float32 ...
func (bigEndian) Float32(r *offsetReader) (float32, error) {
	b := make([]byte, 4)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(binary.BigEndian.Uint32(b)), nil
}

// Float64 ...
func (bigEndian) Float64(r *offsetReader) (float64, error) {
	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b)), nil
}

// String ...
func (e bigEndian) String(r *offsetReader) (string, error) {
	strLen, err := e.Int16(r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	b := make([]byte, uint16(strLen))
	if _, err := r.Read(b); err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return *(*string)(unsafe.Pointer(&b)), nil
}

// Int32Slice ...
func (e bigEndian) Int32Slice(r *offsetReader) ([]int32, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int32Slice"}
	}
	b := make([]byte, n*4)
	if _, err := r.Read(b); err != nil {
		return nil, BufferOverrunError{Op: "Int32Slice"}
	}
	if n == 0 {
		return []int32{}, nil
	}
	// Manually rotate the bytes, so we can just re-interpret this as a slice.
	for i := int32(0); i < n; i++ {
		off := i * 4
		b[off], b[off+3] = b[off+3], b[off]
		b[off+1], b[off+2] = b[off+2], b[off+1]
	}
	return unsafe.Slice((*int32)(unsafe.Pointer(&b[0])), n), nil
}

// Int64Slice ...
func (e bigEndian) Int64Slice(r *offsetReader) ([]int64, error) {
	n, err := e.Int32(r)
	if err != nil {
		return nil, BufferOverrunError{Op: "Int64Slice"}
	}
	b := make([]byte, n*8)
	if _, err := r.Read(b); err != nil {
		return nil, BufferOverrunError{Op: "Int64Slice"}
	}
	if n == 0 {
		return []int64{}, nil
	}
	// Manually rotate the bytes, so we can just re-interpret this as a slice.
	for i := int32(0); i < n; i++ {
		off := i * 4
		b[off], b[off+7] = b[off+7], b[off]
		b[off+1], b[off+6] = b[off+6], b[off+1]
		b[off+2], b[off+5] = b[off+5], b[off+2]
		b[off+3], b[off+4] = b[off+4], b[off+3]
	}
	return unsafe.Slice((*int64)(unsafe.Pointer(&b[0])), n), nil
}
