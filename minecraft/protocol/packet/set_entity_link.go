package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetEntityLink is sent by the server to initiate an entity link client-side, meaning one entity will start
// riding another.
type SetEntityLink struct {
	// EntityLink is the link to be set client-side. It links two entities together, so that one entity rides
	// another. Note that players that see those entities later will not see the link, unless it is also sent
	// in the AddEntity and AddPlayer packets.
	EntityLink protocol.EntityLink
}

// ID ...
func (*SetEntityLink) ID() uint32 {
	return IDSetEntityLink
}

// Marshal ...
func (pk *SetEntityLink) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteEntityLinkAction(buf, pk.EntityLink)
}

// Unmarshal ...
func (pk *SetEntityLink) Unmarshal(buf *bytes.Buffer) error {
	return protocol.EntityLinkAction(buf, &pk.EntityLink)
}
