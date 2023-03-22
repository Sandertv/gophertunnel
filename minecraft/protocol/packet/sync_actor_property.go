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
	pk.marshal(w)
}

// Unmarshal ...
func (pk *SyncActorProperty) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *SyncActorProperty) marshal(r protocol.IO) {
	r.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)
}
