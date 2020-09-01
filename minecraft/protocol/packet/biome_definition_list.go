package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BiomeDefinitionList is sent by the server to let the client know all biomes that are available and
// implemented on the server side. It is much like the AvailableActorIdentifiers packet, but instead
// functions for biomes.
type BiomeDefinitionList struct {
	// SerialisedBiomeDefinitions is a network NBT serialised compound of all definitions of biomes that are
	// available on the server.
	SerialisedBiomeDefinitions []byte
}

// ID ...
func (*BiomeDefinitionList) ID() uint32 {
	return IDBiomeDefinitionList
}

// Marshal ...
func (pk *BiomeDefinitionList) Marshal(w *protocol.Writer) {
	w.Bytes(&pk.SerialisedBiomeDefinitions)
}

// Unmarshal ...
func (pk *BiomeDefinitionList) Unmarshal(r *protocol.Reader) {
	r.Bytes(&pk.SerialisedBiomeDefinitions)
}
