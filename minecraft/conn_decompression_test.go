package minecraft

import (
	"bytes"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestNormalizeMaxDecompressedLen(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "default", want: defaultMaxDecompressedLen},
		{name: "disabled", in: -1, want: math.MaxInt},
		{name: "custom", in: 4096, want: 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeMaxDecompressedLen(tt.in); got != tt.want {
				t.Fatalf("normalizeMaxDecompressedLen(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestConnNetworkSettingsEnforcesMaxDecompressedLen(t *testing.T) {
	encoded := encodeCompressedBatch(t, bytes.Repeat([]byte{0x01}, 64))
	conn := &Conn{
		enc:                packet.NewEncoder(io.Discard),
		dec:                packet.NewDecoder(bytes.NewReader(encoded)),
		maxDecompressedLen: 16,
	}

	if err := conn.handleNetworkSettings(&packet.NetworkSettings{
		CompressionThreshold: 0,
		CompressionAlgorithm: packet.CompressionAlgorithmFlate,
	}); err != nil {
		t.Fatalf("handleNetworkSettings: %v", err)
	}

	_, err := conn.dec.Decode()
	if err == nil {
		t.Fatal("expected decompression limit error")
	}
	if !strings.Contains(err.Error(), "exceeds limit 16") {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestConnNetworkSettingsAllowsPayloadWithinMaxDecompressedLen(t *testing.T) {
	payload := bytes.Repeat([]byte{0x01}, 16)
	encoded := encodeCompressedBatch(t, payload)
	conn := &Conn{
		enc:                packet.NewEncoder(io.Discard),
		dec:                packet.NewDecoder(bytes.NewReader(encoded)),
		maxDecompressedLen: 64,
	}

	if err := conn.handleNetworkSettings(&packet.NetworkSettings{
		CompressionThreshold: 0,
		CompressionAlgorithm: packet.CompressionAlgorithmFlate,
	}); err != nil {
		t.Fatalf("handleNetworkSettings: %v", err)
	}

	packets, err := conn.dec.Decode()
	if err != nil {
		t.Fatalf("decode compressed batch: %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("got %d packets, want 1", len(packets))
	}
	if !bytes.Equal(packets[0], payload) {
		t.Fatalf("decoded payload mismatch")
	}
}

func encodeCompressedBatch(t *testing.T, payload []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	enc := packet.NewEncoder(&buf)
	enc.EnableCompression(packet.FlateCompression, 0)
	if err := enc.Encode([][]byte{payload}); err != nil {
		t.Fatalf("encode compressed batch: %v", err)
	}
	return buf.Bytes()
}
