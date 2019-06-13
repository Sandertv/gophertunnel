package packet

import (
	"bytes"
	"encoding/binary"
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
func (pk *PlayerHotBar) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, pk.SelectedHotBarSlot)
	_ = binary.Write(buf, binary.LittleEndian, pk.WindowID)
	_ = binary.Write(buf, binary.LittleEndian, pk.SelectHotBarSlot)
}

// Unmarshal ...
func (pk *PlayerHotBar) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint32(buf, &pk.SelectedHotBarSlot),
		binary.Read(buf, binary.LittleEndian, &pk.WindowID),
		binary.Read(buf, binary.LittleEndian, &pk.SelectHotBarSlot),
	)
}
