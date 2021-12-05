package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	LabTableActionCombine = iota
	LabTableActionReact
	LabTableActionReset
)

// LabTable is sent by the client to let the server know it started a chemical reaction in Education Edition,
// and is sent by the server to other clients to show the effects.
// The packet is only functional if Education features are enabled.
type LabTable struct {
	// ActionType is the type of the action that was executed. It is one of the constants above. Typically,
	// only LabTableActionCombine is sent by the client, whereas LabTableActionReact is sent by the server.
	ActionType byte
	// Position is the position at which the lab table used was located.
	Position protocol.BlockPos
	// ReactionType is the type of the reaction that took place as a result of the items put into the lab
	// table. The reaction type can be either that of an item or a particle, depending on whatever the result
	// was of the reaction.
	ReactionType byte
}

// ID ...
func (*LabTable) ID() uint32 {
	return IDLabTable
}

// Marshal ...
func (pk *LabTable) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.ActionType)
	w.BlockPos(&pk.Position)
	w.Uint8(&pk.ReactionType)
}

// Unmarshal ...
func (pk *LabTable) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.ActionType)
	r.BlockPos(&pk.Position)
	r.Uint8(&pk.ReactionType)
}
