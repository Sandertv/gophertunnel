package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type CameraAimAssistPresets struct {
	Categories []protocol.CameraAimAssistCategories
	Presets    []protocol.CameraAimAssistPreset
}

// ID ...
func (*CameraAimAssistPresets) ID() uint32 {
	return IDCameraAimAssistPresets
}

func (pk *CameraAimAssistPresets) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Categories)
	protocol.Slice(io, &pk.Presets)
}
