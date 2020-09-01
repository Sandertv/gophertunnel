package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GUIDataPickItem is sent by the server to make the client 'select' a hot bar slot. It currently appears to
// be broken however, and does not actually set the selected slot to the hot bar slot set in the packet.
type GUIDataPickItem struct {
	// ItemName is the name of the item that shows up in the top part of the popup that shows up when
	// selecting an item. It is shown as if an item was selected by the player itself.
	ItemName string
	// ItemEffects is the line under the ItemName, where the effects of the item are usually situated.
	ItemEffects string
	// HotBarSlot is the hot bar slot to be selected/picked. This does not currently work, so it does not
	// matter what number this is.
	HotBarSlot int32
}

// ID ...
func (*GUIDataPickItem) ID() uint32 {
	return IDGUIDataPickItem
}

// Marshal ...
func (pk *GUIDataPickItem) Marshal(w *protocol.Writer) {
	w.String(&pk.ItemName)
	w.String(&pk.ItemEffects)
	w.Int32(&pk.HotBarSlot)
}

// Unmarshal ...
func (pk *GUIDataPickItem) Unmarshal(r *protocol.Reader) {
	r.String(&pk.ItemName)
	r.String(&pk.ItemEffects)
	r.Int32(&pk.HotBarSlot)
}
