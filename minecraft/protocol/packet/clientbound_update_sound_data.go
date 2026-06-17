package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	SoundDataEventStop = "Stop"
)

// ClientboundUpdateSoundData is sent by the server to update the state of a server-controlled sound.
type ClientboundUpdateSoundData struct {
	// ServerSoundHandle is the server-side handle identifying the sound to update.
	ServerSoundHandle uint64
	// SoundEvent is the action to apply to the sound. It is one of the SoundDataEvent constants.
	SoundEvent string
}

// ID ...
func (*ClientboundUpdateSoundData) ID() uint32 {
	return IDClientboundUpdateSoundData
}

func (pk *ClientboundUpdateSoundData) Marshal(io protocol.IO) {
	io.Uint64(&pk.ServerSoundHandle)
	io.String(&pk.SoundEvent)
}
