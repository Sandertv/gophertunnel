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
	Context byte
	// Ingredients are the ingredients required to unlock the recipe and only used if Context is set to none.
	Ingredients []ItemDescriptorCount
}

// Marshal ...
func (x *RecipeUnlockRequirement) Marshal(r IO) {
	r.Uint8(&x.Context)
	if x.Context == RecipeUnlockContextNone {
		FuncSlice(r, &x.Ingredients, r.ItemDescriptorCount)
	}
}

const (
	RecipeShapeless int32 = iota
	RecipeShaped
	RecipeFurnace
	RecipeFurnaceData
	RecipeMulti
	RecipeShulkerBox
	RecipeShapelessChemistry
	RecipeShapedChemistry
	RecipeSmithingTransform
	RecipeSmithingTrim
)

// Recipe represents a recipe that may be sent in a CraftingData packet to let the client know what recipes
// are available server-side.
type Recipe interface {
	// Marshal encodes the recipe data to its binary representation into buf.
	Marshal(w *Writer)
	// Unmarshal decodes a serialised recipe from Reader r into the recipe instance.
	Unmarshal(r *Reader)
}

// lookupRecipe looks up the Recipe for a recipe type. False is returned if not
// found.
func lookupRecipe(recipeType int32, x *Recipe) bool {
	switch recipeType {
	case RecipeShapeless:
		*x = &ShapelessRecipe{}
	case RecipeShaped:
		*x = &ShapedRecipe{}
	case RecipeFurnace:
		*x = &FurnaceRecipe{}
	case RecipeFurnaceData:
		*x = &FurnaceDataRecipe{}
	case RecipeMulti:
		*x = &MultiRecipe{}
	case RecipeShulkerBox:
		*x = &ShulkerBoxRecipe{}
	case RecipeShapelessChemistry:
		*x = &ShapelessChemistryRecipe{}
	case RecipeShapedChemistry:
		*x = &ShapedChemistryRecipe{}
	case RecipeSmithingTransform:
		*x = &SmithingTransformRecipe{}
	case RecipeSmithingTrim:
		*x = &SmithingTrimRecipe{}
	default:
		return false
	}
	return true
}

// lookupRecipeType looks up the recipe type for a Recipe. False is returned if
// none was found.
func lookupRecipeType(x Recipe, recipeType *int32) bool {
	switch x.(type) {
	case *ShapelessRecipe:
		*recipeType = RecipeShapeless
	case *ShapedRecipe:
		*recipeType = RecipeShaped
	case *FurnaceRecipe:
		*recipeType = RecipeFurnace
	case *FurnaceDataRecipe:
		*recipeType = RecipeFurnaceData
	case *MultiRecipe:
		*recipeType = RecipeMulti
	case *ShulkerBoxRecipe:
		*recipeType = RecipeShulkerBox
	case *ShapelessChemistryRecipe:
		*recipeType = RecipeShapelessChemistry
	case *ShapedChemistryRecipe:
		*recipeType = RecipeShapedChemistry
	case *SmithingTransformRecipe:
		*recipeType = RecipeSmithingTransform
	case *SmithingTrimRecipe:
		*recipeType = RecipeSmithingTrim
	default:
		return false
	}
	return true
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

// FurnaceRecipe is a recipe that is specifically used for all kinds of furnaces. These recipes don't just
// apply to furnaces, but also blast furnaces and smokers.
type FurnaceRecipe struct {
	// InputType is the item type of the input item. The metadata value of the item is not used in the
	// FurnaceRecipe. Use FurnaceDataRecipe to allow an item with only one metadata value.
	InputType ItemType
	// Output is the item that is created as a result of smelting/cooking an item in the furnace.
	Output ItemStack
	// Block is the block name that is required to create the output of the recipe. The block is not prefixed
	// with 'minecraft:', so it will look like 'furnace' as an example.
	Block string
}

// FurnaceDataRecipe is a recipe specifically used for furnace-type crafting stations. It is equal to
// FurnaceRecipe, except it has an input item with a specific metadata value, instead of any metadata value.
type FurnaceDataRecipe struct {
	FurnaceRecipe
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
func (recipe *ShapelessRecipe) Marshal(w *Writer) {
	marshalShapeless(w, recipe)
}

// Unmarshal ...
func (recipe *ShapelessRecipe) Unmarshal(r *Reader) {
	marshalShapeless(r, recipe)
}

// Marshal ...
func (recipe *ShulkerBoxRecipe) Marshal(w *Writer) {
	marshalShapeless(w, &recipe.ShapelessRecipe)
}

// Unmarshal ...
func (recipe *ShulkerBoxRecipe) Unmarshal(r *Reader) {
	marshalShapeless(r, &recipe.ShapelessRecipe)
}

// Marshal ...
func (recipe *ShapelessChemistryRecipe) Marshal(w *Writer) {
	marshalShapeless(w, &recipe.ShapelessRecipe)
}

// Unmarshal ...
func (recipe *ShapelessChemistryRecipe) Unmarshal(r *Reader) {
	marshalShapeless(r, &recipe.ShapelessRecipe)
}

// Marshal ...
func (recipe *ShapedRecipe) Marshal(w *Writer) {
	marshalShaped(w, recipe)
}

// Unmarshal ...
func (recipe *ShapedRecipe) Unmarshal(r *Reader) {
	marshalShaped(r, recipe)
}

// Marshal ...
func (recipe *ShapedChemistryRecipe) Marshal(w *Writer) {
	marshalShaped(w, &recipe.ShapedRecipe)
}

// Unmarshal ...
func (recipe *ShapedChemistryRecipe) Unmarshal(r *Reader) {
	marshalShaped(r, &recipe.ShapedRecipe)
}

// Marshal ...
func (recipe *FurnaceRecipe) Marshal(w *Writer) {
	w.Varint32(&recipe.InputType.NetworkID)
	w.Item(&recipe.Output)
	w.String(&recipe.Block)
}

// Unmarshal ...
func (recipe *FurnaceRecipe) Unmarshal(r *Reader) {
	r.Varint32(&recipe.InputType.NetworkID)
	r.Item(&recipe.Output)
	r.String(&recipe.Block)
}

// Marshal ...
func (recipe *FurnaceDataRecipe) Marshal(w *Writer) {
	w.Varint32(&recipe.InputType.NetworkID)
	aux := int32(recipe.InputType.MetadataValue)
	w.Varint32(&aux)
	w.Item(&recipe.Output)
	w.String(&recipe.Block)
}

// Unmarshal ...
func (recipe *FurnaceDataRecipe) Unmarshal(r *Reader) {
	var dataValue int32
	r.Varint32(&recipe.InputType.NetworkID)
	r.Varint32(&dataValue)
	recipe.InputType.MetadataValue = uint32(dataValue)
	r.Item(&recipe.Output)
	r.String(&recipe.Block)
}

// Marshal ...
func (recipe *MultiRecipe) Marshal(w *Writer) {
	w.UUID(&recipe.UUID)
	w.Varuint32(&recipe.RecipeNetworkID)
}

// Unmarshal ...
func (recipe *MultiRecipe) Unmarshal(r *Reader) {
	r.UUID(&recipe.UUID)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// Marshal ...
func (recipe *SmithingTransformRecipe) Marshal(w *Writer) {
	w.String(&recipe.RecipeID)
	w.ItemDescriptorCount(&recipe.Template)
	w.ItemDescriptorCount(&recipe.Base)
	w.ItemDescriptorCount(&recipe.Addition)
	w.Item(&recipe.Result)
	w.String(&recipe.Block)
	w.Varuint32(&recipe.RecipeNetworkID)
}

// Unmarshal ...
func (recipe *SmithingTransformRecipe) Unmarshal(r *Reader) {
	r.String(&recipe.RecipeID)
	r.ItemDescriptorCount(&recipe.Template)
	r.ItemDescriptorCount(&recipe.Base)
	r.ItemDescriptorCount(&recipe.Addition)
	r.Item(&recipe.Result)
	r.String(&recipe.Block)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// Marshal ...
func (recipe *SmithingTrimRecipe) Marshal(w *Writer) {
	w.String(&recipe.RecipeID)
	w.ItemDescriptorCount(&recipe.Template)
	w.ItemDescriptorCount(&recipe.Base)
	w.ItemDescriptorCount(&recipe.Addition)
	w.String(&recipe.Block)
	w.Varuint32(&recipe.RecipeNetworkID)
}

// Unmarshal ...
func (recipe *SmithingTrimRecipe) Unmarshal(r *Reader) {
	r.String(&recipe.RecipeID)
	r.ItemDescriptorCount(&recipe.Template)
	r.ItemDescriptorCount(&recipe.Base)
	r.ItemDescriptorCount(&recipe.Addition)
	r.String(&recipe.Block)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// marshalShaped ...
func marshalShaped(r IO, recipe *ShapedRecipe) {
	r.String(&recipe.RecipeID)
	r.Varint32(&recipe.Width)
	r.Varint32(&recipe.Height)
	FuncSliceOfLen(r, uint32(recipe.Width*recipe.Height), &recipe.Input, r.ItemDescriptorCount)
	FuncSlice(r, &recipe.Output, r.Item)
	r.UUID(&recipe.UUID)
	r.String(&recipe.Block)
	r.Varint32(&recipe.Priority)
	r.Bool(&recipe.AssumeSymmetry)
	Single(r, &recipe.UnlockRequirement)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// marshalShapeless ...
func marshalShapeless(r IO, recipe *ShapelessRecipe) {
	r.String(&recipe.RecipeID)
	FuncSlice(r, &recipe.Input, r.ItemDescriptorCount)
	FuncSlice(r, &recipe.Output, r.Item)
	r.UUID(&recipe.UUID)
	r.String(&recipe.Block)
	r.Varint32(&recipe.Priority)
	Single(r, &recipe.UnlockRequirement)
	r.Varuint32(&recipe.RecipeNetworkID)
}
