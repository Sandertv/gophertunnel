package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UnlockedRecipes gives the client a list of recipes that have been unlocked, restricting the recipes that appear in
// the recipe book.
type UnlockedRecipes struct {
	// NewUnlocks determines if new recipes have been unlocked since the packet was last sent.
	NewUnlocks bool
	// Recipes is a list of recipe names that have been unlocked.
	Recipes []string
}

// ID ...
func (*UnlockedRecipes) ID() uint32 {
	return IDUnlockedRecipes
}

// Marshal ...
func (pk *UnlockedRecipes) Marshal(w *protocol.Writer) {
	w.Bool(&pk.NewUnlocks)
	protocol.FuncSlice(w, &pk.Recipes, w.String)
}

// Unmarshal ...
func (pk *UnlockedRecipes) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.NewUnlocks)
	protocol.FuncSlice(r, &pk.Recipes, r.String)
}
