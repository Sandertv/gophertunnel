package protocol

const (
	SoundDataUpdateStop = iota
	SoundDataUpdateSetVolume
	SoundDataUpdateSetPitch
	SoundDataUpdateFade
	SoundDataUpdateSeekTo
	SoundDataUpdatePause
	SoundDataUpdateResume
)

// SoundDataUpdate is a single change to a sound that is currently playing.
type SoundDataUpdate struct {
	// Type is the type of the update. It is one of the SoundDataUpdate constants above.
	Type uint8
	// Volume is used if Type is SoundDataUpdateSetVolume.
	Volume float32
	// Pitch is used if Type is SoundDataUpdateSetPitch.
	Pitch float32
	// Duration is the fade duration in seconds. It is used if Type is SoundDataUpdateFade.
	Duration float32
	// TargetVolume is the fade target volume. It is used if Type is SoundDataUpdateFade.
	TargetVolume float32
	// Seconds is the position to seek to. It is used if Type is SoundDataUpdateSeekTo.
	Seconds float32
}

// Marshal encodes/decodes a SoundDataUpdate.
func (x *SoundDataUpdate) Marshal(io IO) {
	io.Uint8(&x.Type)
	switch x.Type {
	case SoundDataUpdateStop, SoundDataUpdatePause, SoundDataUpdateResume:
	case SoundDataUpdateSetVolume:
		io.Float32(&x.Volume)
	case SoundDataUpdateSetPitch:
		io.Float32(&x.Pitch)
	case SoundDataUpdateFade:
		io.Float32(&x.Duration)
		io.Float32(&x.TargetVolume)
	case SoundDataUpdateSeekTo:
		io.Float32(&x.Seconds)
	default:
		io.UnknownEnumOption(x.Type, "sound data update type")
	}
}

// SoundAttachOptions holds the options to attach a sound to an entity.
type SoundAttachOptions struct {
	// EntityRuntimeID is the runtime ID of the entity to attach the sound to. The position
	// of the sound follows the position of this entity.
	EntityRuntimeID uint64
}

// Marshal encodes/decodes a SoundAttachOptions.
func (x *SoundAttachOptions) Marshal(io IO) {
	io.ActorRuntimeID(&x.EntityRuntimeID)
}
