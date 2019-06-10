package packet

import (
	"bytes"
	"encoding/binary"
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
func (pk *GUIDataPickItem) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.ItemName)
	_ = protocol.WriteString(buf, pk.ItemEffects)
	_ = binary.Write(buf, binary.LittleEndian, pk.HotBarSlot)
}

// Unmarshal ...
func (pk *GUIDataPickItem) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.ItemName),
		protocol.String(buf, &pk.ItemEffects),
		binary.Read(buf, binary.LittleEndian, &pk.HotBarSlot),
	)
}
