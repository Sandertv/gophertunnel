package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePackChunkRequest is sent by the client to request a chunk of data from a particular resource pack,
// that it has obtained information about in a ResourcePackDataInfo packet.
type ResourcePackChunkRequest struct {
	// UUID is the unique ID of the resource pack that the chunk of data is requested from.
	UUID string
	// ChunkIndex is the requested chunk index of the chunk. It is a number that starts at 0 and is
	// incremented for each resource pack data chunk requested.
	ChunkIndex int32
}

// ID ...
func (*ResourcePackChunkRequest) ID() uint32 {
	return IDResourcePackChunkRequest
}

// Marshal ...
func (pk *ResourcePackChunkRequest) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.UUID)
	_ = binary.Write(buf, binary.LittleEndian, pk.ChunkIndex)
}

// Unmarshal ...
func (pk *ResourcePackChunkRequest) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.UUID),
		binary.Read(buf, binary.LittleEndian, &pk.ChunkIndex),
	)
}
