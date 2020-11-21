package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ItemComponent is sent by the server to attach client-side components to a custom item.
type ItemComponent struct {
	// Items holds a list of all custom items with their respective components set.
	Items []protocol.ItemComponentEntry
}

// ID ...
func (*ItemComponent) ID() uint32 {
	return IDItemComponent
}

// Marshal ...
func (pk *ItemComponent) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Items))
	w.Varuint32(&l)
	for i := range pk.Items {
		protocol.ItemComponents(w, &pk.Items[i])
	}
}

// Unmarshal ...
func (pk *ItemComponent) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	pk.Items = make([]protocol.ItemComponentEntry, count)
	for i := uint32(0); i < count; i++ {
		protocol.ItemComponents(r, &pk.Items[i])
	}
}
