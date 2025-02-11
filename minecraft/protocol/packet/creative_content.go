package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CreativeContent is a packet sent by the server to set the creative inventory's content for a player.
// Introduced in 1.16, this packet replaces the previous method - sending an InventoryContent packet with
// creative inventory window ID.
// As of v1.21.60, this packet is no longer required to be sent as part of the login sequence however the
// client will crash if they try to open their creative inventory before receiving this packet. Every item
// must be part of a group, any items that are not part of a group will need to reference an "anonymous group"
// which has an empty name OR no icon. The order of Groups and Items is how the client will render items
// in the creative inventory compared to the previous, hard coded order.
//
// Below is an example of defining 2 ungrouped items, 2 grouped items and then another 2 ungrouped items, all
// in the nature category.
//
//	CreativeContent{
//		Groups: []protocol.CreativeGroup{
//			{Category: 1}, // No name or icon, this is the "anonymous group"
//			{Category: 1, Name: "itemGroup.name.planks", Icon: protocol.ItemStack{...}}, // A "planks" group
//			{Category: 1}, // Another "anonymous group"
//		},
//		Items: []protocol.CreativeItem{
//			{CreativeItemNetworkID: 0, Item: protocol.ItemStack{...}, GroupIndex: 0}, // Ungrouped before "planks"
//			{CreativeItemNetworkID: 1, Item: protocol.ItemStack{...}, GroupIndex: 0}, // Ungrouped before "planks"
//			{CreativeItemNetworkID: 2, Item: protocol.ItemStack{...}, GroupIndex: 1}, // Grouped under the "planks" group
//			{CreativeItemNetworkID: 3, Item: protocol.ItemStack{...}, GroupIndex: 1}, // Grouped under the "planks" group
//			{CreativeItemNetworkID: 4, Item: protocol.ItemStack{...}, GroupIndex: 2}, // Ungrouped after "planks"
//			{CreativeItemNetworkID: 5, Item: protocol.ItemStack{...}, GroupIndex: 2}, // Ungrouped after "planks"
//		}
//	}
type CreativeContent struct {
	// Groups is a list of the groups that should be added to the creative inventory.
	Groups []protocol.CreativeGroup
	// Items is a list of the items that should be added to the creative inventory.
	Items []protocol.CreativeItem
}

// ID ...
func (*CreativeContent) ID() uint32 {
	return IDCreativeContent
}

func (pk *CreativeContent) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Groups)
	protocol.Slice(io, &pk.Items)
}
