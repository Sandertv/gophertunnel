package protocol

import (
	"bytes"
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

// WritePotContainerChangeRecipe writes a PotionContainerChangeRecipe x to Buffer dst.
func WritePotContainerChangeRecipe(dst *bytes.Buffer, x PotionContainerChangeRecipe) error {
	return chainErr(
		WriteVarint32(dst, x.InputItemID),
		WriteVarint32(dst, x.ReagentItemID),
		WriteVarint32(dst, x.OutputItemID),
	)
}

// PotContainerChangeRecipe reads a PotionContainerChangeRecipe x from Buffer src.
func PotContainerChangeRecipe(src *bytes.Buffer, x *PotionContainerChangeRecipe) error {
	return chainErr(
		Varint32(src, &x.InputItemID),
		Varint32(src, &x.ReagentItemID),
		Varint32(src, &x.OutputItemID),
	)
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

// WritePotRecipe writes a PotionRecipe x to Buffer dst.
func WritePotRecipe(dst *bytes.Buffer, x PotionRecipe) error {
	return chainErr(
		WriteVarint32(dst, x.InputPotionID),
		WriteVarint32(dst, x.InputPotionMetadata),
		WriteVarint32(dst, x.ReagentItemID),
		WriteVarint32(dst, x.ReagentItemMetadata),
		WriteVarint32(dst, x.OutputPotionID),
		WriteVarint32(dst, x.OutputPotionMetadata),
	)
}

// PotRecipe reads a PotionRecipe x from Buffer src.
func PotRecipe(src *bytes.Buffer, x *PotionRecipe) error {
	return chainErr(
		Varint32(src, &x.InputPotionID),
		Varint32(src, &x.InputPotionMetadata),
		Varint32(src, &x.ReagentItemID),
		Varint32(src, &x.ReagentItemMetadata),
		Varint32(src, &x.OutputPotionID),
		Varint32(src, &x.OutputPotionMetadata),
	)
}

const (
	RecipeShapeless = iota
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
	Marshal(buf *bytes.Buffer)
	// Unmarshal decodes a serialised recipe in buf into the recipe instance.
	Unmarshal(buf *bytes.Buffer) error
}

// ShapelessRecipe is a recipe that has no particular shape. Its functionality is shared with the
// RecipeShulkerBox and RecipeShapelessChemistry types.
type ShapelessRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too,
	// but its functionality is not exactly known.
	RecipeID string
	// Input is a list of items that serve as the input of the shapeless recipe. These items are the items
	// required to craft the output.
	Input []ItemStack
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
	Input []ItemStack
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
func (recipe *ShapelessRecipe) Marshal(buf *bytes.Buffer) {
	marshalShapeless(buf, recipe)
}

// Unmarshal ...
func (recipe *ShapelessRecipe) Unmarshal(buf *bytes.Buffer) error {
	return unmarshalShapeless(buf, recipe)
}

// Marshal ...
func (recipe *ShulkerBoxRecipe) Marshal(buf *bytes.Buffer) {
	r := ShapelessRecipe(*recipe)
	marshalShapeless(buf, &r)
}

// Unmarshal ...
func (recipe *ShulkerBoxRecipe) Unmarshal(buf *bytes.Buffer) error {
	r := ShapelessRecipe{}
	if err := unmarshalShapeless(buf, &r); err != nil {
		return err
	}
	*recipe = ShulkerBoxRecipe(r)
	return nil
}

// Marshal ...
func (recipe *ShapelessChemistryRecipe) Marshal(buf *bytes.Buffer) {
	r := ShapelessRecipe(*recipe)
	marshalShapeless(buf, &r)
}

// Unmarshal ...
func (recipe *ShapelessChemistryRecipe) Unmarshal(buf *bytes.Buffer) error {
	r := ShapelessRecipe{}
	if err := unmarshalShapeless(buf, &r); err != nil {
		return err
	}
	*recipe = ShapelessChemistryRecipe(r)
	return nil
}

// Marshal ...
func (recipe *ShapedRecipe) Marshal(buf *bytes.Buffer) {
	marshalShaped(buf, recipe)
}

// Unmarshal ...
func (recipe *ShapedRecipe) Unmarshal(buf *bytes.Buffer) error {
	return unmarshalShaped(buf, recipe)
}

// Marshal ...
func (recipe *ShapedChemistryRecipe) Marshal(buf *bytes.Buffer) {
	r := ShapedRecipe(*recipe)
	marshalShaped(buf, &r)
}

// Unmarshal ...
func (recipe *ShapedChemistryRecipe) Unmarshal(buf *bytes.Buffer) error {
	r := ShapedRecipe{}
	if err := unmarshalShaped(buf, &r); err != nil {
		return err
	}
	*recipe = ShapedChemistryRecipe(r)
	return nil
}

// Marshal ...
func (recipe *FurnaceRecipe) Marshal(buf *bytes.Buffer) {
	_ = WriteVarint32(buf, recipe.InputType.NetworkID)
	_ = WriteItem(buf, recipe.Output)
	_ = WriteString(buf, recipe.Block)
}

// Unmarshal ...
func (recipe *FurnaceRecipe) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		Varint32(buf, &recipe.InputType.NetworkID),
		Item(buf, &recipe.Output),
		String(buf, &recipe.Block),
	)
}

// Marshal ...
func (recipe *FurnaceDataRecipe) Marshal(buf *bytes.Buffer) {
	_ = WriteVarint32(buf, recipe.InputType.NetworkID)
	_ = WriteVarint32(buf, int32(recipe.InputType.MetadataValue))
	_ = WriteItem(buf, recipe.Output)
	_ = WriteString(buf, recipe.Block)
}

// Unmarshal ...
func (recipe *FurnaceDataRecipe) Unmarshal(buf *bytes.Buffer) error {
	var dataValue int32
	if err := chainErr(
		Varint32(buf, &recipe.InputType.NetworkID),
		Varint32(buf, &dataValue),
		Item(buf, &recipe.Output),
		String(buf, &recipe.Block),
	); err != nil {
		return err
	}
	recipe.InputType.MetadataValue = int16(dataValue)
	return nil
}

// Marshal ...
func (recipe *MultiRecipe) Marshal(buf *bytes.Buffer) {
	_ = WriteUUID(buf, recipe.UUID)
	_ = WriteVaruint32(buf, recipe.RecipeNetworkID)
}

// Unmarshal ...
func (recipe *MultiRecipe) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		UUID(buf, &recipe.UUID),
		Varuint32(buf, &recipe.RecipeNetworkID),
	)
}

// marshalShaped ...
func marshalShaped(buf *bytes.Buffer, recipe *ShapedRecipe) {
	_ = WriteString(buf, recipe.RecipeID)
	_ = WriteVarint32(buf, recipe.Width)
	_ = WriteVarint32(buf, recipe.Height)
	itemCount := int(recipe.Width * recipe.Height)
	if len(recipe.Input) != itemCount {
		// We got an input count that was not as as big as the full size of the recipe, so we panic as this is
		// a user error.
		panic(fmt.Sprintf("shaped recipe must have exactly %vx%v input items, but got %v", recipe.Width, recipe.Height, len(recipe.Input)))
	}
	for _, input := range recipe.Input {
		_ = WriteRecipeIngredient(buf, input)
	}
	_ = WriteVaruint32(buf, uint32(len(recipe.Output)))
	for _, output := range recipe.Output {
		_ = WriteItem(buf, output)
	}
	_ = WriteUUID(buf, recipe.UUID)
	_ = WriteString(buf, recipe.Block)
	_ = WriteVarint32(buf, recipe.Priority)
	_ = WriteVaruint32(buf, recipe.RecipeNetworkID)
}

// unmarshalShaped ...
func unmarshalShaped(buf *bytes.Buffer, recipe *ShapedRecipe) error {
	if err := chainErr(
		String(buf, &recipe.RecipeID),
		Varint32(buf, &recipe.Width),
		Varint32(buf, &recipe.Height),
	); err != nil {
		return err
	}
	if recipe.Width <= 0 || recipe.Height <= 0 {
		// Make sure we don't have a width/height smaller than or equal to 0, as it means we get an invalid
		// item count.
		return fmt.Errorf("recipe width and height must be bigger than 0, but got %v by %v", recipe.Width, recipe.Height)
	}
	if recipe.Width > lowerLimit || recipe.Height > lowerLimit {
		return LimitHitError{Type: "shaped recipe dimensions", Limit: lowerLimit}
	}
	itemCount := int(recipe.Width * recipe.Height)
	recipe.Input = make([]ItemStack, itemCount)
	for i := 0; i < itemCount; i++ {
		if err := RecipeIngredient(buf, &recipe.Input[i]); err != nil {
			return err
		}
	}
	var outputCount uint32
	if err := Varuint32(buf, &outputCount); err != nil {
		return err
	}
	if outputCount > lowerLimit {
		return LimitHitError{Type: "shaped recipe output", Limit: lowerLimit}
	}
	recipe.Output = make([]ItemStack, outputCount)
	for i := uint32(0); i < outputCount; i++ {
		if err := Item(buf, &recipe.Output[i]); err != nil {
			return err
		}
	}
	return chainErr(
		UUID(buf, &recipe.UUID),
		String(buf, &recipe.Block),
		Varint32(buf, &recipe.Priority),
		Varuint32(buf, &recipe.RecipeNetworkID),
	)
}

// marshalShapeless ...
func marshalShapeless(buf *bytes.Buffer, recipe *ShapelessRecipe) {
	_ = WriteString(buf, recipe.RecipeID)
	_ = WriteVaruint32(buf, uint32(len(recipe.Input)))
	for _, input := range recipe.Input {
		_ = WriteRecipeIngredient(buf, input)
	}
	_ = WriteVaruint32(buf, uint32(len(recipe.Output)))
	for _, output := range recipe.Output {
		_ = WriteItem(buf, output)
	}
	_ = WriteUUID(buf, recipe.UUID)
	_ = WriteString(buf, recipe.Block)
	_ = WriteVarint32(buf, recipe.Priority)
	_ = WriteVaruint32(buf, recipe.RecipeNetworkID)
}

// unmarshalShapeless ...
func unmarshalShapeless(buf *bytes.Buffer, recipe *ShapelessRecipe) error {
	var count uint32
	if err := chainErr(
		String(buf, &recipe.RecipeID),
		Varuint32(buf, &count),
	); err != nil {
		return err
	}
	if count > lowerLimit {
		return LimitHitError{Type: "shapeless recipe input", Limit: lowerLimit}
	}
	recipe.Input = make([]ItemStack, count)
	for i := uint32(0); i < count; i++ {
		if err := RecipeIngredient(buf, &recipe.Input[i]); err != nil {
			return wrap(err)
		}
	}
	if err := Varuint32(buf, &count); err != nil {
		return wrap(err)
	}
	if count > lowerLimit {
		return LimitHitError{Type: "shapeless recipe output", Limit: lowerLimit}
	}
	recipe.Output = make([]ItemStack, count)
	for i := uint32(0); i < count; i++ {
		if err := Item(buf, &recipe.Output[i]); err != nil {
			return wrap(err)
		}
	}
	return chainErr(
		UUID(buf, &recipe.UUID),
		String(buf, &recipe.Block),
		Varint32(buf, &recipe.Priority),
		Varuint32(buf, &recipe.RecipeNetworkID),
	)
}
