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

// Marshal ...
func (pk *CameraPresets) Marshal(w *protocol.Writer) {
	w.NBT(&pk.Data, nbt.NetworkLittleEndian)
}

// Unmarshal ...
func (pk *CameraPresets) Unmarshal(r *protocol.Reader) {
	r.NBT(&pk.Data, nbt.NetworkLittleEndian)
}
