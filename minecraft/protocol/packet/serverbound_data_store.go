package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerboundDataStore is sent by the client to the server to store data.
type ServerboundDataStore struct {
	// DataStoreName is the name of the data store.
	DataStoreName string
	// Property is the property name.
	Property string
	// Path is the path to the data.
	Path string
	// Data is the data to store.
	Data []byte
	// UpdateCount is the update count.
	UpdateCount int32
}

// ID ...
func (*ServerboundDataStore) ID() uint32 {
	return IDServerboundDataStore
}

func (pk *ServerboundDataStore) Marshal(io protocol.IO) {
	io.String(&pk.DataStoreName)
	io.String(&pk.Property)
	io.String(&pk.Path)
	io.ByteSlice(&pk.Data)
	io.Varint32(&pk.UpdateCount)
}
