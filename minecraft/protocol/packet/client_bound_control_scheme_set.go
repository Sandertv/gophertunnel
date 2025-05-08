package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ControlSchemeLockedPlayerRelativeStrafe = iota
	ControlSchemeCameraRelative
	ControlSchemeCameraRelativeStrafe
	ControlSchemePlayerRelative
	ControlSchemePlayerRelativeStrafe
)

// ClientBoundControlSchemeSet is sent by the server upon the client's request or the usage of the vanilla
// /controlscheme command. It is used to set the control scheme of the client, often used in combination with
// custom cameras.
type ClientBoundControlSchemeSet struct {
	// ControlScheme is the control scheme that the client should use. It is one of the following:
	//  - ControlSchemeLockedPlayerRelativeStrafe is the default behaviour, this cannot be set when the client
	//    is in a custom camera.
	//  - ControlSchemeCameraRelative makes movement relative to the camera's transform, with the client's
	//    rotation being relative to the client's movement.
	//  - ControlSchemeCameraRelativeStrafe makes movement relative to the camera's transform, with the
	//    client's rotation being locked.
	//  - ControlSchemePlayerRelative makes movement relative to the player's transform, meaning holding
	//    left/right will make the player turn in a circle.
	//  - ControlSchemePlayerRelativeStrafe makes movement the same as the default behaviour, but can be
	//    used in a custom camera.
	ControlScheme byte
}

// ID ...
func (*ClientBoundControlSchemeSet) ID() uint32 {
	return IDClientBoundControlSchemeSet
}

func (pk *ClientBoundControlSchemeSet) Marshal(io protocol.IO) {
	io.Uint8(&pk.ControlScheme)
}
