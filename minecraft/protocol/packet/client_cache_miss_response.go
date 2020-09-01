package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientCacheMissResponse is part of the blob cache protocol. It is sent by the server in response to a
// ClientCacheBlobStatus packet and contains the blob data of all blobs that the client acknowledged not to
// have yet.
type ClientCacheMissResponse struct {
	// Blobs is a list of all blobs that the client sent misses for in the ClientCacheBlobStatus. These blobs
	// hold the data of the blobs with the hashes they are matched with.
	Blobs []protocol.CacheBlob
}

// ID ...
func (pk *ClientCacheMissResponse) ID() uint32 {
	return IDClientCacheMissResponse
}

// Marshal ...
func (pk *ClientCacheMissResponse) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Blobs))
	w.Varuint32(&l)
	for _, blob := range pk.Blobs {
		protocol.Blob(w, &blob)
	}
}

// Unmarshal ...
func (pk *ClientCacheMissResponse) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	pk.Blobs = make([]protocol.CacheBlob, count)
	for i := uint32(0); i < count; i++ {
		protocol.Blob(r, &pk.Blobs[i])
	}
}
