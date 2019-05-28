package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateAttributes is sent by the server to update an amount of attributes of any entity in the world. These
// attributes include ones such as the health or the movement speed of the entity.
type UpdateAttributes struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Attributes is a slice of new attributes that the entity gets. It includes attributes such as its
	// health, movement speed, etc. Note that only changed attributes have to be sent in this packet. It is
	// not required to send attributes that did not have their values changed.
	Attributes []protocol.Attribute
}

// ID ...
func (*UpdateAttributes) ID() uint32 {
	return IDUpdateAttributes
}

// Marshal ...
func (pk *UpdateAttributes) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteAttributes(buf, pk.Attributes)
}

// Unmarshal ...
func (pk *UpdateAttributes) Unmarshal(buf *bytes.Buffer) error {
	return ChainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Attributes(buf, &pk.Attributes),
	)
}
