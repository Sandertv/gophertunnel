package packet

import (
	"bytes"
)

// PlayerEnchantOptions is sent by the server to update the enchantment options displayed when the user opens
// the enchantment table. This packet was added in 1.16 and allows the server to decide on the enchantments
// that can be selected by the player.
type PlayerEnchantOptions struct {
}

// TODO

// ID ...
func (*PlayerEnchantOptions) ID() uint32 {
	return IDPlayerEnchantOptions
}

// Marshal ...
func (pk *PlayerEnchantOptions) Marshal(buf *bytes.Buffer) {

}

// Unmarshal ...
func (pk *PlayerEnchantOptions) Unmarshal(buf *bytes.Buffer) error {
	return nil
}
