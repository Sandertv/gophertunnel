package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientboundPlayAudioContent is sent by the server to make the client play signed audio content, using the
// sound options authored by the server.
type ClientboundPlayAudioContent struct {
	// SharedMetadata is a compact JWT holding signed metadata of the audio content.
	SharedMetadata string
	// PlaybackContent is a compact JWT holding the signed content used to play back the audio.
	PlaybackContent string
	// PlaybackType is the type of playback. It is one of the protocol.AudioContentPlaybackType constants.
	PlaybackType string
	// Sound holds the options used to play the sound, such as its position, volume and pitch.
	Sound PlaySound
}

// ID ...
func (*ClientboundPlayAudioContent) ID() uint32 {
	return IDClientboundPlayAudioContent
}

func (pk *ClientboundPlayAudioContent) Marshal(io protocol.IO) {
	io.String(&pk.SharedMetadata)
	io.String(&pk.PlaybackContent)
	io.String(&pk.PlaybackType)
	pk.Sound.Marshal(io)
}
