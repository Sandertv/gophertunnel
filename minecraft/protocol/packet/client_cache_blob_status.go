package packet

import (
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
func (pk *ClientCacheBlobStatus) Marshal(w *protocol.Writer) {
	missLen, hitLen := uint32(len(pk.MissHashes)), uint32(len(pk.HitHashes))
	w.Varuint32(&missLen)
	w.Varuint32(&hitLen)
	for _, hash := range pk.MissHashes {
		w.Uint64(&hash)
	}
	for _, hash := range pk.HitHashes {
		w.Uint64(&hash)
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
}
