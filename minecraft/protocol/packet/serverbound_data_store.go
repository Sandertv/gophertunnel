package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerBoundDataStore is sent by the client to update a data store property on the server.
type ServerBoundDataStore struct {
	// Update contains the data store update.
	Update protocol.DataStoreUpdate
}

// ID ...
func (*ServerBoundDataStore) ID() uint32 {
	return IDServerBoundDataStore
}

func (pk *ServerBoundDataStore) Marshal(io protocol.IO) {
	protocol.Single(io, &pk.Update)
}
