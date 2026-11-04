package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerboundRequestOnlineMusicPlayback is sent by the client to request online music playback information
// for a sound event. The server responds with a ClientboundOnlineMusicPlaybackResponse packet.
type ServerboundRequestOnlineMusicPlayback struct {
	// RequestID is the ID of the request, which is sent back in the response.
	RequestID uint32
	// SoundEventName is the name of the sound event to request playback information for.
	SoundEventName string
}

// ID ...
func (*ServerboundRequestOnlineMusicPlayback) ID() uint32 {
	return IDServerboundRequestOnlineMusicPlayback
}

func (pk *ServerboundRequestOnlineMusicPlayback) Marshal(io protocol.IO) {
	io.Uint32(&pk.RequestID)
	io.String(&pk.SoundEventName)
}
