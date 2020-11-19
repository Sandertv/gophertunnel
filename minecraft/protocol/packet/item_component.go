package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type ItemComponent struct {
	Items []protocol.ItemComponentEntry
}

// ID ...
func (pk ItemComponent) ID() uint32 {
	return IDItemComponent
}

// Marshal ...
func (pk ItemComponent) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Items))
	w.Varuint32(&l)
	for i := range pk.Items {
		w.String(&pk.Items[i].Name)
		w.NBT(&pk.Items[i].Data, nbt.NetworkLittleEndian)
	}
}

// Unmarshal ...
func (pk ItemComponent) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	pk.Items = make([]protocol.ItemComponentEntry, count)
	for i := uint32(0); i < count; i++ {
		r.String(&pk.Items[i].Name)
		r.NBT(&pk.Items[i].Data, nbt.NetworkLittleEndian)
	}
}
