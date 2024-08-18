package protocol

import (
	"fmt"
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
	SoundPos(x *mgl32.Vec3)
	ByteFloat(x *float32)
	Bytes(p *[]byte)
	NBT(m *map[string]any, encoding nbt.Encoding)
	NBTList(m *[]any, encoding nbt.Encoding)
	UUID(x *uuid.UUID)
	RGB(x *color.RGBA)
	RGBA(x *color.RGBA)
	VarRGBA(x *color.RGBA)
	EntityMetadata(x *map[uint32]any)
	Item(x *ItemStack)
	ItemInstance(i *ItemInstance)
	ItemDescriptorCount(i *ItemDescriptorCount)
	StackRequestAction(x *StackRequestAction)
	MaterialReducer(x *MaterialReducer)
	Recipe(x *Recipe)
	EventType(x *Event)
	TransactionDataType(x *InventoryTransactionData)
	PlayerInventoryAction(x *UseItemTransactionData)
	GameRule(x *GameRule)
	AbilityValue(x *any)
	CompressedBiomeDefinitions(x *map[string]any)

	ShieldID() int32
	UnknownEnumOption(value any, enum string)
	InvalidValue(value any, forField, reason string)
}

// Marshaler is a type that can be written to or read from an IO.
type Marshaler interface {
	Marshal(r IO)
}

// Slice reads/writes a slice of T with a varuint32 prefix.
func Slice[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	SliceOfLen[T, S, A](r, count, x)
}

// SliceUint8Length reads/writes a slice of T with a uint8 prefix.
func SliceUint8Length[T any, S *[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint8(len(*x))
	r.Uint8(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// SliceUint16Length reads/writes a slice of T with a uint16 prefix.
func SliceUint16Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint16(len(*x))
	r.Uint16(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// SliceUint32Length reads/writes a slice of T with a uint32 prefix.
func SliceUint32Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint32(len(*x))
	r.Uint32(&count)
	SliceOfLen[T, S, A](r, count, x)
}

// SliceVarint32Length reads/writes a slice of T with a varint32 prefix.
func SliceVarint32Length[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := int32(len(*x))
	r.Varint32(&count)
	SliceOfLen[T, S, A](r, uint32(count), x)
}

// FuncSliceUint16Length reads/writes a slice of T using function f with a uint16 length prefix.
func FuncSliceUint16Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint16(len(*x))
	r.Uint16(&count)
	FuncSliceOfLen(r, uint32(count), x, f)
}

// FuncSliceUint32Length reads/writes a slice of T using function f with a uint32 length prefix.
func FuncSliceUint32Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Uint32(&count)
	FuncSliceOfLen(r, count, x, f)
}

// FuncSlice reads/writes a slice of T using function f with a varuint32 length prefix.
func FuncSlice[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	FuncSliceOfLen(r, count, x, f)
}

// FuncIOSlice reads/writes a slice of T using a function f with a varuint32 length prefix.
func FuncIOSlice[T any, S ~*[]T](r IO, x S, f func(IO, *T)) {
	FuncSlice(r, x, func(v *T) {
		f(r, v)
	})
}

// FuncIOSliceUint32Length reads/writes a slice of T using a function with a uint32 length prefix.
func FuncIOSliceUint32Length[T any, S ~*[]T](r IO, x S, f func(IO, *T)) {
	count := uint32(len(*x))
	r.Uint32(&count)
	FuncIOSliceOfLen(r, count, x, f)
}

const maxSliceLength = 1024

// SliceOfLen reads/writes the elements of a slice of type T with length l.
func SliceOfLen[T any, S ~*[]T, A PtrMarshaler[T]](r IO, l uint32, x S) {
	rd, reader := r.(Reads)
	if reader {
		if rd.LimitsEnabled() && l > maxSliceLength {
			panic(fmt.Errorf("slice length was too long: length of %v", l))
		}
		*x = make([]T, l)
	}

	for i := uint32(0); i < l; i++ {
		A(&(*x)[i]).Marshal(r)
	}
}

// FuncSliceOfLen reads/writes the elements of a slice of type T with length l using func f.
func FuncSliceOfLen[T any, S ~*[]T](r IO, l uint32, x S, f func(*T)) {
	rd, reader := r.(Reads)
	if reader {
		if rd.LimitsEnabled() && l > maxSliceLength {
			panic(fmt.Errorf("slice length was too long: length of %v", l))
		}
		*x = make([]T, l)
	}

	for i := uint32(0); i < l; i++ {
		f(&(*x)[i])
	}
}

// FuncIOSliceOfLen reads/writes the elements of a slice of type T with length l using func f.
func FuncIOSliceOfLen[T any, S ~*[]T](r IO, l uint32, x S, f func(IO, *T)) {
	FuncSliceOfLen(r, l, x, func(v *T) {
		f(r, v)
	})
}

// PtrMarshaler represents a type that implements Marshaler for its pointer.
type PtrMarshaler[T any] interface {
	Marshaler
	*T
}

// Single reads/writes a single Marshaler x.
func Single[T any, S PtrMarshaler[T]](r IO, x S) {
	x.Marshal(r)
}

// Optional is an optional type in the protocol. If not set, only a false bool is written. If set, a true bool is
// written and the Marshaler.
type Optional[T any] struct {
	set bool
	val T
}

// Option creates an Optional[T] with the value passed.
func Option[T any](val T) Optional[T] {
	return Optional[T]{set: true, val: val}
}

// Value returns the value set in the Optional. If no value was set, false is returned.
func (o Optional[T]) Value() (T, bool) {
	return o.val, o.set
}

// OptionalFunc reads/writes an Optional[T].
func OptionalFunc[T any](r IO, x *Optional[T], f func(*T)) any {
	r.Bool(&x.set)
	if x.set {
		f(&x.val)
	}
	return x
}

// OptionalFuncIO reads/writes an Optional[T].
func OptionalFuncIO[T any](r IO, x *Optional[T], f func(IO, *T)) any {
	r.Bool(&x.set)
	if x.set {
		f(r, &x.val)
	}
	return x
}

// OptionalMarshaler reads/writes an Optional assuming *T implements Marshaler.
func OptionalMarshaler[T any, A PtrMarshaler[T]](r IO, x *Optional[T]) {
	r.Bool(&x.set)
	if x.set {
		A(&x.val).Marshal(r)
	}
}
