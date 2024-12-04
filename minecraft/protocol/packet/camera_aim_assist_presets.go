package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraAimAssistPresets is sent by the server to the client to provide a list of categories and presets
// that can be used when sending a CameraAimAssist packet or a CameraInstruction including aim assist.
type CameraAimAssistPresets struct {
	// CategoryGroups is a list of groups of categories which can be referenced by one of the Presets.
	CategoryGroups []protocol.CameraAimAssistCategoryGroup
	// Presets is a list of presets which define a base for how aim assist should behave
	Presets []protocol.CameraAimAssistPreset
}

// ID ...
func (*CameraAimAssistPresets) ID() uint32 {
	return IDCameraAimAssistPresets
}

func (pk *CameraAimAssistPresets) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.CategoryGroups)
	protocol.Slice(io, &pk.Presets)
}
