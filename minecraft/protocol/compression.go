package protocol

import (
	"bytes"
	"github.com/golang/snappy"
	"github.com/klauspost/compress/flate"
	"io"
)

// Compression represents a compression algorithm that can compress and decompress data.
type Compression interface {
	// EncodeCompression encodes the compression algorithm into a uint16 ID.
	EncodeCompression() uint16
	// Compress compresses the given data and returns the compressed data.
	Compress(decompressed []byte) ([]byte, error)
	// Decompress decompresses the given data and returns the decompressed data.
	Decompress(compressed []byte) ([]byte, error)
}

type (
	// FlateCompression is the implementation of the Flate compression algorithm. This was used by default until v1.19.30.
	FlateCompression struct{}
	// SnappyCompression is the implementation of the Snappy compression algorithm. This is used by default.
	SnappyCompression struct{}
)

// EncodeCompression ...
func (FlateCompression) EncodeCompression() uint16 {
	return 0
}

// Compress ...
func (FlateCompression) Compress(decompressed []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w, err := flate.NewWriter(buf, 6)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(decompressed)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress ...
func (FlateCompression) Decompress(compressed []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewBuffer(compressed))
	decompressed, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

// EncodeCompression ...
func (SnappyCompression) EncodeCompression() uint16 {
	return 1
}

// Compress ...
func (SnappyCompression) Compress(decompressed []byte) ([]byte, error) {
	return snappy.Encode(nil, decompressed), nil
}

// Decompress ...
func (SnappyCompression) Decompress(compressed []byte) ([]byte, error) {
	decompressed, err := snappy.Decode(nil, compressed)
	return decompressed, err
}

// init registers all valid compressions with the protocol.
func init() {
	RegisterCompression(FlateCompression{})
	RegisterCompression(SnappyCompression{})
}

var compressions = map[uint16]Compression{}

// RegisterCompression registers a compression so that it can be used by the protocol.
func RegisterCompression(compression Compression) {
	compressions[compression.EncodeCompression()] = compression
}

// CompressionByID attempts to return a compression by the ID it was registered with. If found, the compression found
// is returned and the bool is true.
func CompressionByID(id uint16) (Compression, bool) {
	c, ok := compressions[id]
	return c, ok
}
