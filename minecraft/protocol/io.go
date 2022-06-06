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
	VarRGBA(x *color.RGBA)
	EntityMetadata(x *map[uint32]any)
	Item(x *ItemStack)
	ItemInstance(i *ItemInstance)
	MaterialReducer(x *MaterialReducer)

	UnknownEnumOption(value any, enum string)
	InvalidValue(value any, forField, reason string)
}
