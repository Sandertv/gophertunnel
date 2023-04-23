package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type TrimData struct {
	Patterns  []protocol.TrimPattern
	Materials []protocol.TrimMaterial
}

// ID ...
func (*TrimData) ID() uint32 {
	return IDTrimData
}

func (pk *TrimData) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Patterns)
	protocol.Slice(io, &pk.Materials)
}
