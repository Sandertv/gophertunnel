package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AutomationClientConnect is used to make the client connect to a websocket server. This websocket server has
// the ability to execute commands on the behalf of the client and it can listen for certain events fired by
// the client.
type AutomationClientConnect struct {
	// ServerURI is the URI to make the client connect to. It can be, for example, 'localhost:8000/ws' to
	// connect to a websocket server on the localhost at port 8000.
	ServerURI string
}

// ID ...
func (*AutomationClientConnect) ID() uint32 {
	return IDAutomationClientConnect
}

// Marshal ...
func (pk *AutomationClientConnect) Marshal(w *protocol.Writer) {
	w.String(&pk.ServerURI)
}

// Unmarshal ...
func (pk *AutomationClientConnect) Unmarshal(r *protocol.Reader) {
	r.String(&pk.ServerURI)
}
