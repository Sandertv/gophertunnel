package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerboundMatchmakingCancel is sent by the client to the server to request cancellation of the matchmaking
// ticket that it currently holds.
type ServerboundMatchmakingCancel struct{}

// ID ...
func (*ServerboundMatchmakingCancel) ID() uint32 {
	return IDServerboundMatchmakingCancel
}

func (*ServerboundMatchmakingCancel) Marshal(protocol.IO) {}
