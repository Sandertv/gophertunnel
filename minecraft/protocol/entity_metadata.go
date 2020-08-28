package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const (
	EntityDataByte = iota
	EntityDataInt16
	EntityDataInt32
	EntityDataFloat32
	EntityDataString
	EntityDataCompoundTag
	EntityDataBlockPos
	EntityDataInt64
	EntityDataVec3
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

// WriteEntityMetadata writes an entity metadata list x to buffer dst. The types held by the map must be one
// of byte, int16, int32, float32, string, map[string]interface{}, BlockPos, int64 or mgl32.Vec3. The function
// will panic if a different type is encountered.
func WriteEntityMetadata(dst *bytes.Buffer, x map[uint32]interface{}) error {
	if x == nil {
		return WriteVaruint32(dst, 0)
	}
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return wrap(err)
	}
	for key, value := range x {
		if err := WriteVaruint32(dst, key); err != nil {
			return wrap(err)
		}
		var typeErr, valueErr error
		switch v := value.(type) {
		case byte:
			typeErr = WriteVaruint32(dst, EntityDataByte)
			valueErr = binary.Write(dst, binary.LittleEndian, v)
		case int16:
			typeErr = WriteVaruint32(dst, EntityDataInt16)
			valueErr = binary.Write(dst, binary.LittleEndian, v)
		case int32:
			typeErr = WriteVaruint32(dst, EntityDataInt32)
			valueErr = WriteVarint32(dst, v)
		case float32:
			typeErr = WriteVaruint32(dst, EntityDataFloat32)
			valueErr = WriteFloat32(dst, v)
		case string:
			typeErr = WriteVaruint32(dst, EntityDataString)
			valueErr = WriteString(dst, v)
		case map[string]interface{}:
			typeErr = WriteVaruint32(dst, EntityDataCompoundTag)
			valueErr = nbt.NewEncoder(dst).Encode(v)
			if valueErr != nil {
				panic(fmt.Errorf("cannot encode entity metadata: %w", valueErr))
			}
		case BlockPos:
			typeErr = WriteVaruint32(dst, EntityDataBlockPos)
			valueErr = WriteBlockPosition(dst, v)
		case int64:
			typeErr = WriteVaruint32(dst, EntityDataInt64)
			valueErr = WriteVarint64(dst, v)
		case mgl32.Vec3:
			typeErr = WriteVaruint32(dst, EntityDataVec3)
			valueErr = WriteVec3(dst, v)
		default:
			panic(fmt.Sprintf("invalid entity metadata value type %T", value))
		}
		if typeErr != nil {
			return wrap(typeErr)
		}
		if valueErr != nil {
			return wrap(valueErr)
		}
	}
	return nil
}
