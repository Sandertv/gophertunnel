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
	// Position is the position at which the sound was played. Some of the sounds do not depend on a position,
	// which will then ignore it, but most of them will play with the direction based on the position compared
	// to the player's position.
	Position mgl32.Vec3
	// Volume is the relative volume of the sound to play. It will be less loud for the player if it is
	// farther away from the position of the sound.
	Volume float32
	// Pitch is the pitch of the sound to play. Some sounds completely ignore this field, whereas others use
	// it to specify the pitch as the field is intended.
	Pitch float32
}

// ID ...
func (*PlaySound) ID() uint32 {
	return IDPlaySound
}

// Marshal ...
func (pk *PlaySound) Marshal(w *protocol.Writer) {
	b := protocol.BlockPos{int32(pk.Position[0] * 8), int32(pk.Position[1] * 8), int32(pk.Position[2] * 8)}

	w.String(&pk.SoundName)
	w.BlockPos(&b)
	w.Float32(&pk.Volume)
	w.Float32(&pk.Pitch)
}

// Unmarshal ...
func (pk *PlaySound) Unmarshal(r *protocol.Reader) {
	var b protocol.BlockPos
	r.String(&pk.SoundName)
	r.BlockPos(&b)
	r.Float32(&pk.Volume)
	r.Float32(&pk.Pitch)

	pk.Position = mgl32.Vec3{float32(b[0]) / 8, float32(b[1]) / 8, float32(b[2]) / 8}
}
