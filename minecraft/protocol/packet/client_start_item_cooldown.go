package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientStartItemCooldown is sent by the server to the client to initiate a cooldown on an item.
type ClientStartItemCooldown struct {
	// Category is the category of the item to start the cooldown on.
	Category string
	// Duration is the duration of ticks the cooldown should last.
	Duration int32
}

// ID ...
func (*ClientStartItemCooldown) ID() uint32 {
	return IDClientStartItemCooldown
}

// Marshal ...
func (pk *ClientStartItemCooldown) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ClientStartItemCooldown) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ClientStartItemCooldown) marshal(r protocol.IO) {
	r.String(&pk.Category)
	r.Varint32(&pk.Duration)
}
