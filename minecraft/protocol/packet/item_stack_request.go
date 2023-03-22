package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ItemStackRequest is sent by the client to change item stacks in an inventory. It is essentially a
// replacement of the InventoryTransaction packet added in 1.16 for inventory specific actions, such as moving
// items around or crafting. The InventoryTransaction packet is still used for actions such as placing blocks
// and interacting with entities.
type ItemStackRequest struct {
	// Requests holds a list of item stack requests. These requests are all separate, but the client buffers
	// the requests, so you might find multiple unrelated requests in this packet.
	Requests []protocol.ItemStackRequest
}

// ID ...
func (*ItemStackRequest) ID() uint32 {
	return IDItemStackRequest
}

// Marshal ...
func (pk *ItemStackRequest) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ItemStackRequest) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ItemStackRequest) marshal(r protocol.IO) {
	protocol.Slice(r, &pk.Requests)
}
