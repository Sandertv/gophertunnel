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
	w.String(&pk.Category)
	w.Varint32(&pk.Duration)
}

// Unmarshal ...
func (pk *ClientStartItemCooldown) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Category)
	r.Varint32(&pk.Duration)
}
