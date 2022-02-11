package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientStartItemCooldown is sent by the client to the server to initiate a cooldown on an item. The purpose of this
// packet isn't entirely clear.
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
func (c *ClientStartItemCooldown) Marshal(w *protocol.Writer) {
	w.String(&c.Category)
	w.Varint32(&c.Duration)
}

// Unmarshal ...
func (c *ClientStartItemCooldown) Unmarshal(r *protocol.Reader) {
	r.String(&c.Category)
	r.Varint32(&c.Duration)
}
