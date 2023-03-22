package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddBehaviourTree is sent by the server to the client. The packet is currently unused by both client and
// server.
type AddBehaviourTree struct {
	// BehaviourTree is an unused string.
	BehaviourTree string
}

// ID ...
func (*AddBehaviourTree) ID() uint32 {
	return IDAddBehaviourTree
}

// Marshal ...
func (pk *AddBehaviourTree) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *AddBehaviourTree) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *AddBehaviourTree) marshal(r protocol.IO) {
	r.String(&pk.BehaviourTree)
}
