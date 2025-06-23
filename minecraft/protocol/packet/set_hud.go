package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	HudElementPaperDoll = iota
	HudElementArmour
	HudElementToolTips
	HudElementTouchControls
	HudElementCrosshair
	HudElementHotBar
	HudElementHealth
	HudElementProgressBar
	HudElementHunger
	HudElementAirBubbles
	HudElementHorseHealth
	HudElementStatusEffects
	HudElementItemText
)

const (
	HudVisibilityHide = iota
	HudVisibilityReset
)

// SetHud is sent by the server to set the visibility of individual HUD elements on the client.
type SetHud struct {
	// Elements is a list of HUD elements that are being modified. The values can be any of the HudElement
	// constants above.
	Elements []int32
	// Visibility represents the new visibility of the specified Elements. It can be any of the HudVisibility
	// constants above.
	Visibility int32
}

// ID ...
func (*SetHud) ID() uint32 {
	return IDSetHud
}

func (pk *SetHud) Marshal(io protocol.IO) {
	protocol.FuncSlice(io, &pk.Elements, io.Varint32)
	io.Varint32(&pk.Visibility)
}
