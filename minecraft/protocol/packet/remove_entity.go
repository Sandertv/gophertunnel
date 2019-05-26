package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RemoveEntity is sent by the server to remove an entity that currently exists in the world from the client-
// side. Sending this packet if the client cannot already see this entity will have no effect.
type RemoveEntity struct {
	// EntityUniqueID is the unique ID of the entity to be removed. The unique ID is a value that remains
	// consistent across different sessions of the same world, but most servers simply fill the runtime ID
	// of the entity out for this field.
	EntityUniqueID int64
}

// ID ...
func (*RemoveEntity) ID() uint32 {
	return IDRemoveEntity
}

// Marshal ...
func (pk *RemoveEntity) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint64(buf, pk.EntityUniqueID)
}

// Unmarshal ...
func (pk *RemoveEntity) Unmarshal(buf *bytes.Buffer) error {
	return protocol.Varint64(buf, &pk.EntityUniqueID)
}
