package protocol

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
