package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// PrimitiveShapes is a packet sent by the server to instruct the client to render one or more shapes in the world.
// Shapes can be added, removed or updated based on the data provided individually.
type PrimitiveShapes struct {
	// Shapes is a list of shapes to draw on the client-side.
	Shapes []protocol.PrimitiveShape
}

// ID ...
func (pk *PrimitiveShapes) ID() uint32 {
	return IDPrimitiveShapes
}

func (pk *PrimitiveShapes) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Shapes)
}
