package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraInstruction gives a custom camera specific instructions to operate.
type CameraInstruction struct {
	// Data is a compound tag of the instructions to sent. The structure of this tag is currently unknown.
	Data map[string]any
}

// ID ...
func (*CameraInstruction) ID() uint32 {
	return IDCameraInstruction
}

func (pk *CameraInstruction) Marshal(io protocol.IO) {
	io.NBT(&pk.Data, nbt.NetworkLittleEndian)
}
