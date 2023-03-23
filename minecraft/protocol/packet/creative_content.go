package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CreativeContent is a packet sent by the server to set the creative inventory's content for a player.
// Introduced in 1.16, this packet replaces the previous method - sending an InventoryContent packet with
// creative inventory window ID.
// As of v1.16.100, this packet must be sent during the login sequence. Not sending it will stop the client
// from joining the server.
type CreativeContent struct {
	// Items is a list of the items that should be added to the creative inventory.
	Items []protocol.CreativeItem
}

// ID ...
func (*CreativeContent) ID() uint32 {
	return IDCreativeContent
}

func (pk *CreativeContent) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Items)
}
