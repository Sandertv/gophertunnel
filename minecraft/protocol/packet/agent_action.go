package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	AgentActionTypeAttack = iota + 1
	AgentActionTypeCollect
	AgentActionTypeDestroy
	AgentActionTypeDetectRedstone
	AgentActionTypeDetectObstacle
	AgentActionTypeDrop
	AgentActionTypeDropAll
	AgentActionTypeInspect
	AgentActionTypeInspectData
	AgentActionTypeInspectItemCount
	AgentActionTypeInspectItemDetail
	AgentActionTypeInspectItemSpace
	AgentActionTypeInteract
	AgentActionTypeMove
	AgentActionTypePlaceBlock
	AgentActionTypeTill
	AgentActionTypeTransferItemTo
	AgentActionTypeTurn
)

// AgentAction is an Education Edition packet sent from the server to the client to return a response to a
// previously requested action.
type AgentAction struct {
	// Identifier is a JSON identifier referenced in the initial action.
	Identifier string
	// Action represents the action type that was requested. It is one of the constants defined above.
	Action int32
	// Response is a JSON string containing the response to the action.
	Response []byte
}

// ID ...
func (*AgentAction) ID() uint32 {
	return IDAgentAction
}

// Marshal ...
func (pk *AgentAction) Marshal(w *protocol.Writer) {
	w.String(&pk.Identifier)
	w.Varint32(&pk.Action)
	w.ByteSlice(&pk.Response)
}

// Unmarshal ...
func (pk *AgentAction) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Identifier)
	r.Varint32(&pk.Action)
	r.ByteSlice(&pk.Response)
}
