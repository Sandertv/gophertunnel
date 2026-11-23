package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientboundOnlineMusicPlaybackResponse is sent by the server in response to a
// ServerboundRequestOnlineMusicPlayback packet, providing the online music playback information.
type ClientboundOnlineMusicPlaybackResponse struct {
	// RequestID is the ID of the request this packet is a response to.
	RequestID uint32
	// State is the playback state of the music. It is one of the protocol.MusicPlaybackState constants.
	State uint8
	// SharedMetadata is a compact JWT holding signed metadata of the audio content.
	SharedMetadata string
	// PlaybackContent is a compact JWT holding the signed content used to play back the audio.
	PlaybackContent string
}

// ID ...
func (*ClientboundOnlineMusicPlaybackResponse) ID() uint32 {
	return IDClientboundOnlineMusicPlaybackResponse
}

func (pk *ClientboundOnlineMusicPlaybackResponse) Marshal(io protocol.IO) {
	io.Uint32(&pk.RequestID)
	io.Uint8(&pk.State)
	io.String(&pk.SharedMetadata)
	io.String(&pk.PlaybackContent)
}
