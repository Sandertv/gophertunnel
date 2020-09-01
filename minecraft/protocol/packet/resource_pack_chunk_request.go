package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePackChunkRequest is sent by the client to request a chunk of data from a particular resource pack,
// that it has obtained information about in a ResourcePackDataInfo packet.
type ResourcePackChunkRequest struct {
	// UUID is the unique ID of the resource pack that the chunk of data is requested from.
	UUID string
	// ChunkIndex is the requested chunk index of the chunk. It is a number that starts at 0 and is
	// incremented for each resource pack data chunk requested.
	ChunkIndex uint32
}

// ID ...
func (*ResourcePackChunkRequest) ID() uint32 {
	return IDResourcePackChunkRequest
}

// Marshal ...
func (pk *ResourcePackChunkRequest) Marshal(w *protocol.Writer) {
	w.String(&pk.UUID)
	w.Uint32(&pk.ChunkIndex)
}

// Unmarshal ...
func (pk *ResourcePackChunkRequest) Unmarshal(r *protocol.Reader) {
	r.String(&pk.UUID)
	r.Uint32(&pk.ChunkIndex)
}
