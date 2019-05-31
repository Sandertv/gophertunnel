package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePackDataInfo is sent by the server to the client to inform the client about the data contained in
// one of the resource packs that are about to be sent.
type ResourcePackDataInfo struct {
	// UUID is the unique ID of the resource pack that the info concerns.
	UUID string
	// DataChunkSize is the maximum size in bytes of the chunks in which the total size of the resource pack
	// to be sent will be divided. A size of 1MB (1024*1024) means that a resource pack of 15.5MB will be
	// split into 16 data chunks.
	DataChunkSize int32
	// ChunkCount is the total amount of data chunks that the sent resource pack will exist out of. It is the
	// total size of the resource pack divided by the DataChunkSize field.
	ChunkCount int32
	// Size is the total size in bytes that the resource pack occupies. This is the size of the compressed
	// archive (zip) of the resource pack.
	Size int64
	// Hash is a SHA256 hash of the content of the resource pack.
	Hash string
}

// ID ...
func (*ResourcePackDataInfo) ID() uint32 {
	return IDResourcePackDataInfo
}

// Marshal ...
func (pk *ResourcePackDataInfo) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.UUID)
	_ = binary.Write(buf, binary.LittleEndian, pk.DataChunkSize)
	_ = binary.Write(buf, binary.LittleEndian, pk.ChunkCount)
	_ = binary.Write(buf, binary.LittleEndian, pk.Size)
	_ = protocol.WriteString(buf, pk.Hash)
}

// Unmarshal ...
func (pk *ResourcePackDataInfo) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.UUID),
		binary.Read(buf, binary.LittleEndian, &pk.DataChunkSize),
		binary.Read(buf, binary.LittleEndian, &pk.ChunkCount),
		binary.Read(buf, binary.LittleEndian, &pk.Size),
		protocol.String(buf, &pk.Hash),
	)
}
