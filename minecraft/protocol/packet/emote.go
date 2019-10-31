package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EmoteFlagServerSide = 1 << iota
)

// Emote is sent by both the server and the client. When the client sends an emote, it sends this packet to
// the server, after which the server will broadcast the packet to other players online.
type Emote struct {
	// EntityRuntimeID is the entity that sent the emote. When a player sends this packet, it has this field
	// set as its own entity runtime ID.
	EntityRuntimeID uint64
	// EmoteID is the ID of the emote to send.
	EmoteID string
	// Flags is a combination of flags that change the way the Emote packet operates. When the server sends
	// this packet to other players, EmoteFlagServerSide must be present.
	Flags byte
}

// ID ...
func (*Emote) ID() uint32 {
	return IDEmote
}

// Marshal ...
func (pk *Emote) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteString(buf, pk.EmoteID)
	_ = binary.Write(buf, binary.LittleEndian, pk.Flags)
}

// Unmarshal ...
func (pk *Emote) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.String(buf, &pk.EmoteID),
		binary.Read(buf, binary.LittleEndian, &pk.Flags),
	)
}
