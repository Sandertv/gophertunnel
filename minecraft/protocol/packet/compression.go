package packet

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/snappy"
	"github.com/sandertv/gophertunnel/minecraft/internal"
)

// Compression represents a compression algorithm that can compress and decompress data.
type Compression interface {
	// EncodeCompression encodes the compression algorithm into a uint16 ID.
	EncodeCompression() uint16
	// Compress compresses the given data and returns the compressed data.
	Compress(decompressed []byte) ([]byte, error)
	// Decompress decompresses the given data and returns the decompressed data.
	Decompress(compressed []byte, limit int) ([]byte, error)
}

var (
	// NopCompression is an empty implementation that does not compress data.
	NopCompression nopCompression
	// FlateCompression is the implementation of the Flate compression
	// algorithm. This is used by default.
	FlateCompression flateCompression
	// SnappyCompression is the implementation of the Snappy compression
	// algorithm. Snappy currently crashes devices without `avx2`.
	SnappyCompression snappyCompression

	DefaultCompression Compression = FlateCompression
)

type (
	// nopCompression is an empty implementation that does not compress data.
	nopCompression struct{}
	// flateCompression is the implementation of the Flate compression algorithm.
	flateCompression struct{}
	// snappyCompression is the implementation of the Snappy compression algorithm.
	snappyCompression struct{}
)

// flateDecompressPool is a sync.Pool for io.ReadCloser flate readers. These are
// pooled for connections.
var (
	flateDecompressPool = sync.Pool{
		New: func() any { return flate.NewReader(bytes.NewReader(nil)) },
	}
	flateCompressPool = sync.Pool{
		New: func() any {
			w, _ := flate.NewWriter(io.Discard, 6)
			return w
		},
	}
)

// EncodeCompression ...
func (nopCompression) EncodeCompression() uint16 {
	return CompressionAlgorithmNone
}

// Compress ...
func (nopCompression) Compress(decompressed []byte) ([]byte, error) {
	return decompressed, nil
}

// Decompress ...
func (nopCompression) Decompress(compressed []byte, limit int) ([]byte, error) {
	if len(compressed) > limit {
		return nil, fmt.Errorf("nop decompression: size %d exceeds limit %d", len(compressed), limit)
	}
	return compressed, nil
}

// EncodeCompression ...
func (flateCompression) EncodeCompression() uint16 {
	return CompressionAlgorithmFlate
}

// Compress ...
func (flateCompression) Compress(decompressed []byte) ([]byte, error) {
	compressed := internal.BufferPool.Get().(*bytes.Buffer)
	w := flateCompressPool.Get().(*flate.Writer)

	defer func() {
		// Reset the buffer, so we can return it to the buffer pool safely.
		compressed.Reset()
		internal.BufferPool.Put(compressed)
		flateCompressPool.Put(w)
	}()

	w.Reset(compressed)
	if _, err := w.Write(decompressed); err != nil {
		return nil, fmt.Errorf("compress flate: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close flate writer: %w", err)
	}
	return append([]byte(nil), compressed.Bytes()...), nil
}

// Decompress ...
func (flateCompression) Decompress(compressed []byte, limit int) ([]byte, error) {
	r := flateDecompressPool.Get().(io.ReadCloser)
	defer func() {
		_ = r.Close()
		flateDecompressPool.Put(r)
	}()

	if err := r.(flate.Resetter).Reset(bytes.NewReader(compressed), nil); err != nil {
		return nil, fmt.Errorf("reset flate: %w", err)
	}

	decompressed := internal.BufferPool.Get().(*bytes.Buffer)
	defer func() {
		// Only return reasonably sized buffers to the pool to avoid retaining very large arrays.
		if decompressed.Cap() <= 1<<20 { // 1 MiB cap
			decompressed.Reset()
			internal.BufferPool.Put(decompressed)
		}
	}()

	// Handle no limit
	if limit == math.MaxInt {
		if _, err := io.Copy(decompressed, r); err != nil {
			return nil, fmt.Errorf("decompress flate: %w", err)
		}
		return append([]byte(nil), decompressed.Bytes()...), nil
	}

	// If the compressed data is less than half the limit, we can safely assume l*2, otherwise cap at limit.
	capHint := limit
	if l := len(compressed); l <= limit/2 {
		capHint = l * 2
	}
	decompressed.Grow(capHint)

	// Read limit+1 bytes to detect overflow
	lr := &io.LimitedReader{R: r, N: int64(limit) + 1}
	if _, err := io.Copy(decompressed, lr); err != nil {
		return nil, fmt.Errorf("decompress flate: %w", err)
	}
	if lr.N <= 0 {
		return nil, fmt.Errorf("decompress flate: size exceeds limit %d", limit)
	}
	return append([]byte(nil), decompressed.Bytes()...), nil
}

// EncodeCompression ...
func (snappyCompression) EncodeCompression() uint16 {
	return CompressionAlgorithmSnappy
}

// Compress ...
func (snappyCompression) Compress(decompressed []byte) ([]byte, error) {
	// Because Snappy allocates a slice only once, it is less important to have
	// a dst slice pre-allocated. With flateCompression this is more important,
	// because flate does a lot of smaller allocations which causes a
	// considerable slowdown.
	return snappy.Encode(nil, decompressed), nil
}

// Decompress ...
func (snappyCompression) Decompress(compressed []byte, limit int) ([]byte, error) {
	// Snappy writes a decoded data length prefix, so it can allocate the
	// perfect size right away and only needs to allocate once. No need to pool
	// byte slices here either.
	decodedLen, err := snappy.DecodedLen(compressed)
	if err != nil {
		return nil, fmt.Errorf("snappy decoded length: %w", err)
	}
	if decodedLen > limit {
		return nil, fmt.Errorf("snappy decoded size %d exceeds limit %d", decodedLen, limit)
	}
	decompressed, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, fmt.Errorf("decompress snappy: %w", err)
	}
	return decompressed, nil
}

// init registers all valid compressions with the protocol.
func init() {
	RegisterCompression(flateCompression{})
	RegisterCompression(snappyCompression{})
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
	if !ok {
		c = DefaultCompression
	}
	return c, ok
}
