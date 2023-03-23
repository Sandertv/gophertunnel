package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraPresets gives the client a list of custom camera presets.
type CameraPresets struct {
	// Data is a compound tag of the presets being set. The structure of this tag is currently unknown.
	Data map[string]any
}

// ID ...
func (*CameraPresets) ID() uint32 {
	return IDCameraPresets
}

func (pk *CameraPresets) Marshal(io protocol.IO) {
	io.NBT(&pk.Data, nbt.NetworkLittleEndian)
}
