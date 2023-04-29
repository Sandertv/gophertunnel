package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CompressedBiomeDefinitionList is sent by the server to send a list of biomes to the client. The contents of this packet
// are very large, even after being compressed. This packet is only required when using client-side chunk generation.
type CompressedBiomeDefinitionList struct {
	// Biomes is a map of biomes with their identifier as key, and the biome data as value. The biome data contains many
	// different fields such as climate, surface materials and generation rules etc.
	Biomes map[string]any
}

// ID ...
func (*CompressedBiomeDefinitionList) ID() uint32 {
	return IDCompressedBiomeDefinitionList
}

func (pk *CompressedBiomeDefinitionList) Marshal(io protocol.IO) {
	io.CompressedBiomeDefinitions(&pk.Biomes)
}
