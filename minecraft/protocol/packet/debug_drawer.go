package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// DebugDrawer is a packet sent by the server to instruct the client to render
// one or more debug shapes. Each packet fully replaces any previously rendered shapes.
// To remove a shape, omit it from the next packet's Shapes slice.
type DebugDrawer struct {
	// Shapes is a list of shapes to draw on the client-side.
	Shapes []protocol.DebugDrawerShape
}

// ID ...
func (pk *DebugDrawer) ID() uint32 {
	return IDDebugDrawer
}

func (pk *DebugDrawer) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Shapes)
}
