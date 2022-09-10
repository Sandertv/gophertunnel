package packet

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"github.com/sandertv/gophertunnel/internal"
	"io"
	"sync"
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

// FlateCompression is the implementation of the Flate compression algorithm.
type FlateCompression struct{}

var (
	// flateDecompressPool is a sync.Pool for io.ReadCloser flate readers. These are pooled for connections.
	flateDecompressPool = sync.Pool{
		New: func() any { return flate.NewReader(bytes.NewReader(nil)) },
	}
	// flateCompressPool is a sync.Pool for io.ReadCloser flate writers. These are pooled for connections.
	flateCompressPool = sync.Pool{
		New: func() any {
			w, _ := flate.NewWriter(io.Discard, 6)
			return w
		},
	}
)

// EncodeCompression ...
func (FlateCompression) EncodeCompression() uint16 {
	return 0
}

// Compress ...
func (FlateCompression) Compress(decompressed []byte) ([]byte, error) {
	compressed := internal.BufferPool.Get().(*bytes.Buffer)
	w := flateCompressPool.Get().(*flate.Writer)

	defer func() {
		// Reset the buffer, so we can return it to the buffer pool safely.
		compressed.Reset()
		internal.BufferPool.Put(compressed)
		flateCompressPool.Put(w)
	}()

	w.Reset(compressed)

	_, err := w.Write(decompressed)
	if err != nil {
		return nil, fmt.Errorf("compress flate: %w", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("close flate writer: %w", err)
	}
	return compressed.Bytes(), nil
}

// Decompress ...
func (FlateCompression) Decompress(compressed []byte) ([]byte, error) {
	buf := bytes.NewReader(compressed)
	c := flateDecompressPool.Get().(io.ReadCloser)
	defer flateDecompressPool.Put(c)

	if err := c.(flate.Resetter).Reset(buf, nil); err != nil {
		return nil, fmt.Errorf("reset flate: %w", err)
	}
	_ = c.Close()

	// Guess an uncompressed size of 2*len(compressed).
	decompressed := bytes.NewBuffer(make([]byte, 0, len(compressed)*2))
	if _, err := io.Copy(decompressed, c); err != nil {
		return nil, fmt.Errorf("decompress flate: %v", err)
	}
	return decompressed.Bytes(), nil
}

// init registers all valid compressions with the protocol.
func init() {
	RegisterCompression(FlateCompression{})
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
