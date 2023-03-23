package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RemoveActor is sent by the server to remove an entity that currently exists in the world from the client-
// side. Sending this packet if the client cannot already see this entity will have no effect.
type RemoveActor struct {
	// EntityUniqueID is the unique ID of the entity to be removed. The unique ID is a value that remains
	// consistent across different sessions of the same world, but most servers simply fill the runtime ID
	// of the entity out for this field.
	EntityUniqueID int64
}

// ID ...
func (*RemoveActor) ID() uint32 {
	return IDRemoveActor
}

func (pk *RemoveActor) Marshal(io protocol.IO) {
	io.Varint64(&pk.EntityUniqueID)
}
