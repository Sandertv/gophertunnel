package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientCacheBlobStatus is part of the blob cache protocol. It is sent by the client to let the server know
// what blobs it needs and which blobs it already has, in an ACK type system.
type ClientCacheBlobStatus struct {
	// MissHashes is a list of blob hashes that the client does not have a blob available for. The server
	// should send the blobs matching these hashes as soon as possible.
	MissHashes []uint64
	// HitHashes is a list of blob hashes that the client has a blob available for. The blobs hashes here mean
	// that the client already has them: The server does not need to send the blobs anymore.
	HitHashes []uint64
}

// ID ...
func (pk *ClientCacheBlobStatus) ID() uint32 {
	return IDClientCacheBlobStatus
}

// Marshal ...
func (pk *ClientCacheBlobStatus) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.MissHashes)))
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.HitHashes)))
	for _, hash := range pk.MissHashes {
		_ = binary.Write(buf, binary.LittleEndian, hash)
	}
	for _, hash := range pk.HitHashes {
		_ = binary.Write(buf, binary.LittleEndian, hash)
	}
}

// Unmarshal ...
func (pk *ClientCacheBlobStatus) Unmarshal(r *protocol.Reader) {
	var hitCount, missCount uint32
	r.Varuint32(&missCount)
	r.Varuint32(&hitCount)

	r.LimitUint32(missCount+hitCount, 4096)

	pk.MissHashes = make([]uint64, missCount)
	pk.HitHashes = make([]uint64, hitCount)
	for i := uint32(0); i < missCount; i++ {
		r.Uint64(&pk.MissHashes[i])
	}
	for i := uint32(0); i < hitCount; i++ {
		r.Uint64(&pk.HitHashes[i])
	}
	return
}
