package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateBlockProperties is sent by the server to update the available block properties.
type UpdateBlockProperties struct {
	// SerialisedBlockProperties is a network little endian NBT serialised structure of the updated block
	// properties.
	SerialisedBlockProperties []byte
}

// ID ...
func (pk *UpdateBlockProperties) ID() uint32 {
	return IDUpdateBlockProperties
}

// Marshal ...
func (pk *UpdateBlockProperties) Marshal(w *protocol.Writer) {
	w.Bytes(&pk.SerialisedBlockProperties)
}

// Unmarshal ...
func (pk *UpdateBlockProperties) Unmarshal(r *protocol.Reader) {
	r.Bytes(&pk.SerialisedBlockProperties)
}
