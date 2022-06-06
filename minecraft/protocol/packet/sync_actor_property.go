package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SyncActorProperty is an alternative to synced actor data.
type SyncActorProperty struct {
	// PropertyData ...
	PropertyData map[string]any
}

// ID ...
func (*SyncActorProperty) ID() uint32 {
	return IDSyncActorProperty
}

// Marshal ...
func (pk *SyncActorProperty) Marshal(w *protocol.Writer) {
	w.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
}

// Unmarshal ...
func (pk *SyncActorProperty) Unmarshal(r *protocol.Reader) {
	r.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
}
