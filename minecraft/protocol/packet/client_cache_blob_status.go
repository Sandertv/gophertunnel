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
	protocol.FuncSliceOfLen(w, missLen, &pk.MissHashes, w.Uint64)
	protocol.FuncSliceOfLen(w, hitLen, &pk.HitHashes, w.Uint64)
}

// Unmarshal ...
func (pk *ClientCacheBlobStatus) Unmarshal(r *protocol.Reader) {
	missLen, hitLen := uint32(len(pk.MissHashes)), uint32(len(pk.HitHashes))
	r.Varuint32(&missLen)
	r.Varuint32(&hitLen)
	protocol.FuncSliceOfLen(r, missLen, &pk.MissHashes, r.Uint64)
	protocol.FuncSliceOfLen(r, hitLen, &pk.HitHashes, r.Uint64)
}
