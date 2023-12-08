package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerToggleCrafterSlotRequest is sent by the client when it tries to toggle the state of a slot within a Crafter.
type PlayerToggleCrafterSlotRequest struct {
	// PosX is the X position of the Crafter that is being modified.
	PosX int32
	// PosY is the Y position of the Crafter that is being modified.
	PosY int32
	// PosZ is the Z position of the Crafter that is being modified.
	PosZ int32
	// Slot is the index of the slot that was toggled. This should be a value between 0 and 8.
	Slot byte
	// Disabled is the new state of the slot. If true, the slot is disabled, if false, the slot is enabled.
	Disabled bool
}

// ID ...
func (*PlayerToggleCrafterSlotRequest) ID() uint32 {
	return IDPlayerToggleCrafterSlotRequest
}

func (pk *PlayerToggleCrafterSlotRequest) Marshal(io protocol.IO) {
	io.Int32(&pk.PosX)
	io.Int32(&pk.PosY)
	io.Int32(&pk.PosZ)
	io.Uint8(&pk.Slot)
	io.Bool(&pk.Disabled)
}
