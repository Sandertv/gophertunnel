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

// EntityMetadata reads an entity metadata list from buffer src into map x. The types in the map will be one
// of byte, int16, int32, float32, string, map[string]interface{}, BlockPos, int64 or mgl32.Vec3.
func EntityMetadata(src *bytes.Buffer, x *map[uint32]interface{}) error {
	var count uint32
	var err error
	if err = Varuint32(src, &count); err != nil {
		return wrap(err)
	}
	if count > mediumLimit {
		return LimitHitError{Limit: mediumLimit, Type: "entity metadata"}
	}
	for i := uint32(0); i < count; i++ {
		var key, dataType uint32
		if err = Varuint32(src, &key); err != nil {
			return wrap(err)
		}
		if err = Varuint32(src, &dataType); err != nil {
			return wrap(err)
		}
		switch dataType {
		case EntityDataByte:
			var v byte
			err = binary.Read(src, binary.LittleEndian, &v)
			(*x)[key] = v
		case EntityDataInt16:
			var v int16
			err = binary.Read(src, binary.LittleEndian, &v)
			(*x)[key] = v
		case EntityDataInt32:
			var v int32
			err = Varint32(src, &v)
			(*x)[key] = v
		case EntityDataFloat32:
			var v float32
			err = Float32(src, &v)
			(*x)[key] = v
		case EntityDataString:
			var v string
			err = String(src, &v)
			(*x)[key] = v
		case EntityDataCompoundTag:
			var v map[string]interface{}
			err = nbt.NewDecoder(src).Decode(&v)
			(*x)[key] = v
		case EntityDataBlockPos:
			var v BlockPos
			err = BlockPosition(src, &v)
			(*x)[key] = v
		case EntityDataInt64:
			var v int64
			err = Varint64(src, &v)
			(*x)[key] = v
		case EntityDataVec3:
			var v mgl32.Vec3
			err = Vec3(src, &v)
			(*x)[key] = v
		default:
			return fmt.Errorf("unknown entity data type %v", dataType)
		}
		if err != nil {
			// If the error from reading the entity data property was not nil, we return right away.
			return wrap(err)
		}
	}
	return nil
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
