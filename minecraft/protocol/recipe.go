package protocol

import (
	"fmt"
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

// PotContainerChangeRecipe reads/writes a PotionContainerChangeRecipe x using IO r.
func PotContainerChangeRecipe(r IO, x *PotionContainerChangeRecipe) {
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

// PotRecipe reads/writes a PotionRecipe x using IO r.
func PotRecipe(r IO, x *PotionRecipe) {
	r.Varint32(&x.InputPotionID)
	r.Varint32(&x.InputPotionMetadata)
	r.Varint32(&x.ReagentItemID)
	r.Varint32(&x.ReagentItemMetadata)
	r.Varint32(&x.OutputPotionID)
	r.Varint32(&x.OutputPotionMetadata)
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
)

// Recipe represents a recipe that may be sent in a CraftingData packet to let the client know what recipes
// are available server-side.
type Recipe interface {
	// Marshal encodes the recipe data to its binary representation into buf.
	Marshal(w *Writer)
	// Unmarshal decodes a serialised recipe from Reader r into the recipe instance.
	Unmarshal(r *Reader)
}

// ShapelessRecipe is a recipe that has no particular shape. Its functionality is shared with the
// RecipeShulkerBox and RecipeShapelessChemistry types.
type ShapelessRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too,
	// but its functionality is not exactly known.
	RecipeID string
	// Input is a list of items that serve as the input of the shapeless recipe. These items are the items
	// required to craft the output.
	Input []RecipeIngredientItem
	// Output is a list of items that are created as a result of crafting the recipe.
	Output []ItemStack
	// UUID is a UUID identifying the recipe. This can actually be set to an empty UUID if the CraftingEvent
	// packet is not used.
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
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
}

// ShulkerBoxRecipe is a shapeless recipe made specifically for shulker box crafting, so that they don't lose
// their user data when dyeing a shulker box.
type ShulkerBoxRecipe ShapelessRecipe

// ShapelessChemistryRecipe is a recipe specifically made for chemistry related features, which exist only in
// the Education Edition. They function the same as shapeless recipes do.
type ShapelessChemistryRecipe ShapelessRecipe

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
	Input []RecipeIngredientItem
	// Output is a list of items that are created as a result of crafting the recipe.
	Output []ItemStack
	// UUID is a UUID identifying the recipe. This can actually be set to an empty UUID if the CraftingEvent
	// packet is not used.
	UUID uuid.UUID
	// Block is the block name that is required to craft the output of the recipe. The block is not prefixed
	// with 'minecraft:', so it will look like 'crafting_table' as an example.
	Block string
	// Priority ...
	Priority int32
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
}

// ShapedChemistryRecipe is a recipe specifically made for chemistry related features, which exist only in the
// Education Edition. It functions the same as a normal ShapedRecipe.
type ShapedChemistryRecipe ShapedRecipe

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
type FurnaceDataRecipe FurnaceRecipe

// MultiRecipe serves as an 'enable' switch for multi-shape recipes.
type MultiRecipe struct {
	// UUID is a UUID identifying the recipe. This can actually be set to an empty UUID if the CraftingEvent
	// packet is not used.
	UUID uuid.UUID
	// RecipeNetworkID is a unique ID used to identify the recipe over network. Each recipe must have a unique
	// network ID. Recommended is to just increment a variable for each unique recipe registered.
	// This field must never be 0.
	RecipeNetworkID uint32
}

// Marshal ...
func (recipe *ShapelessRecipe) Marshal(w *Writer) {
	marshalShapeless(w, recipe)
}

// Unmarshal ...
func (recipe *ShapelessRecipe) Unmarshal(r *Reader) {
	unmarshalShapeless(r, recipe)
}

// Marshal ...
func (recipe *ShulkerBoxRecipe) Marshal(w *Writer) {
	r := ShapelessRecipe(*recipe)
	marshalShapeless(w, &r)
}

// Unmarshal ...
func (recipe *ShulkerBoxRecipe) Unmarshal(r *Reader) {
	shapeless := ShapelessRecipe{}
	unmarshalShapeless(r, &shapeless)
	*recipe = ShulkerBoxRecipe(shapeless)
}

// Marshal ...
func (recipe *ShapelessChemistryRecipe) Marshal(w *Writer) {
	r := ShapelessRecipe(*recipe)
	marshalShapeless(w, &r)
}

// Unmarshal ...
func (recipe *ShapelessChemistryRecipe) Unmarshal(r *Reader) {
	shapeless := ShapelessRecipe{}
	unmarshalShapeless(r, &shapeless)
	*recipe = ShapelessChemistryRecipe(shapeless)
}

// Marshal ...
func (recipe *ShapedRecipe) Marshal(w *Writer) {
	marshalShaped(w, recipe)
}

// Unmarshal ...
func (recipe *ShapedRecipe) Unmarshal(r *Reader) {
	unmarshalShaped(r, recipe)
}

// Marshal ...
func (recipe *ShapedChemistryRecipe) Marshal(w *Writer) {
	r := ShapedRecipe(*recipe)
	marshalShaped(w, &r)
}

// Unmarshal ...
func (recipe *ShapedChemistryRecipe) Unmarshal(r *Reader) {
	shaped := ShapedRecipe{}
	unmarshalShaped(r, &shaped)
	*recipe = ShapedChemistryRecipe(shaped)
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
	r.Item(&recipe.Output)
	r.String(&recipe.Block)

	recipe.InputType.MetadataValue = uint32(dataValue)
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

// marshalShaped ...
func marshalShaped(w *Writer, recipe *ShapedRecipe) {
	w.String(&recipe.RecipeID)
	w.Varint32(&recipe.Width)
	w.Varint32(&recipe.Height)
	itemCount := int(recipe.Width * recipe.Height)
	if len(recipe.Input) != itemCount {
		// We got an input count that was not as as big as the full size of the recipe, so we panic as this is
		// a user error.
		panic(fmt.Sprintf("shaped recipe must have exactly %vx%v input items, but got %v", recipe.Width, recipe.Height, len(recipe.Input)))
	}
	for _, input := range recipe.Input {
		RecipeIngredient(w, &input)
	}
	l := uint32(len(recipe.Output))
	w.Varuint32(&l)
	for _, output := range recipe.Output {
		w.Item(&output)
	}
	w.UUID(&recipe.UUID)
	w.String(&recipe.Block)
	w.Varint32(&recipe.Priority)
	w.Varuint32(&recipe.RecipeNetworkID)
}

// unmarshalShaped ...
func unmarshalShaped(r *Reader, recipe *ShapedRecipe) {
	r.String(&recipe.RecipeID)
	r.Varint32(&recipe.Width)
	r.Varint32(&recipe.Height)
	r.LimitInt32(recipe.Width, 0, lowerLimit)
	r.LimitInt32(recipe.Height, 0, lowerLimit)

	itemCount := int(recipe.Width * recipe.Height)
	recipe.Input = make([]RecipeIngredientItem, itemCount)
	for i := 0; i < itemCount; i++ {
		RecipeIngredient(r, &recipe.Input[i])
	}
	var outputCount uint32
	r.Varuint32(&outputCount)
	r.LimitUint32(outputCount, lowerLimit)

	recipe.Output = make([]ItemStack, outputCount)
	for i := uint32(0); i < outputCount; i++ {
		r.Item(&recipe.Output[i])
	}
	r.UUID(&recipe.UUID)
	r.String(&recipe.Block)
	r.Varint32(&recipe.Priority)
	r.Varuint32(&recipe.RecipeNetworkID)
}

// marshalShapeless ...
func marshalShapeless(w *Writer, recipe *ShapelessRecipe) {
	inputLen, outputLen := uint32(len(recipe.Input)), uint32(len(recipe.Output))
	w.String(&recipe.RecipeID)
	w.Varuint32(&inputLen)
	for _, input := range recipe.Input {
		RecipeIngredient(w, &input)
	}
	w.Varuint32(&outputLen)
	for _, output := range recipe.Output {
		w.Item(&output)
	}
	w.UUID(&recipe.UUID)
	w.String(&recipe.Block)
	w.Varint32(&recipe.Priority)
	w.Varuint32(&recipe.RecipeNetworkID)
}

// unmarshalShapeless ...
func unmarshalShapeless(r *Reader, recipe *ShapelessRecipe) {
	var count uint32
	r.String(&recipe.RecipeID)
	r.Varuint32(&count)
	r.LimitUint32(count, lowerLimit)
	recipe.Input = make([]RecipeIngredientItem, count)
	for i := uint32(0); i < count; i++ {
		RecipeIngredient(r, &recipe.Input[i])
	}
	r.Varuint32(&count)
	r.LimitUint32(count, lowerLimit)
	recipe.Output = make([]ItemStack, count)
	for i := uint32(0); i < count; i++ {
		r.Item(&recipe.Output[i])
	}
	r.UUID(&recipe.UUID)
	r.String(&recipe.Block)
	r.Varint32(&recipe.Priority)
	r.Varuint32(&recipe.RecipeNetworkID)
}
