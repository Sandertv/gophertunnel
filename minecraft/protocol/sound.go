package protocol

// The constants below identify the type of update that a SoundDataUpdate holds. The type decides which of
// the update's fields are used.
const (
	SoundDataUpdateStop uint32 = iota
	SoundDataUpdateSetVolume
	SoundDataUpdateSetPitch
	SoundDataUpdateFade
	SoundDataUpdateSeekTo
	SoundDataUpdatePause
	SoundDataUpdateResume
)

// SoundDataUpdate is a single change to a sound that is currently playing, such as a change of its volume or
// a request to pause it.
type SoundDataUpdate struct {
	// Type is the type of the update. It is one of the constants above and decides which of the fields below
	// are written.
	Type uint32
	// Volume is the new volume of the sound. It is only used if Type is SoundDataUpdateSetVolume.
	Volume float32
	// Pitch is the new pitch of the sound. It is only used if Type is SoundDataUpdateSetPitch.
	Pitch float32
	// Duration is the time in seconds that the sound takes to fade to TargetVolume. It is only used if Type
	// is SoundDataUpdateFade.
	Duration float32
	// TargetVolume is the volume that the sound fades to over Duration. It is only used if Type is
	// SoundDataUpdateFade.
	TargetVolume float32
	// Seconds is the time in the sound to seek to. It is only used if Type is SoundDataUpdateSeekTo.
	Seconds float32
}

// Marshal encodes/decodes a SoundDataUpdate.
func (x *SoundDataUpdate) Marshal(r IO) {
	r.Varuint32(&x.Type)
	switch x.Type {
	case SoundDataUpdateStop, SoundDataUpdatePause, SoundDataUpdateResume:
	case SoundDataUpdateSetVolume:
		r.Float32(&x.Volume)
	case SoundDataUpdateSetPitch:
		r.Float32(&x.Pitch)
	case SoundDataUpdateFade:
		r.Float32(&x.Duration)
		r.Float32(&x.TargetVolume)
	case SoundDataUpdateSeekTo:
		r.Float32(&x.Seconds)
	default:
		r.UnknownEnumOption(x.Type, "sound data update type")
	}
}
