package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerboundRegisterAudioContent is sent by the client to register service signed audio content for the
// authenticated player session.
type ServerboundRegisterAudioContent struct {
	// Registrations is a list of audio content registered by the client.
	Registrations []protocol.AudioContentRegistrationEntry
}

// ID ...
func (*ServerboundRegisterAudioContent) ID() uint32 {
	return IDServerboundRegisterAudioContent
}

func (pk *ServerboundRegisterAudioContent) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Registrations)
}
