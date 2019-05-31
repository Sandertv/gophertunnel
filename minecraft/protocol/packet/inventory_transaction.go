package packet

import (
	"bytes"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	InventoryTransactionTypeNormal = iota
	InventoryTransactionTypeMismatch
	InventoryTransactionTypeUseItem
	InventoryTransactionTypeUseItemOnEntity
	InventoryTransactionTypeReleaseItem
)

// InventoryTransaction is a packet sent by the client. It essentially exists out of multiple sub-packets,
// each of which have something to do with the inventory in one way or another. Some of these sub-packets
// directly relate to the inventory, others relate to interaction with the world, that could potentially
// result in a change in the inventory.
type InventoryTransaction struct {
	// Actions is a list of actions that took place, that form the inventory transaction together. Each of
	// these actions hold one slot in which one item was changed to another. In general, the combination of
	// all of these actions results in a balanced inventory transaction. This should be checked to ensure that
	// no items are cheated into the inventory.
	Actions []protocol.InventoryAction
	// TransactionData is a data object that holds data specific to the type of transaction that the
	// TransactionPacket held. Its concrete type must be one of NormalTransactionData, MismatchTransactionData
	// UseItemTransactionData, UseItemOnEntityTransactionData or ReleaseItemTransactionData. If nil is set,
	// the transaction will be assumed to of type InventoryTransactionTypeNormal.
	TransactionData protocol.InventoryTransactionData
}

// ID ...
func (*InventoryTransaction) ID() uint32 {
	return IDInventoryTransaction
}

// Marshal ...
func (pk *InventoryTransaction) Marshal(buf *bytes.Buffer) {
	switch pk.TransactionData.(type) {
	case nil, *protocol.NormalTransactionData:
		_ = protocol.WriteVaruint32(buf, InventoryTransactionTypeNormal)
	case *protocol.MismatchTransactionData:
		_ = protocol.WriteVaruint32(buf, InventoryTransactionTypeMismatch)
	case *protocol.UseItemTransactionData:
		_ = protocol.WriteVaruint32(buf, InventoryTransactionTypeUseItem)
	case *protocol.UseItemOnEntityTransactionData:
		_ = protocol.WriteVaruint32(buf, InventoryTransactionTypeUseItemOnEntity)
	case *protocol.ReleaseItemTransactionData:
		_ = protocol.WriteVaruint32(buf, InventoryTransactionTypeReleaseItem)
	}
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Actions)))
	for _, action := range pk.Actions {
		_ = protocol.WriteInvAction(buf, action)
	}
	if pk.TransactionData != nil {
		pk.TransactionData.Marshal(buf)
	}
}

// Unmarshal ...
func (pk *InventoryTransaction) Unmarshal(buf *bytes.Buffer) error {
	var length, transactionType uint32
	if err := chainErr(
		protocol.Varuint32(buf, &transactionType),
		protocol.Varuint32(buf, &length),
	); err != nil {
		return err
	}
	pk.Actions = make([]protocol.InventoryAction, length)
	for i := uint32(0); i < length; i++ {
		// Each InventoryTransaction packet has a list of actions at the start, with a transaction data object
		// after that, depending on the transaction type.
		if err := protocol.InvAction(buf, &pk.Actions[i]); err != nil {
			return err
		}
	}
	switch transactionType {
	case InventoryTransactionTypeNormal:
		pk.TransactionData = &protocol.NormalTransactionData{}
	case InventoryTransactionTypeMismatch:
		pk.TransactionData = &protocol.MismatchTransactionData{}
	case InventoryTransactionTypeUseItem:
		pk.TransactionData = &protocol.UseItemTransactionData{}
	case InventoryTransactionTypeUseItemOnEntity:
		pk.TransactionData = &protocol.UseItemOnEntityTransactionData{}
	case InventoryTransactionTypeReleaseItem:
		pk.TransactionData = &protocol.ReleaseItemTransactionData{}
	default:
		// We don't try to decode transactions that have some transaction type we don't know.
		return fmt.Errorf("unknown inventory transaction type %v", transactionType)
	}
	if err := pk.TransactionData.Unmarshal(buf); err != nil {
		return err
	}
	return nil
}
