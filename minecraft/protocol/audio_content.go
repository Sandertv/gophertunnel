package protocol

const (
	AudioContentPlaybackTypeMusic = "Music"
	AudioContentPlaybackTypeSound = "Sound"
)

// AudioContentRegistrationEntry is an entry sent by the client in the ServerboundRegisterAudioContent packet
// to register service signed audio content for the session.
type AudioContentRegistrationEntry struct {
	// AudioContentID is the ID of the audio content that is registered.
	AudioContentID string
	// SharedMetadata is a compact JWT holding signed metadata shared between the client and server.
	SharedMetadata string
	// ServerContent is a compact JWT holding signed content intended for the server.
	ServerContent string
	// PlaybackContent is a compact JWT holding signed content used to play back the audio.
	PlaybackContent string
}

// Marshal encodes/decodes an AudioContentRegistrationEntry.
func (x *AudioContentRegistrationEntry) Marshal(r IO) {
	r.String(&x.AudioContentID)
	r.String(&x.SharedMetadata)
	r.String(&x.ServerContent)
	r.String(&x.PlaybackContent)
}
