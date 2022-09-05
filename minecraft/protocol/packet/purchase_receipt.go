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
	protocol.FuncSlice(w, &pk.Receipts, w.String)
}

// Unmarshal ...
func (pk *PurchaseReceipt) Unmarshal(r *protocol.Reader) {
	protocol.FuncSlice(r, &pk.Receipts, r.String)
}
