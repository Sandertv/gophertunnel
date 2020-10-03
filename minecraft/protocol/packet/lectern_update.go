package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LecternUpdate is sent by the client to update the server on which page was opened in a book on a lectern,
// or if the book should be removed from it.
type LecternUpdate struct {
	// Page is the page number in the book that was opened by the player on the lectern.
	Page byte
	// PageCount is the number of pages that the book opened in the lectern has.
	PageCount byte
	// Position is the position of the lectern that was updated. If no lectern is at the block position,
	// the packet should be ignored.
	Position protocol.BlockPos
	// DropBook specifies if the book currently set on display in the lectern should be dropped server-side.
	DropBook bool
}

// ID ...
func (*LecternUpdate) ID() uint32 {
	return IDLecternUpdate
}

// Marshal ...
func (pk *LecternUpdate) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.Page)
	w.Uint8(&pk.PageCount)
	w.BlockPos(&pk.Position)
	w.Bool(&pk.DropBook)
}

// Unmarshal ...
func (pk *LecternUpdate) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.Page)
	r.Uint8(&pk.PageCount)
	r.BlockPos(&pk.Position)
	r.Bool(&pk.DropBook)
}
