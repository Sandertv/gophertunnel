package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundCloseForm is sent by the server to clear the entire form stack of the client. This means that all
// forms that are currently open will be closed. This does not affect inventories and other containers.
type ClientBoundCloseForm struct{}

// ID ...
func (*ClientBoundCloseForm) ID() uint32 {
	return IDClientBoundCloseForm
}

func (pk *ClientBoundCloseForm) Marshal(protocol.IO) {}
