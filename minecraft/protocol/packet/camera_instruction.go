package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraInstruction gives a custom camera specific instructions to operate.
type CameraInstruction struct {
	// Set is a camera instruction that sets the camera to a specified preset.
	Set protocol.Optional[protocol.CameraInstructionSet]
	// Clear can be set to true to clear all the current camera instructions.
	Clear protocol.Optional[bool]
	// Fade is a camera instruction that fades the screen to a specified colour.
	Fade protocol.Optional[protocol.CameraInstructionFade]
}

// ID ...
func (*CameraInstruction) ID() uint32 {
	return IDCameraInstruction
}

func (pk *CameraInstruction) Marshal(io protocol.IO) {
	protocol.OptionalMarshaler(io, &pk.Set)
	protocol.OptionalFunc(io, &pk.Clear, io.Bool)
	protocol.OptionalMarshaler(io, &pk.Fade)
}
