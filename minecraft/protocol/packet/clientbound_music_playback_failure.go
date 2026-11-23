package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientboundMusicPlaybackFailure is sent by the server to report music discs that could not be played.
type ClientboundMusicPlaybackFailure struct {
	// Failures holds an entry for every disc that could not be played.
	Failures []protocol.MusicPlaybackFailureEntry
}

// ID ...
func (*ClientboundMusicPlaybackFailure) ID() uint32 {
	return IDClientboundMusicPlaybackFailure
}

func (pk *ClientboundMusicPlaybackFailure) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Failures)
}
