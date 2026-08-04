package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ClientboundUpdateSoundData is sent by the server to update a sound that is currently playing, identified by
// the handle that the server sent in the PlaySound packet that started it. Each optional field is a Cereal union
// slot that may hold any SoundDataUpdate variant; its name does not constrain the variant on the wire.
type ClientboundUpdateSoundData struct {
	// ServerSoundHandle is the server-side handle of the sound to update.
	ServerSoundHandle uint64
	// Stop is the optional sound update slot named Stop.
	Stop protocol.Optional[protocol.SoundDataUpdate]
	// SetVolume is the optional sound update slot named SetVolume.
	SetVolume protocol.Optional[protocol.SoundDataUpdate]
	// SetPitch is the optional sound update slot named SetPitch.
	SetPitch protocol.Optional[protocol.SoundDataUpdate]
	// Fade is the optional sound update slot named Fade.
	Fade protocol.Optional[protocol.SoundDataUpdate]
	// SeekTo is the optional sound update slot named SeekTo.
	SeekTo protocol.Optional[protocol.SoundDataUpdate]
	// Pause is the optional sound update slot named Pause.
	Pause protocol.Optional[protocol.SoundDataUpdate]
	// Resume is the optional sound update slot named Resume.
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
