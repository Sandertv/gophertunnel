package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// TakeItemEntity is sent by the server when a player picks up an item entity. It makes the item entity
// disappear to viewers and shows the pick-up animation.
type TakeItemEntity struct {
	// ItemEntityRuntimeID is the entity runtime ID of the item that is being taken by another entity. It will
	// disappear to viewers after showing the pick-up animation.
	ItemEntityRuntimeID uint64
	// TakerEntityRuntimeID is the runtime ID of the entity that took the item, which is usually a player, but
	// could be another entity like a zombie too.
	TakerEntityRuntimeID uint64
}

// ID ...
func (*TakeItemEntity) ID() uint32 {
	return IDTakeItemEntity
}

// Marshal ...
func (pk *TakeItemEntity) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.ItemEntityRuntimeID)
	_ = protocol.WriteVaruint64(buf, pk.TakerEntityRuntimeID)
}

// Unmarshal ...
func (pk *TakeItemEntity) Unmarshal(buf *bytes.Buffer) error {
	return ChainErr(
		protocol.Varuint64(buf, &pk.ItemEntityRuntimeID),
		protocol.Varuint64(buf, &pk.TakerEntityRuntimeID),
	)
}
