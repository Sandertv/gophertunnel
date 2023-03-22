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
	pk.marshal(w)
}

// Unmarshal ...
func (pk *CameraInstruction) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *CameraInstruction) marshal(r protocol.IO) {
	r.NBT(&pk.Data, nbt.NetworkLittleEndian)
}
