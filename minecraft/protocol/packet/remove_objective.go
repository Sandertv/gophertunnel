package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RemoveObjective is sent by the server to remove a scoreboard objective. It is used to stop showing a
// scoreboard to a player.
type RemoveObjective struct {
	// ObjectiveName is the name of the objective that the scoreboard currently active has. This name must
	// be identical to the one sent in the SetDisplayObjective packet.
	ObjectiveName string
}

// ID ...
func (*RemoveObjective) ID() uint32 {
	return IDRemoveObjective
}

// Marshal ...
func (pk *RemoveObjective) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.ObjectiveName)
}

// Unmarshal ...
func (pk *RemoveObjective) Unmarshal(buf *bytes.Buffer) error {
	return protocol.String(buf, &pk.ObjectiveName)
}
