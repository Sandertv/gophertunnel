package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateEquip is sent by the server to the client upon opening a horse inventory. It is used to set the
// content of the inventory and specify additional properties, such as the items that are allowed to be put
// in slots of the inventory.
type UpdateEquip struct {
	// WindowID is the identifier associated with the window that the UpdateEquip packet concerns. It is the
	// ID sent for the horse inventory that was opened before this packet was sent.
	WindowID byte
	// WindowType is the type of the window that was opened. Generally, this is the type of a horse inventory,
	// as the packet is specifically made for that.
	WindowType byte
	// Size is the size of the horse inventory that should be opened. A bigger size does, in fact, change the
	// amount of slots displayed.
	Size int32
	// EntityUniqueID is the unique ID of the entity whose equipment was 'updated' to the player. It is
	// typically the horse entity that had its inventory opened.
	EntityUniqueID int64
	// SerialisedInventoryData is a network NBT serialised compound holding the content of the inventory of
	// the entity (the equipment) and additional data such as the allowed items for a particular slot, used to
	// make sure only saddles can be put in the saddle slot etc.
	SerialisedInventoryData []byte
}

// ID ...
func (*UpdateEquip) ID() uint32 {
	return IDUpdateEquip
}

// Marshal ...
func (pk *UpdateEquip) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.WindowID)
	w.Uint8(&pk.WindowType)
	w.Varint32(&pk.Size)
	w.Varint64(&pk.EntityUniqueID)
	w.Bytes(&pk.SerialisedInventoryData)
}

// Unmarshal ...
func (pk *UpdateEquip) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.WindowID)
	r.Uint8(&pk.WindowType)
	r.Varint32(&pk.Size)
	r.Varint64(&pk.EntityUniqueID)
	r.Bytes(&pk.SerialisedInventoryData)
}
