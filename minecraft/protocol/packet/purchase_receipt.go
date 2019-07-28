package packet

import (
	"bytes"
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
func (pk *PurchaseReceipt) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Receipts)))
	for _, receipt := range pk.Receipts {
		_ = protocol.WriteString(buf, receipt)
	}
}

// Unmarshal ...
func (pk *PurchaseReceipt) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	if count > 64 {
		return protocol.LimitHitError{Type: "purchase receipt", Limit: 64}
	}
	pk.Receipts = make([]string, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.String(buf, &pk.Receipts[i]); err != nil {
			return err
		}
	}
	return nil
}
