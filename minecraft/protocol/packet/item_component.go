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

func (pk *ItemComponent) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Items)
}
