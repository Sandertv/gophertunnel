package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerToggleArmorVisibilityRequest is sent by the client to request that the server updates whether the
// player's equipped armor is hidden.
type PlayerToggleArmorVisibilityRequest struct {
	// HideArmor specifies if the player's equipped armor should be hidden.
	HideArmor bool
}

// ID ...
func (*PlayerToggleArmorVisibilityRequest) ID() uint32 {
	return IDPlayerToggleArmorVisibilityRequest
}

func (pk *PlayerToggleArmorVisibilityRequest) Marshal(io protocol.IO) {
	io.Bool(&pk.HideArmor)
}
