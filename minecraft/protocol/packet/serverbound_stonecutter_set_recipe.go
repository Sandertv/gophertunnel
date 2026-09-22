package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerboundStonecutterSetRecipe is sent by the client to the server when the player selects a recipe in the
// stonecutter UI.
type ServerboundStonecutterSetRecipe struct {
	// ContainerID is the ID of the stonecutter container that the recipe was selected in.
	ContainerID byte
	// RecipeIndex is the index of the recipe that was selected in the stonecutter UI.
	RecipeIndex int32
}

// ID ...
func (*ServerboundStonecutterSetRecipe) ID() uint32 {
	return IDServerboundStonecutterSetRecipe
}

func (pk *ServerboundStonecutterSetRecipe) Marshal(io protocol.IO) {
	io.Uint8(&pk.ContainerID)
	io.Varint32(&pk.RecipeIndex)
}
