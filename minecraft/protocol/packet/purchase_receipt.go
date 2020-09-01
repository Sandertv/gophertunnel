package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PurchaseReceipt is sent by the client to the server to notify the server it purchased an item from the
// Marketplace store that was offered by the server. The packet is only used for partnered servers.
type PurchaseReceipt struct {
	// Receipts is a list of receipts, or proofs of purchases, for the offers that have been purchased by the
	// player.
	Receipts []string
}

// ID ...
func (*PurchaseReceipt) ID() uint32 {
	return IDPurchaseReceipt
}

// Marshal ...
func (pk *PurchaseReceipt) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Receipts))
	w.Varuint32(&l)
	for _, receipt := range pk.Receipts {
		w.String(&receipt)
	}
}

// Unmarshal ...
func (pk *PurchaseReceipt) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, 64)

	pk.Receipts = make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&pk.Receipts[i])
	}
}
