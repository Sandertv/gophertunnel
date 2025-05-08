package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BiomeDefinitionList is sent by the server to let the client know all biomes that are available and
// implemented on the server side. When enabled, it also includes information for the client to
// accurately recreate the server-side generation in vanilla worlds/servers for increased performance.
type BiomeDefinitionList struct {
	// BiomeDefinitions is a list of biomes that are available on the server.
	BiomeDefinitions []protocol.BiomeDefinition
	// StringList is a makeshift dictionary implementation Mojang created to try and reduce the size of the
	// overall packet. It is a list of common strings that are used in the biome definitions, such as
	// biome names, float values or query expressions.
	StringList []string
}

// ID ...
func (*BiomeDefinitionList) ID() uint32 {
	return IDBiomeDefinitionList
}

func (pk *BiomeDefinitionList) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.BiomeDefinitions)
	protocol.FuncSlice(io, &pk.StringList, io.String)
}
