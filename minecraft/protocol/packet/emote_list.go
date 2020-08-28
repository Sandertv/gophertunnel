package packet

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// EmoteList is sent by the client every time it joins the server and when it equips new emotes. It may be
// used by the server to find out which emotes the client has available. If the player has no emotes equipped,
// this packet is not sent.
// Under certain circumstances, this packet is also sent from the server to the client, but I was unable to
// find when this is done.
type EmoteList struct {
	// PlayerRuntimeID is the runtime ID of the player that owns the emote pieces below. If sent by the
	// client, this player runtime ID is always that of the player itself.
	PlayerRuntimeID uint64
	// EmotePieces is a list of emote pieces that the player with the runtime ID above has.
	EmotePieces []uuid.UUID
}

// ID ...
func (*EmoteList) ID() uint32 {
	return IDEmoteList
}

// Marshal ...
func (pk *EmoteList) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.PlayerRuntimeID)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.EmotePieces)))
	if len(pk.EmotePieces) > 6 {
		panic("player can have at most 6 emotes set")
	}
	for _, piece := range pk.EmotePieces {
		_ = protocol.WriteUUID(buf, piece)
	}
}

// Unmarshal ...
func (pk *EmoteList) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint64(&pk.PlayerRuntimeID)
	r.Varuint32(&count)
	r.LimitUint32(count, 6)

	pk.EmotePieces = make([]uuid.UUID, count)
	for i := uint32(0); i < count; i++ {
		r.UUID(&pk.EmotePieces[i])
	}
}
