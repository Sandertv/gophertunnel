package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientboundStonecutterSetRecipe is sent by the server to the client when the selected recipe of a
// stonecutter container is changed server-side.
type ClientboundStonecutterSetRecipe struct {
	// PlayerUniqueID is the unique ID of the player that the stonecutter container belongs to.
	PlayerUniqueID int64
	// ContainerID is the ID of the stonecutter container that the recipe was selected in.
	ContainerID byte
	// RecipeIndex is the index of the recipe that is selected in the stonecutter UI.
	RecipeIndex int32
}

// ID ...
func (*ClientboundStonecutterSetRecipe) ID() uint32 {
	return IDClientboundStonecutterSetRecipe
}

func (pk *ClientboundStonecutterSetRecipe) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.PlayerUniqueID)
	io.Uint8(&pk.ContainerID)
	io.Varint32(&pk.RecipeIndex)
}
