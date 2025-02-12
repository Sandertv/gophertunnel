package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ItemRegistry is sent by the server to send the client a list of available items and attach client-side
// components to a custom item. This packet was formerly known as the ItemComponent packet before 1.21.60,
// which did not include item definitions but only the components.
type ItemRegistry struct {
	// Items is a list of all items with their legacy IDs which are available in the game. Failing to send any
	// of the items that are in the game will crash mobile clients. Any custom components are also attached to
	// the items in this list.
	Items []protocol.ItemEntry
}

// ID ...
func (*ItemRegistry) ID() uint32 {
	return IDItemRegistry
}

func (pk *ItemRegistry) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Items)
}
