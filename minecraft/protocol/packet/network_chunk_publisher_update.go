package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// NetworkChunkPublisherUpdate is sent by the server to change the point around which chunks are and remain
// loaded. This is useful for mini-game servers, where only one area is ever loaded, in which case the
// NetworkChunkPublisherUpdate packet can be sent in the middle of it, so that no chunks ever need to be
// additionally sent during the course of the game.
// In reality, the packet is not extraordinarily useful, and most servers just send it constantly at the
// position of the player.
// If the packet is not sent at all, no chunks will be shown to the player, regardless of where they are sent.
type NetworkChunkPublisherUpdate struct {
	// Position is the block position around which chunks loaded will remain shown to the client. Most servers
	// set this position to the position of the player itself.
	Position protocol.BlockPos
	// Radius is the radius in blocks around Position that chunks sent show up in and will remain loaded in.
	// Unlike the RequestChunkRadius and ChunkRadiusUpdated packets, this radius is in blocks rather than
	// chunks, so the chunk radius needs to be multiplied by 16. (Or shifted to the left by 4.)
	Radius uint32
	// SavedChunks ...
	// TODO: Figure out what this field is used for.
	SavedChunks []protocol.ChunkPos
}

// ID ...
func (*NetworkChunkPublisherUpdate) ID() uint32 {
	return IDNetworkChunkPublisherUpdate
}

// Marshal ...
func (pk *NetworkChunkPublisherUpdate) Marshal(w *protocol.Writer) {
	w.BlockPos(&pk.Position)
	w.Varuint32(&pk.Radius)

	l := uint32(len(pk.SavedChunks))
	w.Uint32(&l)
	for _, chunk := range pk.SavedChunks {
		w.ChunkPos(&chunk)
	}
}

// Unmarshal ...
func (pk *NetworkChunkPublisherUpdate) Unmarshal(r *protocol.Reader) {
	r.BlockPos(&pk.Position)
	r.Varuint32(&pk.Radius)

	var count uint32
	r.Uint32(&count)

	pk.SavedChunks = make([]protocol.ChunkPos, count)
	for i := uint32(0); i < count; i++ {
		r.ChunkPos(&pk.SavedChunks[i])
	}
}
