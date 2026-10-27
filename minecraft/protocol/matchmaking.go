package protocol

// MatchmakingStateOptions holds optional details about what triggered a change in matchmaking state.
type MatchmakingStateOptions struct {
	// TriggeringPlayerName is the name of the player that triggered the change in matchmaking state.
	TriggeringPlayerName Optional[string]
	// TriggeredByLocalPlayer specifies if the change in matchmaking state was triggered by the local player.
	TriggeredByLocalPlayer Optional[bool]
}

// Marshal encodes/decodes a MatchmakingStateOptions.
func (x *MatchmakingStateOptions) Marshal(r IO) {
	OptionalFunc(r, &x.TriggeringPlayerName, r.String)
	OptionalFunc(r, &x.TriggeredByLocalPlayer, r.Bool)
}
