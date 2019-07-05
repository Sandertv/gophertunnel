package packet

import (
	"bytes"
	"encoding/binary"
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
	// UnknownInt32 ...
	UnknownInt32 int32
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
func (pk *UpdateEquip) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.WindowID)
	_ = binary.Write(buf, binary.LittleEndian, pk.WindowType)
	_ = protocol.WriteVarint32(buf, pk.UnknownInt32)
	_ = protocol.WriteVarint64(buf, pk.EntityUniqueID)
	_, _ = buf.Write(pk.SerialisedInventoryData)
}

// Unmarshal ...
func (pk *UpdateEquip) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.WindowID),
		binary.Read(buf, binary.LittleEndian, &pk.WindowType),
		protocol.Varint32(buf, &pk.UnknownInt32),
		protocol.Varint64(buf, &pk.EntityUniqueID),
	); err != nil {
		return err
	}
	pk.SerialisedInventoryData = make([]byte, buf.Len())
	copy(pk.SerialisedInventoryData, buf.Bytes())
	return nil
}
