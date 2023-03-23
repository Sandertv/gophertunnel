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

func (pk *ClientCacheMissResponse) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Blobs)
}
