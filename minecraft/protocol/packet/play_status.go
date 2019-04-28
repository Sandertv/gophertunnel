package packet

import (
	"bytes"
	"encoding/binary"
)

const (
	PlayStatusSuccess int32 = iota
	PlayStatusLoginFailedClient
	PlayStatusLoginFailedServer
	PlayStatusPlayerSpawn
	PlayStatusLoginFailedInvalidTenant
	PlayStatusLoginFailedVanillaEdu
	PlayStatusLoginFailedEduVanilla
	PlayStatusLoginFailedServerFull
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
func (pk *PlayStatus) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.BigEndian, pk.Status)
}

// Unmarshal ...
func (pk *PlayStatus) Unmarshal(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.BigEndian, &pk.Status)
}
