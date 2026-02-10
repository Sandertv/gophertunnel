package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// VoxelShapes is sent by the server to send voxel shape data to the client.
type VoxelShapes struct {
	// Shapes is a list of voxel shapes.
	Shapes []protocol.VoxelShape
	// NameMap is a map of shape names to IDs.
	NameMap []protocol.VoxelShapeNameEntry
}

// ID ...
func (*VoxelShapes) ID() uint32 {
	return IDVoxelShapes
}

func (pk *VoxelShapes) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Shapes)
	protocol.Slice(io, &pk.NameMap)
}
