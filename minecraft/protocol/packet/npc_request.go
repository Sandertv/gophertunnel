package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	// CommandString is used to store a json array of json objects representing buttons.
	NPCRequestActionSetActions = iota
	// ActionType is used to store the index of the button clicked
	// The behaviour of handling this can be different so that we either just handle the action now
	// Or store it and execute it when the ExecuteClosingCommands is called.
	NPCRequestActionExecuteAction
	// ExecuteClosingCommands is used for executing commands (vanilla) or a callback based on when the actual window is closed or when the buttons type is closed.
	NPCRequestActionExecuteClosingCommands
	// SetName sets the name of the given NPC
	// The name is stored in the CommandString and should be used to set the nameTag of the given entity
	NPCRequestActionSetName
	// SetSkin sets the current skin of the NPC
	// This works off of skin index's usually in packs, but as far as I know this doesn't work on servers.
	// The skin index is stored in the CommandString field
	NPCRequestActionSetSkin
	// SetInteractText sets the content (in the nbt data of the entity) to a string
	// This new text is stored in the CommandString field
	NPCRequestActionSetInteractText
	// ExecuteOpeningCommands executes a command thats type is set to open or is forced to by the server!
	NPCRequestActionExecuteOpeningCommands
)

// NPCRequest is sent by the client when it interacts with an NPC.
// The packet is specifically made for Education Edition, where NPCs are available to use.
type NPCRequest struct {
	// EntityRuntimeID is the runtime ID of the NPC entity that the player interacted with. It is the same
	// as sent by the server when spawning the entity.
	EntityRuntimeID uint64
	// RequestType is the type of the request, which depends on the permission that the player has. It will
	// be either a type that indicates that the NPC should show its dialog, or that it should open the
	// editing window.
	RequestType byte
	// CommandString is the command string set in the NPC. It may consist of multiple commands, depending on
	// what the player set in it.
	// Note this depends on the window as SetActions uses this as a json array of buttons.
	CommandString string
	// ActionType is the index of the given button selected
	// Note when this is called we shouldn't run given buttons code as a dialogue isn't submitted until ExecuteClosingCommands is called. 
	ActionType byte
	// SceneName is the name of the scene this is usually just "".
	SceneName string
}

// ID ...
func (*NPCRequest) ID() uint32 {
	return IDNPCRequest
}

// Marshal ...
func (pk *NPCRequest) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Uint8(&pk.RequestType)
	w.String(&pk.CommandString)
	w.Uint8(&pk.ActionType)
	w.String(&pk.SceneName)
}

// Unmarshal ...
func (pk *NPCRequest) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Uint8(&pk.RequestType)
	r.String(&pk.CommandString)
	r.Uint8(&pk.ActionType)
	r.String(&pk.SceneName)
}
