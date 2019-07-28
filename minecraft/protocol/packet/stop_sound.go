package packet

import (
	"bytes"
	"encoding/binary"
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
}

// ID ...
func (*StopSound) ID() uint32 {
	return IDStopSound
}

// Marshal ...
func (pk *StopSound) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.SoundName)
	_ = binary.Write(buf, binary.LittleEndian, pk.StopAll)
}

// Unmarshal ...
func (pk *StopSound) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.SoundName),
		binary.Read(buf, binary.LittleEndian, &pk.StopAll),
	)
}
