package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"reflect"
)

const (
	EntityDataByte uint32 = iota
	EntityDataInt16
	EntityDataInt32
	EntityDataFloat32
	EntityDataString
	EntityDataCompoundTag
	EntityDataBlockPos
	EntityDataInt64
	EntityDataVec3
)

var (
	entityDataByte        = EntityDataByte
	entityDataInt16       = EntityDataInt16
	entityDataInt32       = EntityDataInt32
	entityDataFloat32     = EntityDataFloat32
	entityDataString      = EntityDataString
	entityDataCompoundTag = EntityDataCompoundTag
	entityDataBlockPos    = EntityDataBlockPos
	entityDataInt64       = EntityDataInt64
	entityDataVec3        = EntityDataVec3
)

// EntityMetadata reads an entity metadata list from Reader r into map x. The types in the map will be one
// of byte, int16, int32, float32, string, map[string]interface{}, BlockPos, int64 or mgl32.Vec3.
func EntityMetadata(r *Reader, x *map[uint32]interface{}) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, mediumLimit)
	for i := uint32(0); i < count; i++ {
		var key, dataType uint32
		r.Varuint32(&key)
		r.Varuint32(&dataType)
		switch dataType {
		case EntityDataByte:
			var v byte
			r.Uint8(&v)
			(*x)[key] = v
		case EntityDataInt16:
			var v int16
			r.Int16(&v)
			(*x)[key] = v
		case EntityDataInt32:
			var v int32
			r.Varint32(&v)
			(*x)[key] = v
		case EntityDataFloat32:
			var v float32
			r.Float32(&v)
			(*x)[key] = v
		case EntityDataString:
			var v string
			r.String(&v)
			(*x)[key] = v
		case EntityDataCompoundTag:
			var v map[string]interface{}
			r.NBT(&v, nbt.NetworkLittleEndian)
			(*x)[key] = v
		case EntityDataBlockPos:
			var v BlockPos
			r.BlockPos(&v)
			(*x)[key] = v
		case EntityDataInt64:
			var v int64
			r.Varint64(&v)
			(*x)[key] = v
		case EntityDataVec3:
			var v mgl32.Vec3
			r.Vec3(&v)
			(*x)[key] = v
		default:
			r.UnknownEnumOption(dataType, "entity metadata")
		}
	}
}

// WriteEntityMetadata writes an entity metadata list x to Writer w. The types held by the map must be one
// of byte, int16, int32, float32, string, map[string]interface{}, BlockPos, int64 or mgl32.Vec3. The function
// will panic if a different type is encountered.
func WriteEntityMetadata(w *Writer, x *map[uint32]interface{}) {
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
