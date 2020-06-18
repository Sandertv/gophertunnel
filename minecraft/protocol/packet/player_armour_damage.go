package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerArmourDamage is sent by the server to damage the armour of a player. It is a very efficient packet,
// but generally it's much easier to just send a slot update for the damaged armour.
type PlayerArmourDamage struct {
	// HelmetDamage is the amount of damage that should be dealt to the helmet.
	HelmetDamage int32
	// ChestplateDamage is the amount of damage that should be dealt to the chestplate.
	ChestplateDamage int32
	// LeggingsDamage is the amount of damage that should be dealt to the leggings.
	LeggingsDamage int32
	// BootsDamage is the amount of damage that should be dealt to the boots.
	BootsDamage int32
}

// ID ...
func (pk *PlayerArmourDamage) ID() uint32 {
	return IDPlayerArmourDamage
}

// Marshal ...
func (pk *PlayerArmourDamage) Marshal(buf *bytes.Buffer) {
	pk.HelmetDamage, pk.ChestplateDamage, pk.LeggingsDamage, pk.BootsDamage = 0, 0, 0, 0

	var flags byte
	if pk.HelmetDamage != 0 {
		flags |= 0b0001
	}
	if pk.ChestplateDamage != 0 {
		flags |= 0b0010
	}
	if pk.LeggingsDamage != 0 {
		flags |= 0b0100
	}
	if pk.BootsDamage != 0 {
		flags |= 0b1000
	}
	buf.WriteByte(flags)
	if pk.HelmetDamage != 0 {
		_ = protocol.WriteVarint32(buf, pk.HelmetDamage)
	}
	if pk.ChestplateDamage != 0 {
		_ = protocol.WriteVarint32(buf, pk.ChestplateDamage)
	}
	if pk.LeggingsDamage != 0 {
		_ = protocol.WriteVarint32(buf, pk.LeggingsDamage)
	}
	if pk.BootsDamage != 0 {
		_ = protocol.WriteVarint32(buf, pk.BootsDamage)
	}
}

// Unmarshal ...
func (pk *PlayerArmourDamage) Unmarshal(buf *bytes.Buffer) error {
	bitset, err := buf.ReadByte()
	if err != nil {
		return err
	}
	if bitset&0b0001 != 0 {
		if err := protocol.Varint32(buf, &pk.HelmetDamage); err != nil {
			return err
		}
	}
	if bitset&0b0010 != 0 {
		if err := protocol.Varint32(buf, &pk.ChestplateDamage); err != nil {
			return err
		}
	}
	if bitset&0b0100 != 0 {
		if err := protocol.Varint32(buf, &pk.LeggingsDamage); err != nil {
			return err
		}
	}
	if bitset&0b1000 != 0 {
		if err := protocol.Varint32(buf, &pk.BootsDamage); err != nil {
			return err
		}
	}
	return nil
}
