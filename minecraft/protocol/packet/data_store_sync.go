package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// DataStoreSync is sent by the server to synchronize data store information.
type DataStoreSync struct{}

// ID ...
func (*DataStoreSync) ID() uint32 {
	return IDDataStoreSync
}

func (pk *DataStoreSync) Marshal(io protocol.IO) {
	// TODO: Implement data store sync marshaling
}
