package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// SoundFadeData holds a target volume and fade duration, in seconds.
type SoundFadeData struct {
	TargetVolume float32
	Duration     float32
}

// ClientboundUpdateSoundData is sent by the server to update the state of a server-controlled sound.
type ClientboundUpdateSoundData struct {
	// ServerSoundHandle is the server-side handle identifying the sound to update.
	ServerSoundHandle uint64
	// Stop, Volume, Pitch, Fade, SeekTo, Pause and Resume are independently optional updates.
	Stop   protocol.Optional[struct{}]
	Volume protocol.Optional[float32]
	Pitch  protocol.Optional[float32]
	Fade   protocol.Optional[SoundFadeData]
	SeekTo protocol.Optional[float32]
	Pause  protocol.Optional[struct{}]
	Resume protocol.Optional[struct{}]
}

// ID ...
func (*ClientboundUpdateSoundData) ID() uint32 {
	return IDClientboundUpdateSoundData
}

func (pk *ClientboundUpdateSoundData) Marshal(io protocol.IO) {
	io.Uint64(&pk.ServerSoundHandle)
	marshalSoundData(io, &pk.Stop, func(*struct{}) {})
	marshalSoundData(io, &pk.Volume, io.Float32)
	marshalSoundData(io, &pk.Pitch, io.Float32)
	marshalSoundData(io, &pk.Fade, func(fade *SoundFadeData) {
		io.Float32(&fade.TargetVolume)
		io.Float32(&fade.Duration)
	})
	marshalSoundData(io, &pk.SeekTo, io.Float32)
	marshalSoundData(io, &pk.Pause, func(*struct{}) {})
	marshalSoundData(io, &pk.Resume, func(*struct{}) {})
}

func marshalSoundData[T any](io protocol.IO, value *protocol.Optional[T], payload func(*T)) {
	protocol.OptionalFunc(io, value, func(value *T) {
		var variant uint32
		io.Varuint32(&variant)
		if variant != 0 {
			io.UnknownEnumOption(variant, "sound data variant")
		}
		payload(value)
	})
}
