package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlaySound is sent by the server to play a sound to the client. Some of the sounds may only be started using
// this packet and must be stopped using the StopSound packet.
type PlaySound struct {
	// SoundName is the name of the sound to play.
	SoundName string
	// Position is the position at which the sound was played. Some sounds do not depend on a position,
	// which will then ignore it, but most of them will play with the direction based on the position compared
	// to the player's position.
	Position mgl32.Vec3
	// Volume is the relative volume of the sound to play. It will be less loud for the player if it is
	// farther away from the position of the sound.
	Volume float32
	// Pitch is the pitch of the sound to play. Some sounds completely ignore this field, whereas others use
	// it to specify the pitch as the field is intended.
	Pitch float32
	// LoopCount is the number of times to loop the sound before stopping. -1 means no looping at all.
	LoopCount int32
	// BypassListenerRangeCheck specifies if the sound should be played regardless of how far away the player
	// is from the position of the sound.
	BypassListenerRangeCheck bool
	// Handle is an optional sound handle ID. It is currently unknown what this is for, and is not required
	// to be set by servers.
	Handle protocol.Optional[uint64]
	// PlaybackPositionSeconds is an optional offset, in seconds, into the sound at which playback should
	// start. If not set, the sound is played from the start.
	PlaybackPositionSeconds protocol.Optional[float32]
}

// ID ...
func (*PlaySound) ID() uint32 {
	return IDPlaySound
}

func (pk *PlaySound) Marshal(io protocol.IO) {
	io.String(&pk.SoundName)
	io.SoundPos(&pk.Position)
	io.Float32(&pk.Volume)
	io.Float32(&pk.Pitch)
	io.Varint32(&pk.LoopCount)
	io.Bool(&pk.BypassListenerRangeCheck)
	protocol.OptionalFunc(io, &pk.Handle, io.Uint64)
	protocol.OptionalFunc(io, &pk.PlaybackPositionSeconds, io.Float32)
}
