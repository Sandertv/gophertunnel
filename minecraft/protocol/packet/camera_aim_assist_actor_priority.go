package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraAimAssistActorPriority is sent by the server to define actor-specific aim assist priorities.
type CameraAimAssistActorPriority struct {
	// PriorityData is a list of aim assist actor priority entries.
	PriorityData []protocol.CameraAimAssistActorPriorityData
}

// ID ...
func (*CameraAimAssistActorPriority) ID() uint32 {
	return IDCameraAimAssistActorPriority
}

func (pk *CameraAimAssistActorPriority) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.PriorityData)
}
