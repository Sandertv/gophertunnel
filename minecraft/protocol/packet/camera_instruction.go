package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraInstruction gives a custom camera specific instructions to operate.
type CameraInstruction struct {
	// Instructions is a list of the instructions that should be executed by the camera.
	Instructions []protocol.CameraInstruction
}

// ID ...
func (*CameraInstruction) ID() uint32 {
	return IDCameraInstruction
}

func (pk *CameraInstruction) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Instructions)
}
