package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GameRulesChanged is sent by the server to the client to update client-side game rules, such as game rules
// like the 'showCoordinates' game rule.
type GameRulesChanged struct {
	// GameRules defines game rules changed with their respective values. The value of these game rules may be
	// either 'bool', 'int32' or 'float32'.
	// Note that some game rules are server side only, and don't necessarily need to be sent to the client.
	// Only changed game rules need to be sent in this packet. Game rules that were not changed do not need to
	// be sent if the client is already updated on them.
	GameRules map[string]interface{}
}

// ID ...
func (*GameRulesChanged) ID() uint32 {
	return IDGameRulesChanged
}

// Marshal ...
func (pk *GameRulesChanged) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteGameRules(buf, pk.GameRules)
}

// Unmarshal ...
func (pk *GameRulesChanged) Unmarshal(r *protocol.Reader) {
	pk.GameRules = make(map[string]interface{})
	protocol.GameRules(r, &pk.GameRules)
}
