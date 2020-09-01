package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LecternUpdate is sent by the server to update a lectern block, so that players can see that the block
// changed. It is used, for example, when a page is turned.
type LecternUpdate struct {
	// Page is the page number in the book that was opened on the lectern. If no book was opened, the field
	// is ignored.
	Page byte
	// PageCount is the number of pages that the book opened in the lectern has. If can be ignored if no
	// book is on the lectern.
	PageCount byte
	// Position is the position of the lectern that was updated. If no lectern is at the block position,
	// the packet is ignored.
	Position protocol.BlockPos
	// DropBook specifies if the book currently set on display in the lectern should be dropped.
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
