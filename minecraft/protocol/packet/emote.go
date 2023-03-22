package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EmoteFlagServerSide = 1 << iota
	EmoteFlagMuteChat
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
func (pk *Emote) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *Emote) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *Emote) marshal(r protocol.IO) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.String(&pk.EmoteID)
	r.Uint8(&pk.Flags)
}
