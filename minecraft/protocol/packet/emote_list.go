package packet

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// EmoteList is sent to allow clients to download emotes that other clients have equipped.
// TODO: Find when/how this is sent
type EmoteList struct {
	PlayerRuntimeID uint64
	// EmotePieces is a list of emote pieces.
	EmotePieces []uuid.UUID
}

// ID ...
func (*EmoteList) ID() uint32 {
	return IDEmoteList
}

// Marshal ...
func (e *EmoteList) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, e.PlayerRuntimeID)
	_ = protocol.WriteVaruint32(buf, uint32(len(e.EmotePieces)))
	for _, piece := range e.EmotePieces {
		_ = protocol.WriteUUID(buf, piece)
	}
}

// Unmarshal ...
func (e *EmoteList) Unmarshal(buf *bytes.Buffer) error {
	if err := protocol.Varuint64(buf, &e.PlayerRuntimeID); err != nil {
		return err
	}
	var count uint32
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	e.EmotePieces = make([]uuid.UUID, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.UUID(buf, &e.EmotePieces[i]); err != nil {
			return err
		}
	}
	return nil
}
