package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GraphicsOverrideParameter is sent by the server to override graphics parameters.
type GraphicsOverrideParameter struct {
	// Values is a list of parameter keyframe values.
	Values []protocol.ParameterKeyframeValue
	// FloatValue ...
	FloatValue float32
	// Vec3Value ...
	Vec3Value mgl32.Vec3
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
	io.Float32(&pk.FloatValue)
	io.Vec3(&pk.Vec3Value)
	io.String(&pk.BiomeIdentifier)
	io.Uint8(&pk.ParameterType)
	io.Bool(&pk.Reset)
}
