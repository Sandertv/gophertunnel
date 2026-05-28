package packet

import (
	"bytes"
	"testing"
)

func TestEncoderBatchEncodeObserverBelowThreshold(t *testing.T) {
	var out bytes.Buffer
	enc := NewEncoder(&out)
	enc.EnableCompression(SnappyCompression, 1024)

	var stats BatchEncodeStats
	enc.SetBatchEncodeObserver(func(s BatchEncodeStats) {
		stats = s
	})

	payload := []byte{1, 2, 3}
	if err := enc.Encode([][]byte{payload}); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if stats.PacketCount != 1 {
		t.Fatalf("PacketCount = %d, want 1", stats.PacketCount)
	}
	if !stats.BelowThreshold {
		t.Fatal("BelowThreshold = false, want true")
	}
	if stats.Compressed {
		t.Fatal("Compressed = true, want false")
	}
	if stats.CompressionID != CompressionAlgorithmNone {
		t.Fatalf("CompressionID = %d, want %d", stats.CompressionID, CompressionAlgorithmNone)
	}
	if stats.UncompressedLen == 0 || stats.OutputLen == 0 {
		t.Fatalf("stats lengths were not populated: %+v", stats)
	}
}

func TestEncoderBatchEncodeObserverCompressed(t *testing.T) {
	var out bytes.Buffer
	enc := NewEncoder(&out)
	enc.EnableCompression(SnappyCompression, 1)

	var stats BatchEncodeStats
	enc.SetBatchEncodeObserver(func(s BatchEncodeStats) {
		stats = s
	})

	payload := bytes.Repeat([]byte{7}, 2048)
	if err := enc.Encode([][]byte{payload}); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if !stats.Compressed {
		t.Fatal("Compressed = false, want true")
	}
	if stats.CompressionID != CompressionAlgorithmSnappy {
		t.Fatalf("CompressionID = %d, want %d", stats.CompressionID, CompressionAlgorithmSnappy)
	}
	if stats.MaxCompressedLen == 0 {
		t.Fatalf("MaxCompressedLen = 0, want populated stats: %+v", stats)
	}
	if stats.BufferCap == 0 || !stats.PooledBuffer {
		t.Fatalf("pool stats not populated as expected: %+v", stats)
	}
	if stats.UncompressedLen == 0 || stats.OutputLen == 0 {
		t.Fatalf("stats lengths were not populated: %+v", stats)
	}
}

func TestEncoderBatchEncodeObserverOutputLenIncludesEncryptionChecksum(t *testing.T) {
	var out bytes.Buffer
	enc := NewEncoder(&out)
	enc.EnableEncryption([32]byte{1})

	var stats BatchEncodeStats
	enc.SetBatchEncodeObserver(func(s BatchEncodeStats) {
		stats = s
	})

	payload := []byte{1, 2, 3}
	if err := enc.Encode([][]byte{payload}); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if got, want := stats.OutputLen, stats.UncompressedLen+8; got != want {
		t.Fatalf("OutputLen = %d, want %d", got, want)
	}
}
