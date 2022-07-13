package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayStatusLoginSuccess int32 = iota
	PlayStatusLoginFailedClient
	PlayStatusLoginFailedServer
	PlayStatusPlayerSpawn
	PlayStatusLoginFailedInvalidTenant
	PlayStatusLoginFailedVanillaEdu
	PlayStatusLoginFailedEduVanilla
	PlayStatusLoginFailedServerFull
	PlayStatusLoginFailedEditorVanilla
	PlayStatusLoginFailedVanillaEditor
)

// PlayStatus is sent by the server to update a player on the play status. This includes failed statuses due
// to a mismatched version, but also success statuses.
type PlayStatus struct {
	// Status is the status of the packet. It is one of the constants found above.
	Status int32
}

// ID ...
func (*PlayStatus) ID() uint32 {
	return IDPlayStatus
}

// Marshal ...
func (pk *PlayStatus) Marshal(w *protocol.Writer) {
	w.BEInt32(&pk.Status)
}

// Unmarshal ...
func (pk *PlayStatus) Unmarshal(r *protocol.Reader) {
	r.BEInt32(&pk.Status)
}
