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
	Type uint32
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
	io.ActorUniqueID(&pk.EntityUniqueID)
	io.Varuint32(&pk.PropertyIndex)
	io.Varuint32(&pk.Type)
	names := [...]string{"clearoverrides", "removeoverride", "setintoverride", "setfloatoverride"}
	if pk.Type >= uint32(len(names)) {
		io.UnknownEnumOption(pk.Type, "entity override type")
		return
	}
	name := names[pk.Type]
	io.String(&name)
	switch pk.Type {
	case PlayerUpdateEntityOverridesTypeClearAll, PlayerUpdateEntityOverridesTypeRemove:
	case PlayerUpdateEntityOverridesTypeInt:
		io.Int32(&pk.IntValue)
	case PlayerUpdateEntityOverridesTypeFloat:
		io.Float32(&pk.FloatValue)
	}
}
