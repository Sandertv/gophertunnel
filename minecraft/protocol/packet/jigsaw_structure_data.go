package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// JigsawStructureData is sent by the server to let the client know all the rules for jigsaw structures.
type JigsawStructureData struct {
	// StructureData is a network NBT serialised compound of all the jigsaw structure rules defined
	// on the server.
	StructureData []byte
}

// ID ...
func (*JigsawStructureData) ID() uint32 {
	return IDJigsawStructureData
}

func (pk *JigsawStructureData) Marshal(io protocol.IO) {
	io.Bytes(&pk.StructureData)
}
