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

// Marshal ...
func (pk *CameraInstruction) Marshal(w *protocol.Writer) {
	w.NBT(&pk.Data, nbt.NetworkLittleEndian)
}

// Unmarshal ...
func (pk *CameraInstruction) Unmarshal(r *protocol.Reader) {
	r.NBT(&pk.Data, nbt.NetworkLittleEndian)
}
