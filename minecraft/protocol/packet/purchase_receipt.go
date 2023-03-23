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

func (pk *PurchaseReceipt) Marshal(io protocol.IO) {
	protocol.FuncSlice(io, &pk.Receipts, io.String)
}
