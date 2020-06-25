package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ItemStackRequest represents a single request present in an ItemStackRequest packet sent by the client to
// change an item in an inventory.
// Item stack requests are either approved or rejected by the server using the ItemStackResponse packet.
type ItemStackRequest struct {
	// RequestID is a unique ID for the request. This ID is used by the server to send a response for this
	// specific request in the ItemStackResponse packet.
	RequestID int32
	// Actions is a list of actions performed by the client. The actual type of the actions depends on which
	// ID was present, and is one of the concrete types below.
	Actions []StackRequestAction
}

// WriteStackRequest writes an ItemStackRequest x to Buffer dst.
func WriteStackRequest(dst *bytes.Buffer, x ItemStackRequest) error {
	if err := WriteVarint32(dst, x.RequestID); err != nil {
		return err
	}
	if err := WriteVaruint32(dst, uint32(len(x.Actions))); err != nil {
		return err
	}
	for _, action := range x.Actions {
		var id byte
		switch action.(type) {
		case *TakeStackRequestAction:
			id = StackRequestActionTake
		case *PlaceStackRequestAction:
			id = StackRequestActionPlace
		case *SwapStackRequestAction:
			id = StackRequestActionSwap
		case *DropStackRequestAction:
			id = StackRequestActionDrop
		case *DestroyStackRequestAction:
			id = StackRequestActionDestroy
		case *ConsumeStackRequestAction:
			id = StackRequestActionConsume
		case *CreateStackRequestAction:
			id = StackRequestActionCreate
		case *LabTableCombineStackRequestAction:
			id = StackRequestActionLabTableCombine
		case *BeaconPaymentStackRequestAction:
			id = StackRequestActionBeaconPayment
		case *CraftRecipeStackRequestAction:
			id = StackRequestActionCraftRecipe
		case *AutoCraftRecipeStackRequestAction:
			id = StackRequestActionCraftRecipeAuto
		case *CraftCreativeStackRequestAction:
			id = StackRequestActionCraftCreative
		case *CraftNonImplementedStackRequestAction:
			id = StackRequestActionCraftNonImplementedDeprecated
		case *CraftResultsDeprecatedStackRequestAction:
			id = StackRequestActionCraftResultsDeprecated
		default:
			panic(fmt.Sprintf("unknown item stack request action type %T", action))
		}
		dst.WriteByte(id)
		action.Marshal(dst)
	}
	return nil
}

// StackRequest reads an ItemStackRequest x from Buffer src.
func StackRequest(src *bytes.Buffer, x *ItemStackRequest) error {
	var count uint32
	if err := Varint32(src, &x.RequestID); err != nil {
		return err
	}
	if err := Varuint32(src, &count); err != nil {
		return err
	}
	if count > mediumLimit {
		return LimitHitError{Limit: mediumLimit, Type: "ItemStackRequest"}
	}
	x.Actions = make([]StackRequestAction, count)
	for i := uint32(0); i < count; i++ {
		id, err := src.ReadByte()
		if err != nil {
			return wrap(err)
		}
		var action StackRequestAction
		switch id {
		case StackRequestActionTake:
			action = &TakeStackRequestAction{}
		case StackRequestActionPlace:
			action = &PlaceStackRequestAction{}
		case StackRequestActionSwap:
			action = &SwapStackRequestAction{}
		case StackRequestActionDrop:
			action = &DropStackRequestAction{}
		case StackRequestActionDestroy:
			action = &DestroyStackRequestAction{}
		case StackRequestActionConsume:
			action = &ConsumeStackRequestAction{}
		case StackRequestActionCreate:
			action = &CreateStackRequestAction{}
		case StackRequestActionLabTableCombine:
			action = &LabTableCombineStackRequestAction{}
		case StackRequestActionBeaconPayment:
			action = &BeaconPaymentStackRequestAction{}
		case StackRequestActionCraftRecipe:
			action = &CraftRecipeStackRequestAction{}
		case StackRequestActionCraftRecipeAuto:
			action = &AutoCraftRecipeStackRequestAction{}
		case StackRequestActionCraftCreative:
			action = &CraftCreativeStackRequestAction{}
		case StackRequestActionCraftNonImplementedDeprecated:
			action = &CraftNonImplementedStackRequestAction{}
		case StackRequestActionCraftResultsDeprecated:
			action = &CraftResultsDeprecatedStackRequestAction{}
		default:
			return fmt.Errorf("unknown stack request action %v", id)
		}
		if err := action.Unmarshal(src); err != nil {
			return err
		}
		x.Actions[i] = action
	}
	return nil
}

// ItemStackResponse is a response to an individual ItemStackRequest.
type ItemStackResponse struct {
	// Success specifies if the request with the RequestID below was successful. If this is the case, the
	// ContainerInfo below will have information on what slots ended up changing. If not, the container info
	// will be empty.
	Success bool
	// RequestID is the unique ID of the request that this response is in reaction to. If rejected, the client
	// will undo the actions from the request with this ID.
	RequestID int32
	// ContainerInfo holds information on the containers that had their contents changed as a result of the
	// request.
	ContainerInfo []StackResponseContainerInfo
}

// StackResponseContainerInfo holds information on what slots in a container have what item stack in them.
type StackResponseContainerInfo struct {
	// ContainerID is the container ID of the container that the slots that follow are in. For the main
	// inventory, this value seems to be 0x1b. For the cursor, this value seems to be 0x3a. For the crafting
	// grid, this value seems to be 0x0d.
	ContainerID byte
	// SlotInfo holds information on what item stack should be present in specific slots in the container.
	SlotInfo []StackResponseSlotInfo
}

// StackResponseSlotInfo holds information on what item stack should be present in a specific slot.
type StackResponseSlotInfo struct {
	// Slot and HotbarSlot seem to be the same value every time: The slot that was actually changed. I'm not
	// sure if these slots ever differ.
	Slot, HotbarSlot byte
	// Count is the total count of the item stack. This count will be shown client-side after the response is
	// sent to the client.
	Count byte
	// StackNetworkID is the network ID of the new stack at a specific slot.
	StackNetworkID int32
}

// WriteStackResponse writes an ItemStackResponse x to Buffer dst.
func WriteStackResponse(dst *bytes.Buffer, x ItemStackResponse) error {
	if err := chainErr(
		binary.Write(dst, binary.LittleEndian, x.Success),
		WriteVarint32(dst, x.RequestID),
	); err != nil {
		return err
	}
	if !x.Success {
		return nil
	}
	if err := WriteVaruint32(dst, uint32(len(x.ContainerInfo))); err != nil {
		return err
	}
	for _, info := range x.ContainerInfo {
		if err := WriteStackContainerInfo(dst, info); err != nil {
			return err
		}
	}
	return nil
}

// StackResponse reads an ItemStackResponse x from Buffer src.
func StackResponse(src *bytes.Buffer, x *ItemStackResponse) error {
	if err := chainErr(
		binary.Read(src, binary.LittleEndian, &x.Success),
		Varint32(src, &x.RequestID),
	); err != nil {
		return err
	}
	if !x.Success {
		return nil
	}
	var l uint32
	if err := Varuint32(src, &l); err != nil {
		return err
	}
	x.ContainerInfo = make([]StackResponseContainerInfo, l)
	for i := uint32(0); i < l; i++ {
		if err := StackContainerInfo(src, &x.ContainerInfo[i]); err != nil {
			return err
		}
	}
	return nil
}

// WriteStackContainerInfo writes a StackResponseContainerInfo x to Buffer dst.
func WriteStackContainerInfo(dst *bytes.Buffer, x StackResponseContainerInfo) error {
	dst.WriteByte(x.ContainerID)
	if err := WriteVaruint32(dst, uint32(len(x.SlotInfo))); err != nil {
		return err
	}
	for _, info := range x.SlotInfo {
		if err := WriteStackSlotInfo(dst, info); err != nil {
			return err
		}
	}
	return nil
}

// StackContainerInfo reads a StackResponseContainerInfo x from Buffer src.
func StackContainerInfo(src *bytes.Buffer, x *StackResponseContainerInfo) error {
	if err := binary.Read(src, binary.LittleEndian, &x.ContainerID); err != nil {
		return err
	}
	var l uint32
	if err := Varuint32(src, &l); err != nil {
		return err
	}
	x.SlotInfo = make([]StackResponseSlotInfo, l)
	for i := uint32(0); i < l; i++ {
		if err := StackSlotInfo(src, &x.SlotInfo[i]); err != nil {
			return err
		}
	}
	return nil
}

// WriteStackSlotInfo writes a StackResponseSlotInfo x to Buffer dst.
func WriteStackSlotInfo(dst *bytes.Buffer, x StackResponseSlotInfo) error {
	if x.Slot != x.HotbarSlot {
		panic(fmt.Errorf("%v: Slot and HotbarSlot had different values: %v vs %v", callFrame(), x.Slot, x.HotbarSlot))
	}
	return chainErr(
		binary.Write(dst, binary.LittleEndian, x.Slot),
		binary.Write(dst, binary.LittleEndian, x.HotbarSlot),
		binary.Write(dst, binary.LittleEndian, x.Count),
		WriteVarint32(dst, x.StackNetworkID),
	)
}

// StackSlotInfo reads a StackResponseSlotInfo x from Buffer src.
func StackSlotInfo(src *bytes.Buffer, x *StackResponseSlotInfo) error {
	if err := chainErr(
		binary.Read(src, binary.LittleEndian, &x.Slot),
		binary.Read(src, binary.LittleEndian, &x.HotbarSlot),
		binary.Read(src, binary.LittleEndian, &x.Count),
		Varint32(src, &x.StackNetworkID),
	); err != nil {
		return err
	}
	if x.Slot != x.HotbarSlot {
		return fmt.Errorf("%v: Slot and HotbarSlot had different values: %v vs %v", callFrame(), x.Slot, x.HotbarSlot)
	}
	return nil
}

// StackRequestAction represents a single action related to the inventory present in an ItemStackRequest.
// The action is one of the concrete types below, each of which are indicative of a different action by the
// client, such as moving an item around the inventory or placing a block.
type StackRequestAction interface {
	// Marshal encodes the stack request action its binary representation into buf.
	Marshal(buf *bytes.Buffer)
	// Unmarshal decodes a serialised stack request action object in buf into the InventoryTransactionData
	// instance.
	Unmarshal(buf *bytes.Buffer) error
}

const (
	StackRequestActionTake = iota
	StackRequestActionPlace
	StackRequestActionSwap
	StackRequestActionDrop
	StackRequestActionDestroy
	StackRequestActionConsume
	StackRequestActionCreate
	StackRequestActionLabTableCombine
	StackRequestActionBeaconPayment
	StackRequestActionCraftRecipe
	StackRequestActionCraftRecipeAuto
	StackRequestActionCraftCreative
	StackRequestActionCraftNonImplementedDeprecated
	StackRequestActionCraftResultsDeprecated
)

// transferStackRequestAction is the structure shared by StackRequestActions that transfer items from one
// slot into another.
type transferStackRequestAction struct {
	// Count is the count of the item in the source slot that was taken towards the destination slot.
	Count byte
	// Source and Destination point to the source slot from which Count of the item stack were taken and the
	// destination slot to which this item was moved.
	Source, Destination StackRequestSlotInfo
}

// Marshal ...
func (a *transferStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, a.Count)
	_ = WriteStackReqSlotInfo(buf, a.Source)
	_ = WriteStackReqSlotInfo(buf, a.Destination)
}

// Unmarshal ...
func (a *transferStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &a.Count),
		StackReqSlotInfo(buf, &a.Source),
		StackReqSlotInfo(buf, &a.Destination),
	)
}

// TakeStackRequestAction is sent by the client to the server to take x amount of items from one slot in a
// container to the cursor.
type TakeStackRequestAction struct {
	transferStackRequestAction
}

// PlaceStackRequestAction is sent by the client to the server to place x amount of items from one slot into
// another slot, such as when shift clicking an item in the inventory to move it around or when moving an item
// in the cursor into a slot.
type PlaceStackRequestAction struct {
	transferStackRequestAction
}

// SwapStackRequestAction is sent by the client to swap the item in its cursor with an item present in another
// container. The two item stacks swap places.
type SwapStackRequestAction struct {
	// Source and Destination point to the source slot from which Count of the item stack were taken and the
	// destination slot to which this item was moved.
	Source, Destination StackRequestSlotInfo
}

// Marshal ...
func (a *SwapStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = WriteStackReqSlotInfo(buf, a.Source)
	_ = WriteStackReqSlotInfo(buf, a.Destination)
}

// Unmarshal ...
func (a *SwapStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		StackReqSlotInfo(buf, &a.Source),
		StackReqSlotInfo(buf, &a.Destination),
	)
}

// DropStackRequestAction is sent by the client when it drops an item out of the inventory when it has its
// inventory opened. This action is not sent when a player drops an item out of the hotbar using the Q button
// (or the equivalent on mobile). The InventoryTransaction packet is still used for that action, regardless of
// whether the item stack network IDs are used or not.
type DropStackRequestAction struct {
	// Count is the count of the item in the source slot that was taken towards the destination slot.
	Count byte
	// Source is the source slot from which items were dropped to the ground.
	Source StackRequestSlotInfo
	// Randomly seems to be set to false in most cases. I'm not entirely sure what this does, but this is what
	// vanilla calls this field.
	Randomly bool
}

// Marshal ...
func (a *DropStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, a.Count)
	_ = WriteStackReqSlotInfo(buf, a.Source)
	_ = binary.Write(buf, binary.LittleEndian, a.Randomly)
}

// Unmarshal ...
func (a *DropStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &a.Count),
		StackReqSlotInfo(buf, &a.Source),
		binary.Read(buf, binary.LittleEndian, &a.Randomly),
	)
}

// DestroyStackRequestAction is sent by the client when it destroys an item in creative mode by moving it
// back into the creative inventory.
type DestroyStackRequestAction struct {
	// Count is the count of the item in the source slot that was destroyed.
	Count byte
	// Source is the source slot from which items came that were destroyed by moving them into the creative
	// inventory.
	Source StackRequestSlotInfo
}

// Marshal ...
func (a *DestroyStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, a.Count)
	_ = WriteStackReqSlotInfo(buf, a.Source)
}

// Unmarshal ...
func (a *DestroyStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &a.Count),
		StackReqSlotInfo(buf, &a.Source),
	)
}

// ConsumeStackRequestAction is sent by the client when it uses an item to craft another item. The original
// item is 'consumed'.
type ConsumeStackRequestAction struct {
	DestroyStackRequestAction
}

// CreateStackRequestAction is sent by the client when an item is created through being used as part of a
// recipe. For example, when milk is used to craft a cake, the buckets are leftover. The buckets are moved to
// the slot sent by the client here.
// Note that before this is sent, an action for consuming all items in the crafting table/grid is sent. Items
// that are not fully consumed when used for a recipe should not be destroyed there, but instead, should be
// turned into their respective resulting items.
type CreateStackRequestAction struct {
	// ResultsSlot is the slot in the inventory in which the results of the crafting ingredients are to be
	// placed.
	ResultsSlot byte
}

// Marshal ...
func (a *CreateStackRequestAction) Marshal(buf *bytes.Buffer) {
	buf.WriteByte(a.ResultsSlot)
}

// Unmarshal ...
func (a *CreateStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.LittleEndian, &a.ResultsSlot)
}

// LabTableCombineStackRequestAction is sent by the client when it uses a lab table to combine item stacks.
type LabTableCombineStackRequestAction struct{}

// Marshal ...
func (a *LabTableCombineStackRequestAction) Marshal(*bytes.Buffer) {}

// Unmarshal ...
func (a *LabTableCombineStackRequestAction) Unmarshal(*bytes.Buffer) error { return nil }

// BeaconPaymentStackRequestAction is sent by the client when it submits an item to enable effects from a
// beacon. These items will have been moved into the beacon item slot in advance.
type BeaconPaymentStackRequestAction struct {
	// PrimaryEffect and SecondaryEffect are the effects that were selected from the beacon.
	PrimaryEffect, SecondaryEffect int32
}

// Marshal ...
func (a *BeaconPaymentStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = WriteVarint32(buf, a.PrimaryEffect)
	_ = WriteVarint32(buf, a.SecondaryEffect)
}

// Unmarshal ...
func (a *BeaconPaymentStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		Varint32(buf, &a.PrimaryEffect),
		Varint32(buf, &a.SecondaryEffect),
	)
}

// CraftRecipeStackRequestAction is sent by the client the moment it begins crafting an item. This is the
// first action sent, before the Consume and Create item stack request actions.
// This action is also sent when an item is enchanted. Enchanting should be treated mostly the same way as
// crafting, where the old item is consumed.
type CraftRecipeStackRequestAction struct {
	// RecipeNetworkID is the network ID of the recipe that is about to be crafted. This network ID matches
	// one of the recipes sent in the CraftingData packet, where each of the recipes have a RecipeNetworkID as
	// of 1.16.
	RecipeNetworkID uint32
}

// Marshal ...
func (a *CraftRecipeStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = WriteVaruint32(buf, a.RecipeNetworkID)
}

// Unmarshal ...
func (a *CraftRecipeStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return Varuint32(buf, &a.RecipeNetworkID)
}

// AutoCraftRecipeStackRequestAction is sent by the client similarly to the CraftRecipeStackRequestAction. The
// only difference is that the recipe is automatically created and crafted by shift clicking the recipe book.
type AutoCraftRecipeStackRequestAction struct {
	CraftRecipeStackRequestAction
}

// CraftCreativeStackRequestAction is sent by the client when it takes an item out fo the creative inventory.
// The item is thus not really crafted, but instantly created.
type CraftCreativeStackRequestAction struct {
	// CreativeItemNetworkID is the network ID of the creative item that is being created. This is one of the
	// creative item network IDs sent in the CreativeContent packet.
	CreativeItemNetworkID uint32
}

// Marshal ...
func (a *CraftCreativeStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = WriteVaruint32(buf, a.CreativeItemNetworkID)
}

// Unmarshal ...
func (a *CraftCreativeStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	return Varuint32(buf, &a.CreativeItemNetworkID)
}

// CraftNonImplementedStackRequestAction is an action sent for inventory actions that aren't yet implemented
// in the new system. These include, for example, anvils.
type CraftNonImplementedStackRequestAction struct{}

// Marshal ...
func (*CraftNonImplementedStackRequestAction) Marshal(*bytes.Buffer) {}

// Unmarshal ...
func (*CraftNonImplementedStackRequestAction) Unmarshal(*bytes.Buffer) error { return nil }

// CraftResultsDeprecatedStackRequestAction is an additional, deprecated packet sent by the client after
// crafting. It holds the final results and the amount of times the recipe was crafted. It shouldn't be used.
// This action is also sent when an item is enchanted. Enchanting should be treated mostly the same way as
// crafting, where the old item is consumed.
type CraftResultsDeprecatedStackRequestAction struct {
	ResultItems  []ItemStack
	TimesCrafted byte
}

// Marshal ...
func (a *CraftResultsDeprecatedStackRequestAction) Marshal(buf *bytes.Buffer) {
	_ = WriteVaruint32(buf, uint32(len(a.ResultItems)))
	for _, i := range a.ResultItems {
		_ = WriteItem(buf, i)
	}
	buf.WriteByte(a.TimesCrafted)
}

// Unmarshal ...
func (a *CraftResultsDeprecatedStackRequestAction) Unmarshal(buf *bytes.Buffer) error {
	var l uint32
	if err := Varuint32(buf, &l); err != nil {
		return err
	}
	if l > higherLimit/2 {
		return LimitHitError{Limit: higherLimit / 2, Type: "CraftResultsDeprecated ResultItems"}
	}
	a.ResultItems = make([]ItemStack, l)
	for i := uint32(0); i < l; i++ {
		if err := Item(buf, &a.ResultItems[i]); err != nil {
			return err
		}
	}
	return binary.Read(buf, binary.LittleEndian, &a.TimesCrafted)
}

// StackRequestSlotInfo holds information on a specific slot client-side.
type StackRequestSlotInfo struct {
	// ContainerID is the ID of the container that the slot was in.
	ContainerID byte
	// Slot is the index of the slot within the container with the ContainerID above.
	Slot byte
	// StackNetworkID is the unique stack ID that the client assumes to be present in this slot. The server
	// must check if these IDs match. If they do not match, servers should reject the stack request that the
	// action holding this info was in.
	StackNetworkID int32
}

// WriteStackReqSlotInfo writes a StackRequestSlotInfo x to Buffer dst.
func WriteStackReqSlotInfo(dst *bytes.Buffer, x StackRequestSlotInfo) error {
	return chainErr(
		binary.Write(dst, binary.LittleEndian, x.ContainerID),
		binary.Write(dst, binary.LittleEndian, x.Slot),
		WriteVarint32(dst, x.StackNetworkID),
	)
}

// StackReqSlotInfo reads a StackRequestSlotInfo x from Buffer src.
func StackReqSlotInfo(src *bytes.Buffer, x *StackRequestSlotInfo) error {
	return chainErr(
		binary.Read(src, binary.LittleEndian, &x.ContainerID),
		binary.Read(src, binary.LittleEndian, &x.Slot),
		Varint32(src, &x.StackNetworkID),
	)
}
