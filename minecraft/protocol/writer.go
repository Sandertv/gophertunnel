package protocol

import (
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"image/color"
	"reflect"
	"unsafe"
)

// Writer implements writing methods for data types from Minecraft packets. Each Packet implementation has one
// passed to it when writing.
type Writer struct {
	buf []byte
}

// NewWriter creates a new initialised Writer with an initial underlying buffer size of 1024 bytes.
func NewWriter() *Writer {
	return &Writer{buf: make([]byte, 0, 1024)}
}

// Uint8 writes a uint8 to the underlying buffer.
func (w *Writer) Uint8(x *uint8) {
	w.buf = append(w.buf, *x)
}

// Bool writes a bool as either 0 or 1 to the underlying buffer.
func (w *Writer) Bool(x *bool) {
	if *x {
		w.buf = append(w.buf, 1)
		return
	}
	w.buf = append(w.buf, 0)
}

// String writes a string, prefixed with a varuint32, to the underlying buffer.
func (w *Writer) String(x *string) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	w.buf = append(w.buf, *(*[]byte)(unsafe.Pointer(x))...)
}

// ByteSlice writes a []byte, prefixed with a varuint32, to the underlying buffer.
func (w *Writer) ByteSlice(x *[]byte) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	w.buf = append(w.buf, *x...)
}

// Bytes appends a []byte to the underlying buffer.
func (w *Writer) Bytes(x *[]byte) {
	w.buf = append(w.buf, *x...)
}

// ByteFloat writes a rotational float32 as a single byte to the underlying buffer.
func (w *Writer) ByteFloat(x *float32) {
	w.buf = append(w.buf, byte(*x/(360.0/256.0)))
}

// Vec3 writes an mgl32.Vec3 as 3 float32s to the underlying buffer.
func (w *Writer) Vec3(x *mgl32.Vec3) {
	w.Float32(&x[0])
	w.Float32(&x[1])
	w.Float32(&x[2])
}

// Vec2 writes an mgl32.Vec2 as 2 float32s to the underlying buffer.
func (w *Writer) Vec2(x *mgl32.Vec2) {
	w.Float32(&x[0])
	w.Float32(&x[1])
}

// BlockPos writes a BlockPos as 3 varint32s to the underlying buffer.
func (w *Writer) BlockPos(x *BlockPos) {
	w.Varint32(&x[0])
	w.Varint32(&x[1])
	w.Varint32(&x[2])
}

// UBlockPos writes a BlockPos as 2 varint32s and a varuint32 to the underlying buffer.
func (w *Writer) UBlockPos(x *BlockPos) {
	w.Varint32(&x[0])
	y := uint32(x[1])
	w.Varuint32(&y)
	w.Varint32(&x[2])
}

// VarRGBA writes a color.RGBA x as a varuint32 to the underlying buffer.
func (w *Writer) VarRGBA(x *color.RGBA) {
	val := uint32(x.R) | uint32(x.G)<<8 | uint32(x.B)<<16 | uint32(x.A)<<24
	w.Varuint32(&val)
}

// UUID writes a UUID to the underlying buffer.
func (w *Writer) UUID(x *uuid.UUID) {
	b := append((*x)[8:], (*x)[:8]...)
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	w.buf = append(w.buf, b...)
}

// EntityMetadata writes an entity metadata map x to the underlying buffer.
func (w *Writer) EntityMetadata(x *map[uint32]interface{}) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	for key, value := range *x {
		w.Varuint32(&key)
		switch v := value.(type) {
		case byte:
			w.Varuint32(&entityDataByte)
			w.Uint8(&v)
		case int16:
			w.Varuint32(&entityDataInt16)
			w.Int16(&v)
		case int32:
			w.Varuint32(&entityDataInt32)
			w.Varint32(&v)
		case float32:
			w.Varuint32(&entityDataFloat32)
			w.Float32(&v)
		case string:
			w.Varuint32(&entityDataString)
			w.String(&v)
		case map[string]interface{}:
			w.Varuint32(&entityDataCompoundTag)
			w.NBT(&v, nbt.NetworkLittleEndian)
		case BlockPos:
			w.Varuint32(&entityDataBlockPos)
			w.BlockPos(&v)
		case int64:
			w.Varuint32(&entityDataInt64)
			w.Varint64(&v)
		case mgl32.Vec3:
			w.Varuint32(&entityDataVec3)
			w.Vec3(&v)
		default:
			w.UnknownEnumOption(reflect.TypeOf(value), "entity metadata")
		}
	}
}

// Varint64 writes an int64 as 1-10 bytes to the underlying buffer.
func (w *Writer) Varint64(x *int64) {
	u := *x
	ux := uint64(u) << 1
	if u < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		w.buf = append(w.buf, byte(ux)|0x80)
		ux >>= 7
	}
	w.buf = append(w.buf, byte(ux))
}

// Varuint64 writes a uint64 as 1-10 bytes to the underlying buffer.
func (w *Writer) Varuint64(x *uint64) {
	u := *x
	for u >= 0x80 {
		w.buf = append(w.buf, byte(u)|0x80)
		u >>= 7
	}
	w.buf = append(w.buf, byte(u))
}

// Varint32 writes an int32 as 1-5 bytes to the underlying buffer.
func (w *Writer) Varint32(x *int32) {
	u := *x
	ux := uint32(u) << 1
	if u < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		w.buf = append(w.buf, byte(ux)|0x80)
		ux >>= 7
	}
	w.buf = append(w.buf, byte(ux))
}

// Varuint32 writes a uint32 as 1-5 bytes to the underlying buffer.
func (w *Writer) Varuint32(x *uint32) {
	u := *x
	for u >= 0x80 {
		w.buf = append(w.buf, byte(u)|0x80)
		u >>= 7
	}
	w.buf = append(w.buf, byte(u))
}

// NBT writes a map as NBT to the underlying buffer using the encoding passed.
func (w *Writer) NBT(x *map[string]interface{}, encoding nbt.Encoding) {
	var buf *bytes.Buffer
	if len(*x) == 0 {
		buf = bytes.NewBuffer(make([]byte, 0, 2))
	} else if len(*x) < 4 {
		buf = bytes.NewBuffer(make([]byte, 0, 48))
	} else {
		buf = bytes.NewBuffer(make([]byte, 0, 128))
	}
	if err := nbt.NewEncoderWithEncoding(buf, encoding).Encode(*x); err != nil {
		panic(err)
	}
	w.buf = append(w.buf, buf.Bytes()...)
}

// NBTList writes a slice as NBT to the underlying buffer using the encoding passed.
func (w *Writer) NBTList(x *[]interface{}, encoding nbt.Encoding) {
	var buf *bytes.Buffer
	if len(*x) == 0 {
		buf = bytes.NewBuffer(make([]byte, 0, 2))
	} else if len(*x) < 4 {
		buf = bytes.NewBuffer(make([]byte, 0, 48))
	} else {
		buf = bytes.NewBuffer(make([]byte, 0, 128))
	}
	if err := nbt.NewEncoderWithEncoding(buf, encoding).Encode(*x); err != nil {
		panic(err)
	}
	w.buf = append(w.buf, buf.Bytes()...)
}

// UnknownEnumOption panics with an unknown enum option error.
func (w *Writer) UnknownEnumOption(value interface{}, enum string) {
	w.panicf("unknown value '%v' for enum type '%v'", value, enum)
}

// InvalidValue panics with an invalid value error.
func (w *Writer) InvalidValue(value interface{}, forField, reason string) {
	w.panicf("invalid value '%v' for %v: %v", value, forField, reason)
}

// Data returns all bytes written to the Writer. Note that these bytes are only valid until the next call to
// Reset.
func (w *Writer) Data() []byte {
	return w.buf
}

// Reset resets the length of the underlying buffer. The underlying array is not removed, so writing to the
// Writer again will result in reduced allocations.
func (w *Writer) Reset() {
	w.buf = w.buf[:0]
}

// panicf panics with the format and values passed.
func (w *Writer) panicf(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a...))
}

// panic panics with the error passed, similarly to panicf.
func (w *Writer) panic(err error) {
	panic(err)
}
