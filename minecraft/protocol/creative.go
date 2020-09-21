package protocol

// CreativeItem represents a creative item present in the creative inventory.
type CreativeItem struct {
	// CreativeItemNetworkID is a unique ID for the creative item. It has to be unique for each creative item
	// sent to the client. An incrementing ID per creative item does the job.
	CreativeItemNetworkID uint32
	// Item is the item that should be added to the creative inventory.
	Item ItemStack
}

// CreativeEntry reads/writes a CreativeItem x using IO r.
func CreativeEntry(r IO, x *CreativeItem) {
	r.Varuint32(&x.CreativeItemNetworkID)
	r.Item(&x.Item)
}
