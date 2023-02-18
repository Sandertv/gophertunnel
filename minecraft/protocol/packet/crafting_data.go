package packet

import (
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
	// MaterialReducers is a list of all material reducers which is used in education edition chemistry.
	MaterialReducers []protocol.MaterialReducer
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
func (pk *CraftingData) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Recipes))
	w.Varuint32(&l)
	for _, recipe := range pk.Recipes {
		var c int32
		switch recipe.(type) {
		case *protocol.ShapelessRecipe:
			c = protocol.RecipeShapeless
		case *protocol.ShapedRecipe:
			c = protocol.RecipeShaped
		case *protocol.FurnaceRecipe:
			c = protocol.RecipeFurnace
		case *protocol.FurnaceDataRecipe:
			c = protocol.RecipeFurnaceData
		case *protocol.MultiRecipe:
			c = protocol.RecipeMulti
		case *protocol.ShulkerBoxRecipe:
			c = protocol.RecipeShulkerBox
		case *protocol.ShapelessChemistryRecipe:
			c = protocol.RecipeShapelessChemistry
		case *protocol.ShapedChemistryRecipe:
			c = protocol.RecipeShapedChemistry
		case *protocol.SmithingTransformRecipe:
			c = protocol.RecipeSmithingTransform
		default:
			w.UnknownEnumOption(fmt.Sprintf("%T", recipe), "crafting recipe type")
		}
		w.Varint32(&c)
		recipe.Marshal(w)
	}
	protocol.Slice(w, &pk.PotionRecipes)
	protocol.Slice(w, &pk.PotionContainerChangeRecipes)
	protocol.FuncSlice(w, &pk.MaterialReducers, w.MaterialReducer)
	w.Bool(&pk.ClearRecipes)
}

// Unmarshal ...
func (pk *CraftingData) Unmarshal(r *protocol.Reader) {
	var length uint32
	r.Varuint32(&length)
	pk.Recipes = make([]protocol.Recipe, length)
	for i := uint32(0); i < length; i++ {
		var recipeType int32
		r.Varint32(&recipeType)

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
		case protocol.RecipeSmithingTransform:
			recipe = &protocol.SmithingTransformRecipe{}
		default:
			r.UnknownEnumOption(recipeType, "crafting data recipe type")
			return
		}
		recipe.Unmarshal(r)
		pk.Recipes[i] = recipe
	}
	protocol.Slice(r, &pk.PotionRecipes)
	protocol.Slice(r, &pk.PotionContainerChangeRecipes)
	protocol.FuncSlice(r, &pk.MaterialReducers, r.MaterialReducer)
	r.Bool(&pk.ClearRecipes)
}
