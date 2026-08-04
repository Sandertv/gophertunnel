package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CraftingData is sent by the server to let the client know all crafting data that the server maintains. This
// includes shapeless crafting, crafting table recipes, furnace recipes etc. Each crafting station's recipes
// are included in it.
type CraftingData struct {
	// Recipes is a list of all recipes available on the server. It includes among others shapeless, shaped
	// and furnace recipes. The client will only be able to craft these recipes.
	Recipes []protocol.Recipe
	// ShapedRecipes through SmithingTrimRecipes are the typed recipe vectors used by 1.26.40.
	ShapedRecipes             []protocol.ShapedRecipe
	ShapelessRecipes          []protocol.ShapelessRecipe
	MultiRecipes              []protocol.MultiRecipe
	ShulkerBoxRecipes         []protocol.ShulkerBoxRecipe
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
	if _, writing := io.(*protocol.Writer); writing {
		pk.distributeRecipes(io)
	}
	recipeSlice(io, &pk.ShapedRecipes)
	recipeSlice(io, &pk.ShapelessRecipes)
	recipeSlice(io, &pk.MultiRecipes)
	recipeSlice(io, &pk.ShulkerBoxRecipes)
	recipeSlice(io, &pk.ShapelessChemistryRecipes)
	recipeSlice(io, &pk.ShapedChemistryRecipes)
	recipeSlice(io, &pk.SmithingTransformRecipes)
	recipeSlice(io, &pk.SmithingTrimRecipes)
	protocol.Slice(io, &pk.PotionRecipes)
	protocol.Slice(io, &pk.PotionContainerChangeRecipes)
	protocol.FuncSlice(io, &pk.MaterialReducers, io.MaterialReducer)
	io.Bool(&pk.ClearRecipes)
	if _, reading := io.(*protocol.Reader); reading {
		pk.collectRecipes()
	}
}

type wireRecipe interface {
	Marshal(*protocol.Writer)
	Unmarshal(*protocol.Reader)
}

func recipeSlice[T any, P interface {
	*T
	wireRecipe
}](io protocol.IO, recipes *[]T) {
	length := uint32(len(*recipes))
	io.Varuint32(&length)
	switch v := io.(type) {
	case *protocol.Writer:
		for i := range *recipes {
			P(&(*recipes)[i]).Marshal(v)
		}
	case *protocol.Reader:
		v.SliceLength(length, 1024)
		*recipes = make([]T, length)
		for i := range *recipes {
			P(&(*recipes)[i]).Unmarshal(v)
		}
	default:
		io.InvalidValue(io, "crafting data IO", "must be a protocol reader or writer")
	}
}

func (pk *CraftingData) distributeRecipes(io protocol.IO) {
	if len(pk.ShapedRecipes)+len(pk.ShapelessRecipes)+len(pk.MultiRecipes)+len(pk.ShulkerBoxRecipes)+
		len(pk.ShapelessChemistryRecipes)+len(pk.ShapedChemistryRecipes)+len(pk.SmithingTransformRecipes)+len(pk.SmithingTrimRecipes) != 0 {
		return
	}
	for _, recipe := range pk.Recipes {
		switch recipe := recipe.(type) {
		case *protocol.ShapedRecipe:
			pk.ShapedRecipes = append(pk.ShapedRecipes, *recipe)
		case *protocol.ShapelessRecipe:
			pk.ShapelessRecipes = append(pk.ShapelessRecipes, *recipe)
		case *protocol.MultiRecipe:
			pk.MultiRecipes = append(pk.MultiRecipes, *recipe)
		case *protocol.ShulkerBoxRecipe:
			pk.ShulkerBoxRecipes = append(pk.ShulkerBoxRecipes, *recipe)
		case *protocol.ShapelessChemistryRecipe:
			pk.ShapelessChemistryRecipes = append(pk.ShapelessChemistryRecipes, *recipe)
		case *protocol.ShapedChemistryRecipe:
			pk.ShapedChemistryRecipes = append(pk.ShapedChemistryRecipes, *recipe)
		case *protocol.SmithingTransformRecipe:
			pk.SmithingTransformRecipes = append(pk.SmithingTransformRecipes, *recipe)
		case *protocol.SmithingTrimRecipe:
			pk.SmithingTrimRecipes = append(pk.SmithingTrimRecipes, *recipe)
		default:
			io.UnknownEnumOption(recipe, "crafting recipe type")
			return
		}
	}
}

func (pk *CraftingData) collectRecipes() {
	pk.Recipes = pk.Recipes[:0]
	for i := range pk.ShapedRecipes {
		pk.Recipes = append(pk.Recipes, &pk.ShapedRecipes[i])
	}
	for i := range pk.ShapelessRecipes {
		pk.Recipes = append(pk.Recipes, &pk.ShapelessRecipes[i])
	}
	for i := range pk.MultiRecipes {
		pk.Recipes = append(pk.Recipes, &pk.MultiRecipes[i])
	}
	for i := range pk.ShulkerBoxRecipes {
		pk.Recipes = append(pk.Recipes, &pk.ShulkerBoxRecipes[i])
	}
	for i := range pk.ShapelessChemistryRecipes {
		pk.Recipes = append(pk.Recipes, &pk.ShapelessChemistryRecipes[i])
	}
	for i := range pk.ShapedChemistryRecipes {
		pk.Recipes = append(pk.Recipes, &pk.ShapedChemistryRecipes[i])
	}
	for i := range pk.SmithingTransformRecipes {
		pk.Recipes = append(pk.Recipes, &pk.SmithingTransformRecipes[i])
	}
	for i := range pk.SmithingTrimRecipes {
		pk.Recipes = append(pk.Recipes, &pk.SmithingTrimRecipes[i])
	}
}
