package minecraft

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestResourcePackDownloadConfigNormalized(t *testing.T) {
	for _, test := range []struct {
		name   string
		config ResourcePackDownloadConfig
		want   int
	}{
		{name: "zero", want: DefaultResourcePackMaxInFlightChunks},
		{name: "negative", config: ResourcePackDownloadConfig{MaxInFlightChunks: -1}, want: DefaultResourcePackMaxInFlightChunks},
		{name: "custom", config: ResourcePackDownloadConfig{MaxInFlightChunks: 7}, want: 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.config.normalized().MaxInFlightChunks; got != test.want {
				t.Fatalf("MaxInFlightChunks = %d, want %d", got, test.want)
			}
		})
	}
}

func TestResourcePackChunkCount(t *testing.T) {
	for _, test := range []struct {
		name      string
		size      uint64
		chunkSize uint32
		want      uint32
		wantOK    bool
	}{
		{name: "empty", chunkSize: 16, wantOK: true},
		{name: "exact", size: 32, chunkSize: 16, want: 2, wantOK: true},
		{name: "partial", size: 33, chunkSize: 16, want: 3, wantOK: true},
		{name: "zero chunk size", size: 32},
		{name: "max index", size: 1 << 31, chunkSize: 1, want: 1 << 31, wantOK: true},
		{name: "index overflows int32", size: 1<<31 + 1, chunkSize: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := resourcePackChunkCount(test.size, test.chunkSize)
			if ok != test.wantOK {
				t.Fatalf("ok = %v, want %v", ok, test.wantOK)
			}
			if got != test.want {
				t.Fatalf("count = %d, want %d", got, test.want)
			}
		})
	}
}

func TestResourcePackDownloadReplenishesAfterOutOfOrderChunk(t *testing.T) {
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	go func() { _, _ = io.Copy(io.Discard, peer) }()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	defer conn.Close()

	const id = "550e8400-e29b-41d4-a716-446655440000"
	pack := &downloadingPack{buf: new(bytes.Buffer), size: 200}
	conn.packQueue = &resourcePackQueue{
		downloadingPacks: map[string]*downloadingPack{id: pack},
		awaitingPacks:    make(map[string]*downloadingPack),
	}
	if err := conn.handleResourcePackDataInfo(&packet.ResourcePackDataInfo{
		UUID:          id + "_1.0.0",
		DataChunkSize: 1,
		Size:          200,
	}); err != nil {
		t.Fatalf("handleResourcePackDataInfo: %v", err)
	}

	if !waitForResourcePackRequest(t, pack, 99) {
		t.Fatal("vanilla request window was not filled")
	}
	pack.mu.Lock()
	_, requestedEarly := pack.requested[100]
	requestCount := len(pack.requested)
	pack.mu.Unlock()
	if requestedEarly {
		t.Fatal("request window exceeded before a chunk was received")
	}
	if requestCount != DefaultResourcePackMaxInFlightChunks {
		t.Fatalf("initial request count = %d, want %d", requestCount, DefaultResourcePackMaxInFlightChunks)
	}
	if err := conn.handleResourcePackChunkData(&packet.ResourcePackChunkData{
		UUID:       id + "_1.0.0",
		ChunkIndex: 50,
		Data:       []byte{50},
	}); err != nil {
		t.Fatalf("handleResourcePackChunkData: %v", err)
	}
	if !waitForResourcePackRequest(t, pack, 100) {
		t.Fatal("out-of-order response did not replenish the request window")
	}
}

func waitForResourcePackRequest(t *testing.T, pack *downloadingPack, index uint32) bool {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		pack.mu.Lock()
		_, ok := pack.requested[index]
		pack.mu.Unlock()
		if ok {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}
