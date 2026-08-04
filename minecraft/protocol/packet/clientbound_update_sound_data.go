package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ClientboundUpdateSoundData is sent by the server to update a sound that is currently playing, identified by
// the handle that the server sent in the PlaySound packet that started it. Every field below holds one update
// to apply to that sound, and the server may leave any of them out.
type ClientboundUpdateSoundData struct {
	// ServerSoundHandle is the server-side handle of the sound to update. It is the handle that was sent in
	// the PlaySound packet that started the sound.
	ServerSoundHandle uint64
	// Stop stops the sound.
	Stop protocol.Optional[protocol.SoundDataUpdate]
	// SetVolume changes the volume of the sound.
	SetVolume protocol.Optional[protocol.SoundDataUpdate]
	// SetPitch changes the pitch of the sound.
	SetPitch protocol.Optional[protocol.SoundDataUpdate]
	// Fade fades the volume of the sound to a target volume over a duration.
	Fade protocol.Optional[protocol.SoundDataUpdate]
	// SeekTo seeks to a specific time in the sound.
	SeekTo protocol.Optional[protocol.SoundDataUpdate]
	// Pause pauses the sound, so that it may be resumed later.
	Pause protocol.Optional[protocol.SoundDataUpdate]
	// Resume resumes a sound that was paused before.
	Resume protocol.Optional[protocol.SoundDataUpdate]
}

// ID ...
func (*ClientboundUpdateSoundData) ID() uint32 {
	return IDClientboundUpdateSoundData
}

func (pk *ClientboundUpdateSoundData) Marshal(io protocol.IO) {
	io.Uint64(&pk.ServerSoundHandle)
	protocol.OptionalMarshaler(io, &pk.Stop)
	protocol.OptionalMarshaler(io, &pk.SetVolume)
	protocol.OptionalMarshaler(io, &pk.SetPitch)
	protocol.OptionalMarshaler(io, &pk.Fade)
	protocol.OptionalMarshaler(io, &pk.SeekTo)
	protocol.OptionalMarshaler(io, &pk.Pause)
	protocol.OptionalMarshaler(io, &pk.Resume)
}
