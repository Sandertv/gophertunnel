package packet

import (
	"bytes"
	"encoding/binary"
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
	// LegacyRequestID is an ID that is only non-zero at times when sent by the client. The server should
	// always send 0 for this. When this field is not 0, the LegacySetItemSlots slice below will have values
	// in it.
	// LegacyRequestID ties in with the ItemStackResponse packet. If this field is non-0, the server should
	// respond with an ItemStackResponse packet. Some inventory actions such as dropping an item out of the
	// hotbar are still one using this packet, and the ItemStackResponse packet needs to tie in with it.
	LegacyRequestID int32
	// LegacySetItemSlots are only present if the LegacyRequestID is non-zero. These item slots inform the
	// server of the slots that were changed during the inventory transaction, and the server should send
	// back an ItemStackResponse packet with these slots present in it. (Or false with no slots, if rejected.)
	LegacySetItemSlots []protocol.LegacySetItemSlot
	// HasNetworkIDs specifies if the inventory actions below have network IDs associated with them. It is
	// always set to false when a client sends this packet to the server.
	HasNetworkIDs bool
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
	_ = protocol.WriteVarint32(buf, pk.LegacyRequestID)
	if pk.LegacyRequestID != 0 {
		_ = protocol.WriteVaruint32(buf, uint32(len(pk.LegacySetItemSlots)))
		for _, slot := range pk.LegacySetItemSlots {
			_ = protocol.WriteSetItemSlot(buf, slot)
		}
	}
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
	_ = binary.Write(buf, binary.LittleEndian, pk.HasNetworkIDs)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Actions)))
	for _, action := range pk.Actions {
		_ = protocol.WriteInvAction(buf, action, pk.HasNetworkIDs)
	}
	if pk.TransactionData != nil {
		pk.TransactionData.Marshal(buf)
	}
}

// Unmarshal ...
func (pk *InventoryTransaction) Unmarshal(buf *bytes.Buffer) error {
	var length, transactionType uint32
	if err := protocol.Varint32(buf, &pk.LegacyRequestID); err != nil {
		return err
	}
	if pk.LegacyRequestID != 0 {
		if err := protocol.Varuint32(buf, &length); err != nil {
			return err
		}
		pk.LegacySetItemSlots = make([]protocol.LegacySetItemSlot, length)
		for i := uint32(0); i < length; i++ {
			if err := protocol.SetItemSlot(buf, &pk.LegacySetItemSlots[i]); err != nil {
				return err
			}
		}
	}
	if err := chainErr(
		protocol.Varuint32(buf, &transactionType),
		binary.Read(buf, binary.LittleEndian, &pk.HasNetworkIDs),
		protocol.Varuint32(buf, &length),
	); err != nil {
		return err
	}
	if length > 512 {
		return protocol.LimitHitError{Type: "inventory transaction", Limit: 512}
	}
	pk.Actions = make([]protocol.InventoryAction, length)
	for i := uint32(0); i < length; i++ {
		// Each InventoryTransaction packet has a list of actions at the start, with a transaction data object
		// after that, depending on the transaction type.
		if err := protocol.InvAction(buf, &pk.Actions[i], pk.HasNetworkIDs); err != nil {
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
	return pk.TransactionData.Unmarshal(buf)
}
