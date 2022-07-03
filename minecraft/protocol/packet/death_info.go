package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// DeathInfo is a packet sent from the server to the client expected to be sent when a player dies. It contains messages
// related to the player's death, which are shown on the death screen as of v1.19.10.
type DeathInfo struct {
	// Cause is the cause of the player's death, such as "suffocation" or "suicide".
	Cause string
	// Messages is a list of death messages to be shown on the death screen.
	Messages []string
}

// ID ...
func (*DeathInfo) ID() uint32 {
	return IDDeathInfo
}

// Marshal ...
func (pk *DeathInfo) Marshal(w *protocol.Writer) {
	w.String(&pk.Cause)
	messageLen := uint32(len(pk.Messages))
	w.Varuint32(&messageLen)
	for _, message := range pk.Messages {
		w.String(&message)
	}
}

// Unmarshal ...
func (pk *DeathInfo) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Cause)
	var messageLen uint32
	r.Varuint32(&messageLen)
	pk.Messages = make([]string, messageLen)
	for i := uint32(0); i < messageLen; i++ {
		r.String(&pk.Messages[i])
	}
}
