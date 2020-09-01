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
	l := uint32(len(pk.Requests))
	w.Varuint32(&l)
	for _, req := range pk.Requests {
		protocol.WriteStackRequest(w, &req)
	}
}

// Unmarshal ...
func (pk *ItemStackRequest) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, 64)

	pk.Requests = make([]protocol.ItemStackRequest, count)
	for i := uint32(0); i < count; i++ {
		protocol.StackRequest(r, &pk.Requests[i])
	}
}
