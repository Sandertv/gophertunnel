package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// CreatePhoto is a packet that allows players to export photos from their portfolios into items in their inventory.
// This packet only works on the Education Edition version of Minecraft.
type CreatePhoto struct {
	// EntityUniqueID is the unique ID of the entity.
	EntityUniqueID int64
	// PhotoName is the name of the photo.
	PhotoName string
	// ItemName is the name of the photo as an item.
	ItemName string
}

// ID ...
func (*CreatePhoto) ID() uint32 {
	return IDCreatePhoto
}

// Marshal ...
func (pk *CreatePhoto) Marshal(w *protocol.Writer) {
	w.Int64(&pk.EntityUniqueID)
	w.String(&pk.PhotoName)
	w.String(&pk.ItemName)
}

// Unmarshal ...
func (pk *CreatePhoto) Unmarshal(r *protocol.Reader) {
	r.Int64(&pk.EntityUniqueID)
	r.String(&pk.PhotoName)
	r.String(&pk.ItemName)
}
