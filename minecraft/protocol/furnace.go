package protocol

const (
	FurnaceLayoutNone = iota
	FurnaceLayoutInventoryOnly
	FurnaceLayoutDefault
)

const (
	FurnaceLeftTabNone = iota
	FurnaceLeftTabRecipeFood
	FurnaceLeftTabRecipeItems
	FurnaceLeftTabRecipeBlocks
	FurnaceLeftTabRecipeSearch
	FurnaceLeftTabInventory
)

// FurnaceOptions holds the options that a player has selected in a furnace-like container's UI.
type FurnaceOptions struct {
	// LeftFurnaceTab is the tab that is selected on the left side of the furnace UI. It is one of the
	// FurnaceLeftTab constants above.
	LeftFurnaceTab int32
	// Filtering is whether the player has enabled the filtering between recipes they have unlocked or not.
	Filtering bool
	// Layout is the layout of the furnace UI. It is one of the FurnaceLayout constants above.
	Layout int32
}

// Marshal encodes/decodes a FurnaceOptions.
func (x *FurnaceOptions) Marshal(r IO) {
	r.Varint32(&x.LeftFurnaceTab)
	r.Bool(&x.Filtering)
	r.Varint32(&x.Layout)
}
