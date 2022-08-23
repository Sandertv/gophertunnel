package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"image/color"
)

// IO represents a packet IO direction. Implementations of this interface are Reader and Writer. Reader reads
// data from the input stream into the pointers passed, whereas Writer writes the values the pointers point to
// to the output stream.
type IO interface {
	Uint16(x *uint16)
	Int16(x *int16)
	Uint32(x *uint32)
	Int32(x *int32)
	BEInt32(x *int32)
	Uint64(x *uint64)
	Int64(x *int64)
	Float32(x *float32)
	Uint8(x *uint8)
	Int8(x *int8)
	Bool(x *bool)
	Varint64(x *int64)
	Varuint64(x *uint64)
	Varint32(x *int32)
	Varuint32(x *uint32)
	String(x *string)
	StringUTF(x *string)
	ByteSlice(x *[]byte)
	Vec3(x *mgl32.Vec3)
	Vec2(x *mgl32.Vec2)
	BlockPos(x *BlockPos)
	UBlockPos(x *BlockPos)
	ChunkPos(x *ChunkPos)
	SubChunkPos(x *SubChunkPos)
	ByteFloat(x *float32)
	Bytes(p *[]byte)
	NBT(m *map[string]any, encoding nbt.Encoding)
	NBTList(m *[]any, encoding nbt.Encoding)
	UUID(x *uuid.UUID)
	RGBA(x *color.RGBA)
	VarRGBA(x *color.RGBA)
	EntityMetadata(x *map[uint32]any)
	Item(x *ItemStack)
	ItemInstance(i *ItemInstance)
	MaterialReducer(x *MaterialReducer)

	UnknownEnumOption(value any, enum string)
	InvalidValue(value any, forField, reason string)
}

// Marshaler is a type that can be written to or read from an IO.
type Marshaler interface {
	Marshal(r IO) any
}

// Slice reads/writes a slice of T with a varuint32 prefix.
func Slice[T Marshaler, S ~*[]T](r IO, x S) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	slice(r, count, x)
}

// SliceUint8Length reads/writes a slice of T with a uint8 prefix.
func SliceUint8Length[T Marshaler, S ~*[]T](r IO, x S) {
	count := uint8(len(*x))
	r.Uint8(&count)
	slice(r, uint32(count), x)
}

// SliceUint16Length reads/writes a slice of T with a uint16 prefix.
func SliceUint16Length[T Marshaler, S ~*[]T](r IO, x S) {
	count := uint16(len(*x))
	r.Uint16(&count)
	slice(r, uint32(count), x)
}

// SliceUint32Length reads/writes a slice of T with a uint32 prefix.
func SliceUint32Length[T Marshaler, S ~*[]T](r IO, x S) {
	count := uint32(len(*x))
	r.Uint32(&count)
	slice(r, count, x)
}

// FuncSliceUint16Length reads/writes a slice of T using function f with a uint16 length prefix.
func FuncSliceUint16Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint16(len(*x))
	r.Uint16(&count)

	_, reader := r.(*Reader)
	if reader {
		*x = make([]T, count)
	}

	for i := uint16(0); i < count; i++ {
		f(&(*x)[i])
	}
}

// FuncSliceUint32Length reads/writes a slice of T using function f with a uint32 length prefix.
func FuncSliceUint32Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Uint32(&count)

	_, reader := r.(*Reader)
	if reader {
		*x = make([]T, count)
	}

	for i := uint32(0); i < count; i++ {
		f(&(*x)[i])
	}
}

// FuncSlice reads/writes a slice of T using function f with a varuint32 length prefix.
func FuncSlice[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Varuint32(&count)

	_, reader := r.(*Reader)
	if reader {
		*x = make([]T, count)
	}

	for i := uint32(0); i < count; i++ {
		f(&(*x)[i])
	}
}

// slice reads/writes the elements of a slice of type T.
func slice[T Marshaler, S ~*[]T](r IO, count uint32, x S) {
	_, reader := r.(*Reader)
	if reader {
		*x = make([]T, count)
	}

	for i := uint32(0); i < count; i++ {
		(*x)[i] = (*x)[i].Marshal(r).(T)
	}
}

// Single reads/writes a single Marshaler x.
func Single[T Marshaler, S ~*T](r IO, x S) {
	*x = (*x).Marshal(r).(T)
}

type Opt interface {
	Marshaler | ~int
}

// Optional is an optional type in the protocol. If not set, only a false bool is written. If set, a true bool is
// written and the Marshaler.
type Optional[T any] struct {
	set bool
	val T
}

// MarshalOptional reads/writes an Optional[T].
func MarshalOptional[T any](r IO, x *Optional[T], f func(*T)) any {
	r.Bool(&x.set)
	if x.set && f != nil {
		f(&x.val)
	}
	return x
}

// Option creates an Optional[T] with the value passed.
func Option[T any](val T) Optional[T] {
	return Optional[T]{set: true, val: val}
}
