package protocol

import (
	"bytes"
	"encoding/binary"
)

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

// WriteBlob writes a CacheBlob x to Buffer dst.
func WriteBlob(dst *bytes.Buffer, x CacheBlob) error {
	return chainErr(
		binary.Write(dst, binary.LittleEndian, x.Hash),
		WriteByteSlice(dst, x.Payload),
	)
}

// Blob reads a CacheBlob x from Buffer src.
func Blob(src *bytes.Buffer, x *CacheBlob) error {
	return chainErr(
		binary.Read(src, binary.LittleEndian, &x.Hash),
		ByteSlice(src, &x.Payload),
	)
}
