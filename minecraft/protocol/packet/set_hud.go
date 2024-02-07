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
)

const (
	HudVisibilityHide = iota
	HudVisibilityReset
)

// SetHud is sent by the server to set the visibility of individual HUD elements on the client. It is
// important to note that the client does not reset the state of the HUD elements after it leaves a server,
// meaning they can leak into sessions on different servers. To be safe, you should reset the visibility of
// all HUD elements when a player connects.
type SetHud struct {
	// Elements is a list of HUD elements that are being modified. The values can be any of the HudElement
	// constants above.
	Elements []byte
	// Visibility represents the new visibility of the specified Elements. It can be any of the HudVisibility
	// constants above.
	Visibility byte
}

// ID ...
func (*SetHud) ID() uint32 {
	return IDSetHud
}

func (pk *SetHud) Marshal(io protocol.IO) {
	protocol.FuncSlice(io, &pk.Elements, io.Uint8)
	io.Uint8(&pk.Visibility)
}
