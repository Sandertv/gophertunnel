package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CraftingData is sent by the server to let the client know all crafting data that the server maintains. This
// includes shapeless crafting, crafting table recipes, furnace recipes etc. Each crafting station's recipes
// are included in it.
type CraftingData struct {
	// Recipes is a list of all recipes available on the server. It includes among others shapeless, shaped
	// and furnace recipes. The client will only be able to craft these recipes.
	Recipes []protocol.Recipe
	// PotionRecipes is a list of all potion mixing recipes which may be used in the brewing stand.
	PotionRecipes []protocol.PotionRecipe
	// PotionContainerChangeRecipes is a list of all recipes to convert a potion from one type to another,
	// such as from a drinkable potion to a splash potion, or from a splash potion to a lingering potion.
	PotionContainerChangeRecipes []protocol.PotionContainerChangeRecipe
	// ClearRecipes indicates if all recipes currently active on the client should be cleaned. Doing this
	// means that the client will have no recipes active by itself: Any CraftingData packets previously sent
	// will also be discarded, and only the recipes in this CraftingData packet will be used.
	ClearRecipes bool
}

// ID ...
func (*CraftingData) ID() uint32 {
	return IDCraftingData
}

// Marshal ...
func (pk *CraftingData) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Recipes)))
	for _, recipe := range pk.Recipes {
		switch recipe.(type) {
		case *protocol.ShapelessRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeShapeless)
		case *protocol.ShapedRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeShaped)
		case *protocol.FurnaceRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeFurnace)
		case *protocol.FurnaceDataRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeFurnaceData)
		case *protocol.MultiRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeMulti)
		case *protocol.ShulkerBoxRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeShulkerBox)
		case *protocol.ShapelessChemistryRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeShapelessChemistry)
		case *protocol.ShapedChemistryRecipe:
			_ = protocol.WriteVarint32(buf, protocol.RecipeShapedChemistry)
		default:
			panic(fmt.Sprintf("invalid crafting data recipe type %T", recipe))
		}
		recipe.Marshal(buf)
	}
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.PotionRecipes)))
	for _, mix := range pk.PotionRecipes {
		_ = protocol.WritePotRecipe(buf, mix)
	}
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.PotionContainerChangeRecipes)))
	for _, mix := range pk.PotionContainerChangeRecipes {
		_ = protocol.WritePotContainerChangeRecipe(buf, mix)
	}
	_ = binary.Write(buf, binary.LittleEndian, pk.ClearRecipes)
}

// Unmarshal ...
func (pk *CraftingData) Unmarshal(buf *bytes.Buffer) error {
	var length uint32
	if err := protocol.Varuint32(buf, &length); err != nil {
		return err
	}
	pk.Recipes = make([]protocol.Recipe, length)
	for i := uint32(0); i < length; i++ {
		var recipeType int32
		if err := protocol.Varint32(buf, &recipeType); err != nil {
			return err
		}
		var recipe protocol.Recipe
		switch recipeType {
		case protocol.RecipeShapeless:
			recipe = &protocol.ShapelessRecipe{}
		case protocol.RecipeShaped:
			recipe = &protocol.ShapedRecipe{}
		case protocol.RecipeFurnace:
			recipe = &protocol.FurnaceRecipe{}
		case protocol.RecipeFurnaceData:
			recipe = &protocol.FurnaceDataRecipe{}
		case protocol.RecipeMulti:
			recipe = &protocol.MultiRecipe{}
		case protocol.RecipeShulkerBox:
			recipe = &protocol.ShulkerBoxRecipe{}
		case protocol.RecipeShapelessChemistry:
			recipe = &protocol.ShapelessChemistryRecipe{}
		case protocol.RecipeShapedChemistry:
			recipe = &protocol.ShapedChemistryRecipe{}
		default:
			return fmt.Errorf("unknown crafting data recipe type %v", recipeType)
		}
		if err := recipe.Unmarshal(buf); err != nil {
			return err
		}
		pk.Recipes[i] = recipe
	}
	if err := protocol.Varuint32(buf, &length); err != nil {
		return err
	}
	pk.PotionRecipes = make([]protocol.PotionRecipe, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.PotRecipe(buf, &pk.PotionRecipes[i]); err != nil {
			return err
		}
	}
	if err := protocol.Varuint32(buf, &length); err != nil {
		return err
	}
	pk.PotionContainerChangeRecipes = make([]protocol.PotionContainerChangeRecipe, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.PotContainerChangeRecipe(buf, &pk.PotionContainerChangeRecipes[i]); err != nil {
			return err
		}
	}
	return binary.Read(buf, binary.LittleEndian, &pk.ClearRecipes)
}
