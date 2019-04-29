package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePackChunkData is sent to the client so that the client can download the resource pack. Each packet
// holds a chunk of the compressed resource pack, of which the size is defined in the ResourcePackDataInfo
// packet sent before.
type ResourcePackChunkData struct {
	// UUID is the unique ID of the resource pack that the chunk of data is taken out of.
	UUID string
	// ChunkIndex is the current chunk index of the chunk. It is a number that starts at 0 and is incremented
	// for each resource pack data chunk sent to the client.
	ChunkIndex int32
	// DataOffset is the current progress in bytes or offset in the data that the resource pack data chunk is
	// taken from.
	DataOffset int64
	// Data is a byte slice containing a chunk of data from the resource pack. It must be of the same size or
	// less than the DataChunkSize set in the ResourcePackDataInfo packet.
	Data []byte
}

// ID ...
func (*ResourcePackChunkData) ID() uint32 {
	return IDResourcePackChunkData
}

// Marshal ...
func (pk *ResourcePackChunkData) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.UUID)
	_ = binary.Write(buf, binary.LittleEndian, pk.ChunkIndex)
	_ = binary.Write(buf, binary.LittleEndian, pk.DataOffset)
	_ = binary.Write(buf, binary.LittleEndian, int32(len(pk.Data)))
	_, _ = buf.Write(pk.Data)
}

// Unmarshal ...
func (pk *ResourcePackChunkData) Unmarshal(buf *bytes.Buffer) error {
	if err := protocol.String(buf, &pk.UUID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &pk.ChunkIndex); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &pk.DataOffset); err != nil {
		return err
	}
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return err
	}
	pk.Data = buf.Next(int(length))
	return nil
}
