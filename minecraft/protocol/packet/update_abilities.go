package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateAbilities is a packet sent from the server to the client to update the abilities of the player. It, along with
// the UpdateAdventureSettings packet, are replacements of the AdventureSettings packet since v1.19.10.
type UpdateAbilities struct {
	// EntityUniqueID is the unique ID of the player. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	EntityUniqueID int64
	// PlayerPermissions is the permission level of the player. It is a value from 0-3, with 0 being visitor,
	// 1 being member, 2 being operator and 3 being custom.
	PlayerPermissions uint8
	// CommandPermissions is a permission level that specifies the kind of commands that the player is
	// allowed to use. It is one of the CommandPermissionLevel constants in the AdventureSettings packet.
	CommandPermissions uint8
	// Layers contains all ability layers and their potential values. This should at least have one entry, being the
	// base layer.
	Layers []protocol.AbilityLayer
}

// ID ...
func (*UpdateAbilities) ID() uint32 {
	return IDUpdateAbilities
}

// Marshal ...
func (pk *UpdateAbilities) Marshal(w *protocol.Writer) {
	w.Int64(&pk.EntityUniqueID)
	w.Uint8(&pk.PlayerPermissions)
	w.Uint8(&pk.CommandPermissions)
	layersLen := uint8(len(pk.Layers))
	w.Uint8(&layersLen)
	for _, layer := range pk.Layers {
		protocol.SerializedLayer(w, &layer)
	}
}

// Unmarshal ...
func (pk *UpdateAbilities) Unmarshal(r *protocol.Reader) {
	r.Int64(&pk.EntityUniqueID)
	r.Uint8(&pk.PlayerPermissions)
	r.Uint8(&pk.CommandPermissions)
	var layersLen uint8
	r.Uint8(&layersLen)
	pk.Layers = make([]protocol.AbilityLayer, layersLen)
	for i := uint8(0); i < layersLen; i++ {
		protocol.SerializedLayer(r, &pk.Layers[i])
	}
}
