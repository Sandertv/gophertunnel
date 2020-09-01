package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// DebugInfo is a packet sent by the server to the client. It does not seem to do anything when sent to the
// normal client in 1.16.
type DebugInfo struct {
	// PlayerUniqueID is the unique ID of the player that the packet is sent to.
	PlayerUniqueID int64
	// Data is the debug data.
	Data []byte
}

// ID ...
func (*DebugInfo) ID() uint32 {
	return IDDebugInfo
}

// Marshal ...
func (pk *DebugInfo) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.PlayerUniqueID)
	w.ByteSlice(&pk.Data)
}

// Unmarshal ...
func (pk *DebugInfo) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.PlayerUniqueID)
	r.ByteSlice(&pk.Data)
}
