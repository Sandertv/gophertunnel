package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CursorItemDragStart = iota
	CursorItemDragStop
)

// ServerboundCursorItemDrag is sent by the client to signal the start and end of a cursor item drag
// operation, in which the item held by the cursor is split over multiple slots.
type ServerboundCursorItemDrag struct {
	// State is the state of the drag. It is either CursorItemDragStart or CursorItemDragStop.
	State uint8
}

// ID ...
func (*ServerboundCursorItemDrag) ID() uint32 {
	return IDServerboundCursorItemDrag
}

func (pk *ServerboundCursorItemDrag) Marshal(io protocol.IO) {
	io.Uint8(&pk.State)
}
