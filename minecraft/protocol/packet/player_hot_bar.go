package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerHotBar is sent by the server to the client. It used to be used to link hot bar slots of the player to
// actual slots in the inventory, but as of 1.2, this was changed and hot bar slots are no longer a free
// floating part of the inventory.
// Since 1.2, the packet has been re-purposed, but its new functionality is not clear.
type PlayerHotBar struct {
	// SelectedHotBarSlot ...
	SelectedHotBarSlot uint32
	// WindowID ...
	WindowID byte
	// SelectHotBarSlot ...
	SelectHotBarSlot bool
}

// ID ...
func (*PlayerHotBar) ID() uint32 {
	return IDPlayerHotBar
}

// Marshal ...
func (pk *PlayerHotBar) Marshal(w *protocol.Writer) {
	w.Varuint32(&pk.SelectedHotBarSlot)
	w.Uint8(&pk.WindowID)
	w.Bool(&pk.SelectHotBarSlot)
}

// Unmarshal ...
func (pk *PlayerHotBar) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&pk.SelectedHotBarSlot)
	r.Uint8(&pk.WindowID)
	r.Bool(&pk.SelectHotBarSlot)
}
