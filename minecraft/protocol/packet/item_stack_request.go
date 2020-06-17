package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ItemStackRequest is sent by the client to change item stacks in an inventory. It is essentially a
// replacement of the InventoryTransaction packet added in 1.16. Similarly to the InventoryTransaction packet,
// it is sent for actions such as crafting and placing blocks, but also for actions restricted to inventories
// such as moving items around.
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
func (pk *ItemStackRequest) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Requests)))
	for _, req := range pk.Requests {
		_ = protocol.WriteStackRequest(buf, req)
	}
}

// Unmarshal ...
func (pk *ItemStackRequest) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	if count > 64 {
		return protocol.LimitHitError{Limit: 64, Type: "ItemStackRequest"}
	}
	pk.Requests = make([]protocol.ItemStackRequest, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.StackRequest(buf, &pk.Requests[i]); err != nil {
			return err
		}
	}
	return nil
}
