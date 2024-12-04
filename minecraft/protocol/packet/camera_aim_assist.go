package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CameraAimAssistActionSet = iota
	CameraAimAssistActionClear
)

// CameraAimAssist is sent by the server to the client to set up aim assist for the client's camera.
type CameraAimAssist struct {
	// Preset is the ID of the preset that has previously been defined in the CameraAimAssistPresets packet.
	Preset string
	// Angle is the maximum angle around the playes's cursor that the aim assist should check for a target,
	// if TargetMode is set to protocol.AimAssistTargetModeAngle.
	Angle mgl32.Vec2
	// Distance is the maximum distance from the player's cursor should check for a target, if TargetMode is
	// set to protocol.AimAssistTargetModeDistance.
	Distance float32
	// TargetMode is the mode that the camera should use for detecting targets. This is currently one of
	// protocol.AimAssistTargetModeAngle or protocol.AimAssistTargetModeDistance.
	TargetMode byte
	// Action is the action that should be performed with the aim assist. This is one of the constants above.
	Action byte
}

// ID ...
func (*CameraAimAssist) ID() uint32 {
	return IDCameraAimAssist
}

func (pk *CameraAimAssist) Marshal(io protocol.IO) {
	io.String(&pk.Preset)
	io.Vec2(&pk.Angle)
	io.Float32(&pk.Distance)
	io.Uint8(&pk.TargetMode)
	io.Uint8(&pk.Action)
}
