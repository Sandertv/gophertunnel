package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	MatchmakingStateIdle = iota
	MatchmakingStateMatchmaking
	MatchmakingStateMatchFound
	MatchmakingStateCanceled
	MatchmakingStatePlayerLeftParty
	MatchmakingStatePlayerLeftServer
	MatchmakingStateServerShutdown
	MatchmakingStateTimedOut
	MatchmakingStateRequeueAsParty
)

// ClientboundMatchmakingState is sent by the server to inform the client of the state of the matchmaking
// process that it is part of.
type ClientboundMatchmakingState struct {
	// State is the state that the matchmaking process is currently in. It is one of the MatchmakingState
	// constants above.
	State uint8
	// DestinationName is the name of the destination that the client is being matchmade into.
	DestinationName string
	// Options holds optional details about what triggered the change in matchmaking state.
	Options protocol.Optional[protocol.MatchmakingStateOptions]
}

// ID ...
func (*ClientboundMatchmakingState) ID() uint32 {
	return IDClientboundMatchmakingState
}

func (pk *ClientboundMatchmakingState) Marshal(io protocol.IO) {
	io.Uint8(&pk.State)
	io.String(&pk.DestinationName)
	protocol.OptionalMarshaler(io, &pk.Options)
}
