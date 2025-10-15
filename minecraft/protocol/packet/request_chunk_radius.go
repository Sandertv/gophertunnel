package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RequestChunkRadius is sent by the client to the server to update the server on the chunk view radius that
// it has set in the settings. The server may respond with a ChunkRadiusUpdated packet with either the chunk
// radius requested, or a different chunk radius if the server chooses so.
type RequestChunkRadius struct {
	// ChunkRadius is the requested chunk radius. This value is always the value set in the settings of the
	// player.
	ChunkRadius int32
	// MaxChunkRadius is the maximum chunk radius that the player wants to receive. The reason for the client sending this
	// is currently unknown.
	MaxChunkRadius uint8
}

// ID ...
func (*RequestChunkRadius) ID() uint32 {
	return IDRequestChunkRadius
}

func (pk *RequestChunkRadius) Marshal(io protocol.IO) {
	io.Varint32(&pk.ChunkRadius)
	io.Uint8(&pk.MaxChunkRadius)
}
