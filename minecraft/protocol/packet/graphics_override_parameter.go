package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GraphicsOverrideParameter is sent by the server to override graphics parameters.
type GraphicsOverrideParameter struct {
	// Values is a list of parameter keyframe values.
	Values []protocol.ParameterKeyframeValue
	// BiomeIdentifier is the identifier of the biome for which the parameters apply.
	BiomeIdentifier string
	// ParameterType is the type of parameter being overridden.
	ParameterType uint8
	// Reset indicates whether to reset the parameters.
	Reset bool
}

// ID ...
func (*GraphicsOverrideParameter) ID() uint32 {
	return IDGraphicsOverrideParameter
}

func (pk *GraphicsOverrideParameter) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Values)
	io.String(&pk.BiomeIdentifier)
	io.Uint8(&pk.ParameterType)
	io.Bool(&pk.Reset)
}
