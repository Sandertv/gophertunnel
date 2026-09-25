package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CraftingData is sent by the server to let the client know all crafting data that the server maintains. This
// includes shapeless crafting, crafting table recipes, furnace recipes etc. Each crafting station's recipes
// are included in it.
type CraftingData struct {
	// ShapedRecipes through SmithingTrimRecipes are the typed recipe vectors used by 1.26.40.
	ShapedRecipes             []protocol.ShapedRecipe
	ShapelessRecipes          []protocol.ShapelessRecipe
	MultiRecipes              []protocol.MultiRecipe
	UserDataShapelessRecipes  []protocol.UserDataShapelessRecipe
	ShapelessChemistryRecipes []protocol.ShapelessChemistryRecipe
	ShapedChemistryRecipes    []protocol.ShapedChemistryRecipe
	SmithingTransformRecipes  []protocol.SmithingTransformRecipe
	SmithingTrimRecipes       []protocol.SmithingTrimRecipe
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

func (pk *CraftingData) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.ShapedRecipes)
	protocol.Slice(io, &pk.ShapelessRecipes)
	protocol.Slice(io, &pk.MultiRecipes)
	protocol.Slice(io, &pk.UserDataShapelessRecipes)
	protocol.Slice(io, &pk.ShapelessChemistryRecipes)
	protocol.Slice(io, &pk.ShapedChemistryRecipes)
	protocol.Slice(io, &pk.SmithingTransformRecipes)
	protocol.Slice(io, &pk.SmithingTrimRecipes)
	protocol.Slice(io, &pk.PotionRecipes)
	protocol.Slice(io, &pk.PotionContainerChangeRecipes)
	protocol.FuncSlice(io, &pk.MaterialReducers, io.MaterialReducer)
	io.Bool(&pk.ClearRecipes)
}
