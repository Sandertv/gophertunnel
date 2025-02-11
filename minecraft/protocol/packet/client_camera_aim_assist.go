package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientCameraAimAssistActionSet = iota
	ClientCameraAimAssistActionClear
)

// ClientCameraAimAssist is sent by the server to send a player animation from one player to all viewers of that player. It
// is used for a couple of actions, such as arm swimming and critical hits.
type ClientCameraAimAssist struct {
	// PresetID is the identifier of the preset to use which was previously defined in the CameraAimAssistPresets
	// packet.
	PresetID string
	// Action is the action to perform with the aim assist. It is one of the constants above.
	Action byte
	// AllowAimAssist specifies the client can use aim assist or not.
	AllowAimAssist bool
}

// ID ...
func (*ClientCameraAimAssist) ID() uint32 {
	return IDClientCameraAimAssist
}

func (pk *ClientCameraAimAssist) Marshal(io protocol.IO) {
	io.String(&pk.PresetID)
	io.Uint8(&pk.Action)
	io.Bool(&pk.AllowAimAssist)
}
