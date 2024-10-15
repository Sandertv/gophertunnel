package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CameraAimAssistActionSet = iota
	CameraAimAssistActionClear
)

const (
	CameraAimAssistTargetModeAngle = iota
	CameraAimAssistTargetModeDistance
)

// CameraAimAssist is sent by the server to the client to set up aim assist for the client's camera.
type CameraAimAssist struct {
	// ViewAngle is the angle that the camera should aim at, if TargetMode is set to
	// CameraAimAssistTargetModeAngle.
	ViewAngle mgl32.Vec2
	// Distance is the distance that the camera should keep from the target, if TargetMode is set to
	// CameraAimAssistTargetModeDistance.
	Distance float32
	// TargetMode is the mode that the camera should use to aim at the target. This is one of the constants
	// above.
	TargetMode byte
	// Action is the action that should be performed with the aim assist. This is one of the constants above.
	Action byte
}

// ID ...
func (*CameraAimAssist) ID() uint32 {
	return IDCameraAimAssist
}

func (pk *CameraAimAssist) Marshal(io protocol.IO) {
	io.Vec2(&pk.ViewAngle)
	io.Float32(&pk.Distance)
	io.Uint8(&pk.TargetMode)
	io.Uint8(&pk.Action)
}
