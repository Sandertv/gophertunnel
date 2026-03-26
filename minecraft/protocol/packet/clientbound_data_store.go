package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundDataStore is sent by the server to update, change or remove data store entries on the client.
type ClientBoundDataStore struct {
	// Updates is an array of data store changes. Each entry has its own change type discriminator.
	Updates []protocol.DataStoreChangeEntry
}

// ID ...
func (*ClientBoundDataStore) ID() uint32 {
	return IDClientBoundDataStore
}

func (pk *ClientBoundDataStore) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Updates)
}
