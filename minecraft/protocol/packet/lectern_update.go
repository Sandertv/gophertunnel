package packet

import (
	"bytes"
	"encoding/binary"
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
func (pk *LecternUpdate) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.Page)
	_ = binary.Write(buf, binary.LittleEndian, pk.PageCount)
	_ = protocol.WriteBlockPosition(buf, pk.Position)
	_ = binary.Write(buf, binary.LittleEndian, pk.DropBook)
}

// Unmarshal ...
func (pk *LecternUpdate) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.Page),
		binary.Read(buf, binary.LittleEndian, &pk.PageCount),
		protocol.BlockPosition(buf, &pk.Position),
		binary.Read(buf, binary.LittleEndian, &pk.DropBook),
	)
}
