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

// Marshal ...
func (pk *PlayerFog) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Stack))
	w.Varuint32(&l)
	for i := range pk.Stack {
		w.String(&pk.Stack[i])
	}
}

// Unmarshal ...
func (pk *PlayerFog) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	pk.Stack = make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&pk.Stack[i])
	}
}
