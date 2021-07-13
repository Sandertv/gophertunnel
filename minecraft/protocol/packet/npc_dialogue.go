package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	NPCDialogueActionOpen int32 = iota
	NPCDialogueActionClose
)

// NPCDialogue is a packet that allows the client to display dialog boxes for interacting with NPCs.
type NPCDialogue struct {
	// ActorUniqueID is the ID of the NPC being requested.
	ActorUniqueID uint64
	// ActionType is the type of action for the packet.
	ActionType int32
	// Dialogue is the text that the client should see.
	Dialogue string
	// SceneName is the scene the data was pulled from for the client.
	SceneName string
	// NPCName is the name of the NPC to be displayed to the client.
	NPCName string
	// ActionJSON is the JSON string of the buttons/actions the server can perform.
	ActionJSON string
}

// ID ...
func (*NPCDialogue) ID() uint32 {
	return IDNPCDialogue
}

// Marshal ...
func (pk *NPCDialogue) Marshal(w *protocol.Writer) {
	w.Uint64(&pk.ActorUniqueID)
	w.Varint32(&pk.ActionType)
	w.String(&pk.Dialogue)
	w.String(&pk.SceneName)
	w.String(&pk.NPCName)
	w.String(&pk.ActionJSON)
}

// Unmarshal ...
func (pk *NPCDialogue) Unmarshal(r *protocol.Reader) {
	r.Uint64(&pk.ActorUniqueID)
	r.Varint32(&pk.ActionType)
	r.String(&pk.Dialogue)
	r.String(&pk.SceneName)
	r.String(&pk.NPCName)
	r.String(&pk.ActionJSON)
}
