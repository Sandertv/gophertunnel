package protocol

// CacheBlob represents a blob as used in the client side blob cache protocol. It holds a hash of its data and
// the full data of it.
type CacheBlob struct {
	// Hash is the hash of the blob. The hash is computed using xxHash, and must be deterministic for the same
	// chunk data.
	Hash uint64
	// Payload is the data of the blob. When sent, the client will associate the Hash of the blob with the
	// Payload in it.
	Payload []byte
}

// Blob reads/writes a CacheBlob x using IO r.
func Blob(r IO, x *CacheBlob) {
	r.Uint64(&x.Hash)
	r.ByteSlice(&x.Payload)
}
