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

// Marshal ...
func (pk *CreativeContent) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *CreativeContent) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *CreativeContent) marshal(r protocol.IO) {
	protocol.Slice(r, &pk.Items)
}
