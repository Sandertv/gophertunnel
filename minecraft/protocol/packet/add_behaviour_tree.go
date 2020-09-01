package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddBehaviourTree is sent by the server to the client. Its usage remains unknown, as behaviour packs are
// typically all sent at the start of the game.
type AddBehaviourTree struct {
	// BehaviourTree is a JSON encoded tree containing behaviour. It does not seem like it has any real
	// effect on the client.
	BehaviourTree string
}

// ID ...
func (*AddBehaviourTree) ID() uint32 {
	return IDAddBehaviourTree
}

// Marshal ...
func (pk *AddBehaviourTree) Marshal(w *protocol.Writer) {
	w.String(&pk.BehaviourTree)
}

// Unmarshal ...
func (pk *AddBehaviourTree) Unmarshal(r *protocol.Reader) {
	r.String(&pk.BehaviourTree)
}
