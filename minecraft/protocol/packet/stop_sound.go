package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// StopSound is sent by the server to stop a sound playing to the player, such as a playing music disk track
// or other long-lasting sounds.
type StopSound struct {
	// SoundName is the name of the sound that should be stopped from playing. If no sound with this name is
	// currently active, the packet is ignored.
	SoundName string
	// StopAll specifies if all sounds currently playing to the player should be stopped. If set to true, the
	// SoundName field may be left empty.
	StopAll bool
	// StopMusicLegacy is currently unknown.
	StopMusicLegacy bool
}

// ID ...
func (*StopSound) ID() uint32 {
	return IDStopSound
}

func (pk *StopSound) Marshal(io protocol.IO) {
	io.String(&pk.SoundName)
	io.Bool(&pk.StopAll)
	io.Bool(&pk.StopMusicLegacy)
}
