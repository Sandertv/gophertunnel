package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// MapCreateLockedCopy is sent by the server to create a locked copy of one map into another map. In vanilla,
// it is used in the cartography table to create a map that is locked and cannot be modified.
type MapCreateLockedCopy struct {
	// OriginalMapID is the ID of the map that is being copied. The locked copy will obtain all content that
	// is visible on this map, except the content will not change.
	OriginalMapID int64
	// NewMapID is the ID of the map that holds the locked copy of the map that OriginalMapID points to. Its
	// contents will be impossible to change.
	NewMapID int64
}

// ID ...
func (*MapCreateLockedCopy) ID() uint32 {
	return IDMapCreateLockedCopy
}

// Marshal ...
func (pk *MapCreateLockedCopy) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.OriginalMapID)
	w.Varint64(&pk.NewMapID)
}

// Unmarshal ...
func (pk *MapCreateLockedCopy) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.OriginalMapID)
	r.Varint64(&pk.NewMapID)
}
