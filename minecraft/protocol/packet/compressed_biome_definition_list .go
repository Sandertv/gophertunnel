package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type CompressedBiomeDefinitionList struct {
	SerialisedBiomeDefinitions []byte
}

// ID ...
func (*CompressedBiomeDefinitionList) ID() uint32 {
	return IDCompressedBiomeDefinitionList
}

func (pk *CompressedBiomeDefinitionList) Marshal(io protocol.IO) {
	io.Bytes(&pk.SerialisedBiomeDefinitions)
}
