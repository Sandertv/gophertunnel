package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerFog is sent by the server to render the different fogs in the Stack. The types of fog are controlled
// by resource packs to change how they are rendered, and the ability to create custom fog.
type PlayerFog struct {
	// Stack is a list of fog identifiers to be sent to the client. Examples of fog identifiers are
	// "minecraft:fog_ocean" and "minecraft:fog_hell".
	Stack []string
}

// ID ...
func (*PlayerFog) ID() uint32 {
	return IDPlayerFog
}

func (pk *PlayerFog) Marshal(io protocol.IO) {
	protocol.FuncSlice(io, &pk.Stack, io.String)
}
