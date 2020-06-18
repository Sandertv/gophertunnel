package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerEnchantOptions is sent by the server to update the enchantment options displayed when the user opens
// the enchantment table and puts an item in. This packet was added in 1.16 and allows the server to decide on
// the enchantments that can be selected by the player.
// The PlayerEnchantOptions packet should be sent once for every slot update of the enchantment table. The
// vanilla server sends an empty PlayerEnchantOptions packet when the player opens the enchantment table
// (air is present in the enchantment table slot) and sends the packet with actual enchantments in it when
// items are put in that can have enchantments.
type PlayerEnchantOptions struct {
	// Options is a list of possible enchantment options for the item that was put into the enchantment table.
	Options []protocol.EnchantmentOption
}

// ID ...
func (*PlayerEnchantOptions) ID() uint32 {
	return IDPlayerEnchantOptions
}

// Marshal ...
func (pk *PlayerEnchantOptions) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Options)))
	for _, option := range pk.Options {
		_ = protocol.WriteEnchantOption(buf, option)
	}
}

// Unmarshal ...
func (pk *PlayerEnchantOptions) Unmarshal(buf *bytes.Buffer) error {
	var l uint32
	if err := protocol.Varuint32(buf, &l); err != nil {
		return err
	}
	pk.Options = make([]protocol.EnchantmentOption, l)
	for i := uint32(0); i < l; i++ {
		if err := protocol.EnchantOption(buf, &pk.Options[i]); err != nil {
			return err
		}
	}
	return nil
}
