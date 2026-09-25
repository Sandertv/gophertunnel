package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ClientboundUpdateSoundData is sent by the server to update a sound that is currently playing, identified by
// the handle that the server sent in the PlaySound packet that started it. Currently, the data field is
// repeated 7 times, but only the last Resume field is actually used by the client for the actual value.
type ClientboundUpdateSoundData struct {
	// ServerSoundHandle is the server-side handle of the sound to update.
	ServerSoundHandle uint64

	Stop              protocol.SoundDataUpdate
	SetVolume         protocol.SoundDataUpdate
	SetPitch          protocol.SoundDataUpdate
	Fade              protocol.SoundDataUpdate
	SeekTo            protocol.SoundDataUpdate
	Pause             protocol.SoundDataUpdate
	Resume            protocol.SoundDataUpdate
}

// ID ...
func (*ClientboundUpdateSoundData) ID() uint32 {
	return IDClientboundUpdateSoundData
}

func (pk *ClientboundUpdateSoundData) Marshal(io protocol.IO) {
	io.Uint64(&pk.ServerSoundHandle)
	protocol.Single(io, &pk.Stop)
	protocol.Single(io, &pk.SetVolume)
	protocol.Single(io, &pk.SetPitch)
	protocol.Single(io, &pk.Fade)
	protocol.Single(io, &pk.SeekTo)
	protocol.Single(io, &pk.Pause)
	protocol.Single(io, &pk.Resume)
}
