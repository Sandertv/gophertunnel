package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ShowCredits is sent by the server to show the Minecraft credits screen to the client. It is typically sent
// when the player beats the ender dragon and leaves the End.
type ShowCredits struct {
	// PlayerRuntimeID is the entity runtime ID of the player to show the credits to. It's not clear why this
	// field is actually here in the first place.
	PlayerRuntimeID uint64
	// StatusType is the status type of the credits. It is one of the constants above, and either starts or
	// stops the credits.
	StatusType int32
}

// ID ...
func (*ShowCredits) ID() uint32 {
	return IDShowCredits
}

// Marshal ...
func (pk *ShowCredits) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.PlayerRuntimeID)
	_ = protocol.WriteVarint32(buf, pk.StatusType)
}

// Unmarshal ...
func (pk *ShowCredits) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint64(buf, &pk.PlayerRuntimeID),
		protocol.Varint32(buf, &pk.StatusType),
	)
}
