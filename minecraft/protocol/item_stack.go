package protocol

import (
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
	// FilterStrings is a list of filter strings involved in the request. This is typically filled with one string
	// when an anvil or cartography is used.
	FilterStrings []string
}

// WriteStackRequest writes an ItemStackRequest x to Writer w.
func WriteStackRequest(w *Writer, x *ItemStackRequest) {
	l := uint32(len(x.Actions))
	w.Varint32(&x.RequestID)
	w.Varuint32(&l)
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
		case *MineBlockStackRequestAction:
			id = StackRequestActionMineBlock
		case *CraftRecipeStackRequestAction:
			id = StackRequestActionCraftRecipe
		case *AutoCraftRecipeStackRequestAction:
			id = StackRequestActionCraftRecipeAuto
		case *CraftCreativeStackRequestAction:
			id = StackRequestActionCraftCreative
		case *CraftRecipeOptionalStackRequestAction:
			id = StackRequestActionCraftRecipeOptional
		case *CraftGrindstoneRecipeStackRequestAction:
			id = StackRequestActionCraftGrindstone
		case *CraftLoomRecipeStackRequestAction:
			id = StackRequestActionCraftLoom
		case *CraftNonImplementedStackRequestAction:
			id = StackRequestActionCraftNonImplementedDeprecated
		case *CraftResultsDeprecatedStackRequestAction:
			id = StackRequestActionCraftResultsDeprecated
		default:
			w.UnknownEnumOption(fmt.Sprintf("%T", action), "stack request action type")
		}
		w.Uint8(&id)
		action.Marshal(w)
	}
	l = uint32(len(x.FilterStrings))
	w.Varuint32(&l)
	for _, n := range x.FilterStrings {
		w.String(&n)
	}
}

// StackRequest reads an ItemStackRequest x from Reader r.
func StackRequest(r *Reader, x *ItemStackRequest) {
	var count uint32
	r.Varint32(&x.RequestID)
	r.Varuint32(&count)
	r.LimitUint32(count, mediumLimit)

	x.Actions = make([]StackRequestAction, count)
	for i := uint32(0); i < count; i++ {
		var id uint8
		r.Uint8(&id)

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
		case StackRequestActionPlaceInContainer:
			action = &PlaceInContainerStackRequestAction{}
		case StackRequestActionTakeOutContainer:
			action = &TakeOutContainerStackRequestAction{}
		case StackRequestActionLabTableCombine:
			action = &LabTableCombineStackRequestAction{}
		case StackRequestActionBeaconPayment:
			action = &BeaconPaymentStackRequestAction{}
		case StackRequestActionMineBlock:
			action = &MineBlockStackRequestAction{}
		case StackRequestActionCraftRecipe:
			action = &CraftRecipeStackRequestAction{}
		case StackRequestActionCraftRecipeAuto:
			action = &AutoCraftRecipeStackRequestAction{}
		case StackRequestActionCraftCreative:
			action = &CraftCreativeStackRequestAction{}
		case StackRequestActionCraftRecipeOptional:
			action = &CraftRecipeOptionalStackRequestAction{}
		case StackRequestActionCraftGrindstone:
			action = &CraftGrindstoneRecipeStackRequestAction{}
		case StackRequestActionCraftLoom:
			action = &CraftLoomRecipeStackRequestAction{}
		case StackRequestActionCraftNonImplementedDeprecated:
			action = &CraftNonImplementedStackRequestAction{}
		case StackRequestActionCraftResultsDeprecated:
			action = &CraftResultsDeprecatedStackRequestAction{}
		default:
			r.UnknownEnumOption(id, "stack request action type")
			return
		}
		action.Unmarshal(r)
		x.Actions[i] = action
	}

	r.Varuint32(&count)
	r.LimitUint32(count, 64)

	x.FilterStrings = make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&x.FilterStrings[i])
	}
}

const (
	ItemStackResponseStatusOK = iota
	ItemStackResponseStatusError
	// There are lots more of these statuses for specific errors, but they don't seem to be very useful.
)

// ItemStackResponse is a response to an individual ItemStackRequest.
type ItemStackResponse struct {
	// Status specifies if the request with the RequestID below was successful. If this is the case, the
	// ContainerInfo below will have information on what slots ended up changing. If not, the container info
	// will be empty.
	// A non-0 status means an error occurred and will result in the action being reverted.
	Status uint8
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
	// CustomName is the custom name of the item stack. It is used in relation to text filtering.
	CustomName string
	// DurabilityCorrection is the current durability of the item stack. This durability will be shown
	// client-side after the response is sent to the client.
	DurabilityCorrection int32
}

// WriteStackResponse writes an ItemStackResponse x to Writer w.
func WriteStackResponse(w *Writer, x *ItemStackResponse) {
	w.Uint8(&x.Status)
	w.Varint32(&x.RequestID)
	if x.Status != ItemStackResponseStatusOK {
		return
	}
	l := uint32(len(x.ContainerInfo))
	w.Varuint32(&l)
	for _, info := range x.ContainerInfo {
		WriteStackContainerInfo(w, &info)
	}
}

// StackResponse reads an ItemStackResponse x from Reader r.
func StackResponse(r *Reader, x *ItemStackResponse) {
	var l uint32
	r.Uint8(&x.Status)
	r.Varint32(&x.RequestID)
	if x.Status != ItemStackResponseStatusOK {
		return
	}
	r.Varuint32(&l)

	x.ContainerInfo = make([]StackResponseContainerInfo, l)
	for i := uint32(0); i < l; i++ {
		StackContainerInfo(r, &x.ContainerInfo[i])
	}
}

// WriteStackContainerInfo writes a StackResponseContainerInfo x to Writer w.
func WriteStackContainerInfo(w *Writer, x *StackResponseContainerInfo) {
	w.Uint8(&x.ContainerID)
	l := uint32(len(x.SlotInfo))
	w.Varuint32(&l)
	for _, info := range x.SlotInfo {
		StackSlotInfo(w, &info)
	}
}

// StackContainerInfo reads a StackResponseContainerInfo x from Reader r.
func StackContainerInfo(r *Reader, x *StackResponseContainerInfo) {
	var l uint32
	r.Uint8(&x.ContainerID)
	r.Varuint32(&l)

	x.SlotInfo = make([]StackResponseSlotInfo, l)
	for i := uint32(0); i < l; i++ {
		StackSlotInfo(r, &x.SlotInfo[i])
	}
}

// StackSlotInfo reads/writes a StackResponseSlotInfo x using IO r.
func StackSlotInfo(r IO, x *StackResponseSlotInfo) {
	r.Uint8(&x.Slot)
	r.Uint8(&x.HotbarSlot)
	r.Uint8(&x.Count)
	r.Varint32(&x.StackNetworkID)
	if x.Slot != x.HotbarSlot {
		r.InvalidValue(x.HotbarSlot, "hotbar slot", "hot bar slot must be equal to normal slot")
	}
	r.String(&x.CustomName)
	r.Varint32(&x.DurabilityCorrection)
}

// StackRequestAction represents a single action related to the inventory present in an ItemStackRequest.
// The action is one of the concrete types below, each of which are indicative of a different action by the
// client, such as moving an item around the inventory or placing a block.
type StackRequestAction interface {
	// Marshal encodes the stack request action its binary representation into buf.
	Marshal(w *Writer)
	// Unmarshal decodes a serialised stack request action object from Reader r into the
	// InventoryTransactionData instance.
	Unmarshal(r *Reader)
}

const (
	StackRequestActionTake = iota
	StackRequestActionPlace
	StackRequestActionSwap
	StackRequestActionDrop
	StackRequestActionDestroy
	StackRequestActionConsume
	StackRequestActionCreate
	StackRequestActionPlaceInContainer
	StackRequestActionTakeOutContainer
	StackRequestActionLabTableCombine
	StackRequestActionBeaconPayment
	StackRequestActionMineBlock
	StackRequestActionCraftRecipe
	StackRequestActionCraftRecipeAuto
	StackRequestActionCraftCreative
	StackRequestActionCraftRecipeOptional
	StackRequestActionCraftGrindstone
	StackRequestActionCraftLoom
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
func (a *transferStackRequestAction) Marshal(w *Writer) {
	w.Uint8(&a.Count)
	StackReqSlotInfo(w, &a.Source)
	StackReqSlotInfo(w, &a.Destination)
}

// Unmarshal ...
func (a *transferStackRequestAction) Unmarshal(r *Reader) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
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
func (a *SwapStackRequestAction) Marshal(w *Writer) {
	StackReqSlotInfo(w, &a.Source)
	StackReqSlotInfo(w, &a.Destination)
}

// Unmarshal ...
func (a *SwapStackRequestAction) Unmarshal(r *Reader) {
	StackReqSlotInfo(r, &a.Source)
	StackReqSlotInfo(r, &a.Destination)
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
func (a *DropStackRequestAction) Marshal(w *Writer) {
	w.Uint8(&a.Count)
	StackReqSlotInfo(w, &a.Source)
	w.Bool(&a.Randomly)
}

// Unmarshal ...
func (a *DropStackRequestAction) Unmarshal(r *Reader) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
	r.Bool(&a.Randomly)
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
func (a *DestroyStackRequestAction) Marshal(w *Writer) {
	w.Uint8(&a.Count)
	StackReqSlotInfo(w, &a.Source)
}

// Unmarshal ...
func (a *DestroyStackRequestAction) Unmarshal(r *Reader) {
	r.Uint8(&a.Count)
	StackReqSlotInfo(r, &a.Source)
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
func (a *CreateStackRequestAction) Marshal(w *Writer) {
	w.Uint8(&a.ResultsSlot)
}

// Unmarshal ...
func (a *CreateStackRequestAction) Unmarshal(r *Reader) {
	r.Uint8(&a.ResultsSlot)
}

// PlaceInContainerStackRequestAction currently has no known purpose.
type PlaceInContainerStackRequestAction struct {
	transferStackRequestAction
}

// TakeOutContainerStackRequestAction currently has no known purpose.
type TakeOutContainerStackRequestAction struct {
	transferStackRequestAction
}

// LabTableCombineStackRequestAction is sent by the client when it uses a lab table to combine item stacks.
type LabTableCombineStackRequestAction struct{}

// Marshal ...
func (a *LabTableCombineStackRequestAction) Marshal(*Writer) {}

// Unmarshal ...
func (a *LabTableCombineStackRequestAction) Unmarshal(*Reader) {}

// BeaconPaymentStackRequestAction is sent by the client when it submits an item to enable effects from a
// beacon. These items will have been moved into the beacon item slot in advance.
type BeaconPaymentStackRequestAction struct {
	// PrimaryEffect and SecondaryEffect are the effects that were selected from the beacon.
	PrimaryEffect, SecondaryEffect int32
}

// Marshal ...
func (a *BeaconPaymentStackRequestAction) Marshal(w *Writer) {
	w.Varint32(&a.PrimaryEffect)
	w.Varint32(&a.SecondaryEffect)
}

// Unmarshal ...
func (a *BeaconPaymentStackRequestAction) Unmarshal(r *Reader) {
	r.Varint32(&a.PrimaryEffect)
	r.Varint32(&a.SecondaryEffect)
}

// MineBlockStackRequestAction is sent by the client when it breaks a block.
type MineBlockStackRequestAction struct {
	// HotbarSlot is the slot held by the player while mining a block.
	HotbarSlot int32
	// PredictedDurability is the durability of the item that the client assumes to be present at the time.
	PredictedDurability int32
	// StackNetworkID is the unique stack ID that the client assumes to be present at the time. The server
	// must check if these IDs match. If they do not match, servers should reject the stack request that the
	// action holding this info was in.
	StackNetworkID int32
}

// Marshal ...
func (a *MineBlockStackRequestAction) Marshal(w *Writer) {
	w.Varint32(&a.HotbarSlot)
	w.Varint32(&a.PredictedDurability)
	w.Varint32(&a.StackNetworkID)
}

// Unmarshal ...
func (a *MineBlockStackRequestAction) Unmarshal(r *Reader) {
	r.Varint32(&a.HotbarSlot)
	r.Varint32(&a.PredictedDurability)
	r.Varint32(&a.StackNetworkID)
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
func (a *CraftRecipeStackRequestAction) Marshal(w *Writer) {
	w.Varuint32(&a.RecipeNetworkID)
}

// Unmarshal ...
func (a *CraftRecipeStackRequestAction) Unmarshal(r *Reader) {
	r.Varuint32(&a.RecipeNetworkID)
}

// AutoCraftRecipeStackRequestAction is sent by the client similarly to the CraftRecipeStackRequestAction. The
// only difference is that the recipe is automatically created and crafted by shift clicking the recipe book.
type AutoCraftRecipeStackRequestAction struct {
	// RecipeNetworkID is the network ID of the recipe that is about to be crafted. This network ID matches
	// one of the recipes sent in the CraftingData packet, where each of the recipes have a RecipeNetworkID as
	// of 1.16.
	RecipeNetworkID uint32
	// TimesCrafted is how many times the recipe was crafted.
	TimesCrafted byte
}

// Marshal ...
func (a *AutoCraftRecipeStackRequestAction) Marshal(w *Writer) {
	w.Varuint32(&a.RecipeNetworkID)
	w.Uint8(&a.TimesCrafted)
}

// Unmarshal ...
func (a *AutoCraftRecipeStackRequestAction) Unmarshal(r *Reader) {
	r.Varuint32(&a.RecipeNetworkID)
	r.Uint8(&a.TimesCrafted)
}

// CraftCreativeStackRequestAction is sent by the client when it takes an item out fo the creative inventory.
// The item is thus not really crafted, but instantly created.
type CraftCreativeStackRequestAction struct {
	// CreativeItemNetworkID is the network ID of the creative item that is being created. This is one of the
	// creative item network IDs sent in the CreativeContent packet.
	CreativeItemNetworkID uint32
}

// Marshal ...
func (a *CraftCreativeStackRequestAction) Marshal(w *Writer) {
	w.Varuint32(&a.CreativeItemNetworkID)
}

// Unmarshal ...
func (a *CraftCreativeStackRequestAction) Unmarshal(r *Reader) {
	r.Varuint32(&a.CreativeItemNetworkID)
}

// CraftRecipeOptionalStackRequestAction is sent when using an anvil. When this action is sent, the
// FilterStrings field in the respective stack request is non-empty and contains the name of the item created
// using the anvil or cartography table.
type CraftRecipeOptionalStackRequestAction struct {
	// RecipeNetworkID is the network ID of the multi-recipe that is about to be crafted. This network ID matches
	// one of the multi-recipes sent in the CraftingData packet, where each of the recipes have a RecipeNetworkID as
	// of 1.16.
	RecipeNetworkID uint32
	// FilterStringIndex is the index of a filter string sent in a ItemStackRequest.
	FilterStringIndex int32
}

// Marshal ...
func (c *CraftRecipeOptionalStackRequestAction) Marshal(w *Writer) {
	w.Varuint32(&c.RecipeNetworkID)
	w.Int32(&c.FilterStringIndex)
}

// Unmarshal ...
func (c *CraftRecipeOptionalStackRequestAction) Unmarshal(r *Reader) {
	r.Varuint32(&c.RecipeNetworkID)
	r.Int32(&c.FilterStringIndex)
}

// CraftGrindstoneRecipeStackRequestAction is sent when a grindstone recipe is crafted. It contains the RecipeNetworkID
// to identify the recipe crafted, and the cost for crafting the recipe.
type CraftGrindstoneRecipeStackRequestAction struct {
	// RecipeNetworkID is the network ID of the recipe that is about to be crafted. This network ID matches
	// one of the recipes sent in the CraftingData packet, where each of the recipes have a RecipeNetworkID as
	// of 1.16.
	RecipeNetworkID uint32
	// Cost is the cost of the recipe that was crafted.
	Cost int32
}

// Marshal ...
func (c *CraftGrindstoneRecipeStackRequestAction) Marshal(w *Writer) {
	w.Varuint32(&c.RecipeNetworkID)
	w.Varint32(&c.Cost)
}

// Unmarshal ...
func (c *CraftGrindstoneRecipeStackRequestAction) Unmarshal(r *Reader) {
	r.Varuint32(&c.RecipeNetworkID)
	r.Varint32(&c.Cost)
}

// CraftLoomRecipeStackRequestAction is sent when a loom recipe is crafted. It simply contains the
// pattern identifier to figure out what pattern is meant to be applied to the item.
type CraftLoomRecipeStackRequestAction struct {
	// Pattern is the pattern identifier for the loom recipe.
	Pattern string
}

// Marshal ...
func (c *CraftLoomRecipeStackRequestAction) Marshal(w *Writer) {
	w.String(&c.Pattern)
}

// Unmarshal ...
func (c *CraftLoomRecipeStackRequestAction) Unmarshal(r *Reader) {
	r.String(&c.Pattern)
}

// CraftNonImplementedStackRequestAction is an action sent for inventory actions that aren't yet implemented
// in the new system. These include, for example, anvils.
type CraftNonImplementedStackRequestAction struct{}

// Marshal ...
func (*CraftNonImplementedStackRequestAction) Marshal(*Writer) {}

// Unmarshal ...
func (*CraftNonImplementedStackRequestAction) Unmarshal(*Reader) {}

// CraftResultsDeprecatedStackRequestAction is an additional, deprecated packet sent by the client after
// crafting. It holds the final results and the amount of times the recipe was crafted. It shouldn't be used.
// This action is also sent when an item is enchanted. Enchanting should be treated mostly the same way as
// crafting, where the old item is consumed.
type CraftResultsDeprecatedStackRequestAction struct {
	ResultItems  []ItemStack
	TimesCrafted byte
}

// Marshal ...
func (a *CraftResultsDeprecatedStackRequestAction) Marshal(w *Writer) {
	l := uint32(len(a.ResultItems))
	w.Varuint32(&l)
	for _, i := range a.ResultItems {
		w.Item(&i)
	}
	w.Uint8(&a.TimesCrafted)
}

// Unmarshal ...
func (a *CraftResultsDeprecatedStackRequestAction) Unmarshal(r *Reader) {
	var l uint32
	r.Varuint32(&l)
	r.LimitUint32(l, mediumLimit*2)

	a.ResultItems = make([]ItemStack, l)
	for i := uint32(0); i < l; i++ {
		r.Item(&a.ResultItems[i])
	}
	r.Uint8(&a.TimesCrafted)
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

// StackReqSlotInfo reads/writes a StackRequestSlotInfo x using IO r.
func StackReqSlotInfo(r IO, x *StackRequestSlotInfo) {
	r.Uint8(&x.ContainerID)
	r.Uint8(&x.Slot)
	r.Varint32(&x.StackNetworkID)
}
