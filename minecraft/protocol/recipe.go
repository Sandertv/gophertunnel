package protocol

import (
	"github.com/google/uuid"
)

// PotionContainerChangeRecipe represents a recipe to turn a potion from one type to another. This means from
// a drinkable potion + gunpowder -> splash potion, and from a splash potion + dragon breath -> lingering
// potion.
type PotionContainerChangeRecipe struct {
	// InputItemID is the item ID of the item to be put in. This is typically either the ID of a normal potion
	// or a splash potion.
	InputItemID int32
	// ReagentItemID is the item ID of the item that needs to be added to the container in order to create the
	// output item.
	ReagentItemID int32
	// OutputItemID is the item that is created using a combination of the InputItem and ReagentItem, which is
	// typically either the ID of a splash potion or a lingering potion.
	OutputItemID int32
}

// Marshal encodes/decodes a PotionContainerChangeRecipe.
func (x *PotionContainerChangeRecipe) Marshal(r IO) {
	r.Varint32(&x.InputItemID)
	r.Varint32(&x.ReagentItemID)
	r.Varint32(&x.OutputItemID)
}

// PotionRecipe represents a potion mixing recipe which may be used in a brewing stand.
type PotionRecipe struct {
	// InputPotionID is the item ID of the potion to be put in.
	InputPotionID int32
	// InputPotionMetadata is the type of the potion to be put in. This is typically the meta of the
	// awkward potion (or water bottle to create an awkward potion).
	InputPotionMetadata int32
	// ReagentItemID is the item ID of the item that needs to be added to the brewing stand in order to brew
	// the output potion.
	ReagentItemID int32
	// ReagentItemMetadata is the metadata value of the item that needs to be added to the brewing stand in
	// order to brew the output potion.
	ReagentItemMetadata int32
	// OutputPotionID is the item ID of the potion obtained as a result of the brewing recipe.
	OutputPotionID int32
	// OutputPotionMetadata is the type of the potion that is obtained as a result of brewing the input
	// potion with the reagent item.
	OutputPotionMetadata int32
}

// Marshal encodes/decodes a PotionRecipe.
func (x *PotionRecipe) Marshal(r IO) {
	r.Varint32(&x.InputPotionID)
	r.Varint32(&x.InputPotionMetadata)
	r.Varint32(&x.ReagentItemID)
	r.Varint32(&x.ReagentItemMetadata)
	r.Varint32(&x.OutputPotionID)
	r.Varint32(&x.OutputPotionMetadata)
}

const (
	RecipeUnlockContextNone = iota
	RecipeUnlockContextAlwaysUnlocked
	RecipeUnlockContextPlayerInWater
	RecipeUnlockContextPlayerHasManyItems
)

// RecipeUnlockRequirement represents a requirement that must be met in order to unlock a recipe. This is used
// for both shaped and shapeless recipes.
type RecipeUnlockRequirement struct {
	// Context is the context in which the recipe is unlocked. This is one of the constants above.
	Context int32
	// Ingredients are the ingredients required to unlock the recipe and only used if Context is set to none.
	Ingredients []ItemDescriptorCount
}

// Marshal ...
func (x *RecipeUnlockRequirement) Marshal(r IO) {
	r.Varint32(&x.Context)
	present := x.Context == RecipeUnlockContextNone
	r.Bool(&present)
	if present {
		FuncSlice(r, &x.Ingredients, r.ItemDescriptorCount)
	} else {
		x.Ingredients = nil
	}
}

// ShapelessRecipe is a recipe that has no particular shape. Its functionality is shared with the
// RecipeShulkerBox and RecipeShapelessChemistry types.
type ShapelessRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too,
	// but its functionality is not exactly known.
	RecipeID string
	// Input is a list of items that serve as the input of the shapeless recipe. These items are the items
	// required to craft the output.
	Input []ItemDescriptorCount
	// Output is a list of items that are created as a result of crafting the recipe.
	Output []ItemStack
	// UUID is a UUID identifying the recipe. Since the CraftingEvent packet no longer exists, this can always be empty.
	UUID uuid.UUID
	// Block is the block name that is required to craft the output of the recipe. The block is not prefixed
	// with 'minecraft:', so it will look like 'crafting_table' as an example.
	// The available blocks are:
	// - crafting_table
	// - cartography_table
	// - stonecutter
	// - furnace
	// - blast_furnace
	// - smoker
	// - campfire
	Block string
	// Priority ...
	Priority int32
	// UnlockRequirement is a requirement that must be met in order to unlock the recipe.
	UnlockRequirement RecipeUnlockRequirement
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
}

// ShulkerBoxRecipe is a shapeless recipe made specifically for shulker box crafting, so that they don't lose
// their user data when dyeing a shulker box.
type ShulkerBoxRecipe struct {
	ShapelessRecipe
}

// ShapelessChemistryRecipe is a recipe specifically made for chemistry related features, which exist only in
// the Education Edition. They function the same as shapeless recipes do.
type ShapelessChemistryRecipe struct {
	ShapelessRecipe
}

// ShapedRecipe is a recipe that has a specific shape that must be used to craft the output of the recipe.
// Trying to craft the item in any other shape will not work. The ShapedRecipe is of the same structure as the
// ShapedChemistryRecipe.
type ShapedRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too,
	// but its functionality is not exactly known.
	RecipeID string
	// Width is the width of the recipe's shape.
	Width int32
	// Height is the height of the recipe's shape.
	Height int32
	// Input is a list of items that serve as the input of the shapeless recipe. These items are the items
	// required to craft the output. The amount of input items must be exactly equal to Width * Height.
	Input []ItemDescriptorCount
	// Output is a list of items that are created as a result of crafting the recipe.
	Output []ItemStack
	// UUID is a UUID identifying the recipe. Since the CraftingEvent packet no longer exists, this can always be empty.
	UUID uuid.UUID
	// Block is the block name that is required to craft the output of the recipe. The block is not prefixed
	// with 'minecraft:', so it will look like 'crafting_table' as an example.
	Block string
	// Priority ...
	Priority int32
	// AssumeSymmetry specifies if the recipe is symmetrical. If this is set to true, the recipe will be
	// mirrored along the diagonal axis. This means that the recipe will be the same if rotated 180 degrees.
	AssumeSymmetry bool
	// UnlockRequirement is a requirement that must be met in order to unlock the recipe.
	UnlockRequirement RecipeUnlockRequirement
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
}

// ShapedChemistryRecipe is a recipe specifically made for chemistry related features, which exist only in the
// Education Edition. It functions the same as a normal ShapedRecipe.
type ShapedChemistryRecipe struct {
	ShapedRecipe
}

// MultiRecipe serves as an 'enable' switch for multi-shape recipes.
type MultiRecipe struct {
	// UUID is a UUID identifying the recipe. Since the CraftingEvent packet no longer exists, this can always be empty.
	UUID uuid.UUID
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
}

// SmithingTransformRecipe is a recipe specifically used for smithing tables. It has three input items and adds them
// together, resulting in a new item.
type SmithingTransformRecipe struct {
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too,
	// but its functionality is not exactly known.
	RecipeID string
	// Template is the item that is used to shape the Base item based on the Addition being applied.
	Template ItemDescriptorCount
	// Base is the item that the Addition is being applied to in the smithing table.
	Base ItemDescriptorCount
	// Addition is the item that is being added to the Base item to result in a modified item.
	Addition ItemDescriptorCount
	// Result is the resulting item from the two items being added together.
	Result ItemStack
	// Block is the block name that is required to create the output of the recipe. The block is not prefixed with
	// 'minecraft:', so it will look like 'smithing_table' as an example.
	Block string
}

// SmithingTrimRecipe is a recipe specifically used for applying armour trims to an armour piece inside a smithing table.
type SmithingTrimRecipe struct {
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too,
	// but its functionality is not exactly known.
	RecipeID string
	// Template is the item that is used to shape the Base item based on the Addition being applied.
	Template ItemDescriptorCount
	// Base is the item that the Addition is being applied to in the smithing table.
	Base ItemDescriptorCount
	// Addition is the item that is being added to the Base item to result in a modified item.
	Addition ItemDescriptorCount
	// Block is the block name that is required to create the output of the recipe. The block is not prefixed with
	// 'minecraft:', so it will look like 'smithing_table' as an example.
	Block string
}

// Marshal ...
func (recipe *ShapelessRecipe) Marshal(r IO) {
	marshalShapeless(r, recipe, true)
}

// Marshal ...
func (recipe *ShulkerBoxRecipe) Marshal(r IO) {
	marshalShapeless(r, &recipe.ShapelessRecipe, true)
}

// Marshal ...
func (recipe *ShapelessChemistryRecipe) Marshal(r IO) {
	marshalShapeless(r, &recipe.ShapelessRecipe, false)
}

// Marshal ...
func (recipe *ShapedRecipe) Marshal(r IO) {
	marshalShaped(r, recipe, true)
}

// Marshal ...
func (recipe *ShapedChemistryRecipe) Marshal(r IO) {
	marshalShaped(r, &recipe.ShapedRecipe, false)
}

// Marshal ...
func (recipe *MultiRecipe) Marshal(r IO) {
	r.UUID(&recipe.UUID)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// Marshal ...
func (recipe *SmithingTransformRecipe) Marshal(r IO) {
	r.String(&recipe.RecipeID)
	r.ItemDescriptorCount(&recipe.Template)
	r.ItemDescriptorCount(&recipe.Base)
	r.ItemDescriptorCount(&recipe.Addition)
	r.Item(&recipe.Result)
	r.String(&recipe.Block)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// Marshal ...
func (recipe *SmithingTrimRecipe) Marshal(r IO) {
	r.String(&recipe.RecipeID)
	r.ItemDescriptorCount(&recipe.Template)
	r.ItemDescriptorCount(&recipe.Base)
	r.ItemDescriptorCount(&recipe.Addition)
	r.String(&recipe.Block)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// marshalShaped ...
func marshalShaped(r IO, recipe *ShapedRecipe, withRequirement bool) {
	r.String(&recipe.RecipeID)
	r.Varint32(&recipe.Width)
	r.Varint32(&recipe.Height)
	FuncSlice(r, &recipe.Input, r.ItemDescriptorCount)
	if recipe.Width <= 0 || recipe.Height <= 0 {
		r.InvalidValue([2]int32{recipe.Width, recipe.Height}, "shaped recipe dimensions", "must both be positive")
	}
	if int64(len(recipe.Input)) != int64(recipe.Width)*int64(recipe.Height) {
		r.InvalidValue(len(recipe.Input), "shaped recipe ingredients", "must equal width multiplied by height")
	}
	FuncSlice(r, &recipe.Output, r.Item)
	r.UUID(&recipe.UUID)
	r.String(&recipe.Block)
	r.Varint32(&recipe.Priority)
	r.Bool(&recipe.AssumeSymmetry)
	present := withRequirement
	r.Bool(&present)
	if present != withRequirement {
		r.InvalidValue(present, "shaped recipe unlock requirement presence", "does not match recipe type")
	}
	if present {
		Single(r, &recipe.UnlockRequirement)
	}
	r.Varuint32(&recipe.RecipeNetworkID)
}

// marshalShapeless ...
func marshalShapeless(r IO, recipe *ShapelessRecipe, withRequirement bool) {
	r.String(&recipe.RecipeID)
	FuncSlice(r, &recipe.Input, r.ItemDescriptorCount)
	FuncSlice(r, &recipe.Output, r.Item)
	r.UUID(&recipe.UUID)
	r.String(&recipe.Block)
	r.Varint32(&recipe.Priority)
	present := withRequirement
	r.Bool(&present)
	if present != withRequirement {
		r.InvalidValue(present, "shapeless recipe unlock requirement presence", "does not match recipe type")
	}
	if present {
		Single(r, &recipe.UnlockRequirement)
	}
	r.Varuint32(&recipe.RecipeNetworkID)
}
