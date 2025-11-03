package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerUpdateEntityOverridesTypeClearAll = iota
	PlayerUpdateEntityOverridesTypeRemove
	PlayerUpdateEntityOverridesTypeInt
	PlayerUpdateEntityOverridesTypeFloat
)

// PlayerUpdateEntityOverrides is sent by the server to modify an entity's properties individually.
type PlayerUpdateEntityOverrides struct {
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	EntityUniqueID int64
	// PropertyIndex is the index of the property to modify. The index is unique for each property of an entity.
	PropertyIndex uint32
	// Type is the type of action to perform with the property. It is one of the constants above.
	Type byte
	// IntValue is the new integer value of the property. It is only used when Type is set to
	// PlayerUpdateEntityOverridesTypeInt.
	IntValue int32
	// FloatValue is the new float value of the property. It is only used when Type is set to
	// PlayerUpdateEntityOverridesTypeFloat.
	FloatValue float32
}

// ID ...
func (*PlayerUpdateEntityOverrides) ID() uint32 {
	return IDPlayerUpdateEntityOverrides
}

func (pk *PlayerUpdateEntityOverrides) Marshal(io protocol.IO) {
	io.Varint64(&pk.EntityUniqueID)
	io.Varuint32(&pk.PropertyIndex)
	io.Uint8(&pk.Type)
	if pk.Type == PlayerUpdateEntityOverridesTypeInt {
		io.Int32(&pk.IntValue)
	} else if pk.Type == PlayerUpdateEntityOverridesTypeFloat {
		io.Float32(&pk.FloatValue)
	}
}
