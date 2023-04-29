package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// OpenSign is sent by the server to open a sign for editing. As of 1.19.80, the player can interact with a sign to edit
// the text on both sides instead of just the front.
type OpenSign struct {
	// Position is the position of the sign to edit. The client uses this position to get the data of the sign, including
	// the existing text and formatting etc.
	Position protocol.BlockPos
	// FrontSide dictates whether the front side of the sign should be opened for editing. If false, the back side is
	// assumed to be edited.
	FrontSide bool
}

// ID ...
func (*OpenSign) ID() uint32 {
	return IDOpenSign
}

func (pk *OpenSign) Marshal(io protocol.IO) {
	io.UBlockPos(&pk.Position)
	io.Bool(&pk.FrontSide)
}
