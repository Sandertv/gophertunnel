package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RecordStarted is sent by the server to notify the client that a record started playing at a specific block
// position, such as when a music disc is inserted into a jukebox.
type RecordStarted struct {
	// Position is the position of the block that the record started playing at.
	Position protocol.BlockPos
	// Handle is the server-side handle of the sound that the record is played through. It may be used in the
	// ClientboundUpdateSoundData packet to update the sound afterwards.
	Handle uint64
}

// ID ...
func (*RecordStarted) ID() uint32 {
	return IDRecordStarted
}

func (pk *RecordStarted) Marshal(io protocol.IO) {
	io.BlockPos(&pk.Position)
	io.Uint64(&pk.Handle)
}
