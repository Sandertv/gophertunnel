package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	LoadingScreenTypeUnknown = iota
	LoadingScreenTypeStart
	LoadingScreenTypeEnd
)

// ServerBoundLoadingScreen is sent by the client to tell the server about the state of the loading
// screen that the client is currently displaying.
type ServerBoundLoadingScreen struct {
	// Type is the type of the loading screen event. It is one of the constants that may be found above.
	Type int32
	// LoadingScreenID is the ID of the screen that was previously sent by the server in the ChangeDimension
	// packet. The server should validate that the ID matches the last one it sent.
	LoadingScreenID protocol.Optional[uint32]
}

// ID ...
func (*ServerBoundLoadingScreen) ID() uint32 {
	return IDServerBoundLoadingScreen
}

func (pk *ServerBoundLoadingScreen) Marshal(io protocol.IO) {
	io.Varint32(&pk.Type)
	protocol.OptionalFunc(io, &pk.LoadingScreenID, io.Uint32)
}
