package packet

import (
	"bytes"
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
func (pk *ClientCacheMissResponse) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Blobs)))
	for _, blob := range pk.Blobs {
		_ = protocol.WriteBlob(buf, blob)
	}
}

// Unmarshal ...
func (pk *ClientCacheMissResponse) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	pk.Blobs = make([]protocol.CacheBlob, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.Blob(buf, &pk.Blobs[i]); err != nil {
			return err
		}
	}
	return nil
}
