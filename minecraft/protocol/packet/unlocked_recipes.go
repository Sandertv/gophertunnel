package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	UnlockedRecipesTypeEmpty = iota
	UnlockedRecipesTypeInitiallyUnlocked
	UnlockedRecipesTypeNewlyUnlocked
	UnlockedRecipesTypeRemoveUnlocked
	UnlockedRecipesTypeRemoveAllUnlocked
)

// UnlockedRecipes gives the client a list of recipes that have been unlocked, restricting the recipes that appear in
// the recipe book.
type UnlockedRecipes struct {
	// UnlockType is the type of unlock that the packet represents, and can either be adding or removing a list of recipes.
	// It is one of the constants listed above.
	UnlockType uint32
	// Recipes is a list of recipe names that have been unlocked.
	Recipes []string
}

// ID ...
func (*UnlockedRecipes) ID() uint32 {
	return IDUnlockedRecipes
}

func (pk *UnlockedRecipes) Marshal(io protocol.IO) {
	io.Uint32(&pk.UnlockType)
	protocol.FuncSlice(io, &pk.Recipes, io.String)
}
