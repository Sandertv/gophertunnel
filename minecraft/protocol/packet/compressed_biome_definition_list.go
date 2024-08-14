package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CompressedBiomeDefinitionList is sent by the server to send a list of biomes to the client. The contents of this packet
// are very large, even after being compressed. This packet is only required when using client-side chunk generation.
type CompressedBiomeDefinitionList struct {
	// SerialisedBiomeDefinitions is a network NBT serialised compound of all definitions of biomes that are
	// available on the server.
	SerialisedBiomeDefinitions []byte
}

// ID ...
func (*CompressedBiomeDefinitionList) ID() uint32 {
	return IDCompressedBiomeDefinitionList
}

func (pk *CompressedBiomeDefinitionList) Marshal(io protocol.IO) {
	io.Bytes(&pk.SerialisedBiomeDefinitions)
}
