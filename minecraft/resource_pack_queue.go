package minecraft

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

// resourcePackQueue is used to aid in the handling of resource pack queueing and downloading. Only one
// resource pack is downloaded at a time.
type resourcePackQueue struct {
	packs           []*resource.Pack
	packsToDownload map[string]*resource.Pack
	currentPack     *resource.Pack
	currentOffset   uint64

	packAmount       int
	downloadingPacks map[string]downloadingPack
	awaitingPacks    map[string]*downloadingPack
	chunkSize        uint32
}

// downloadingPack is a resource pack that is being downloaded by a client connection.
type downloadingPack struct {
	buf        *bytes.Buffer
	chunkSize  uint32
	size       uint64
	chunkCount uint32
	received   []bool
	newFrag    chan packChunk
	contentKey string
	cacheKey   ResourcePackCacheKey
}

// packChunk is a chunk of a downloadingPack received from the server, along with its index.
type packChunk struct {
	index uint32
	data  []byte
}

const (
	// maxPackChunks is the maximum number of chunks a downloaded resource pack may be split into, so that a
	// server cannot make the client track an unbounded number of chunks.
	maxPackChunks = 1 << 16
	// maxPackChunksInFlight is the maximum number of resource pack chunks requested but not yet received. It
	// stays below the burst of packets PocketMine-MP accepts from a client before kicking it (200), and above
	// the number of chunks some proxies hold back until they were all requested (100).
	maxPackChunksInFlight = 128
)

// packChunkCount returns the number of chunks a resource pack of size bytes is split into with chunks of
// chunkSize bytes.
func packChunkCount(size uint64, chunkSize uint32) (uint32, error) {
	if size == 0 || chunkSize == 0 {
		return 0, fmt.Errorf("invalid pack size %v or chunk size %v", size, chunkSize)
	}
	count := size / uint64(chunkSize)
	if size%uint64(chunkSize) != 0 {
		count++
	}
	if count > maxPackChunks {
		return 0, fmt.Errorf("pack of %v bytes has %v chunks of %v bytes, at most %v are allowed",
			size, count, chunkSize, maxPackChunks)
	}
	return uint32(count), nil
}

// Request 'requests' all resource packs passed, provided they all exist in the resourcePackQueue. Clients
// generally request packs as "uuid_version", and ResourcePackDataInfo must use that same identifier shape so
// the client can match the response to its request. Bare UUID requests are accepted for compatibility.
func (queue *resourcePackQueue) Request(packs []string) error {
	queue.packsToDownload = make(map[string]*resource.Pack)
	for _, requestedPackID := range packs {
		uuid, version, hasVersion := strings.Cut(requestedPackID, "_")
		found := false
		for _, pack := range queue.packs {
			if uuid == pack.UUID().String() && (!hasVersion || version == "" || version == pack.Version()) {
				queue.packsToDownload[pack.UUID().String()] = pack
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("resource pack (UUID=%v) not found", requestedPackID)
		}
	}
	return nil
}

// NextPack assigns the next resource pack to the current pack and returns true if successful. If there were
// no more packs to assign, false is returned. If ok is true, a packet with data info is returned.
func (queue *resourcePackQueue) NextPack() (pk *packet.ResourcePackDataInfo, ok bool) {
	for index, pack := range queue.packsToDownload {
		delete(queue.packsToDownload, index)

		queue.currentPack = pack
		queue.currentOffset = 0
		checksum := pack.Checksum()

		var packType byte
		switch {
		case pack.HasWorldTemplate():
			packType = packet.ResourcePackTypeWorldTemplate
		case pack.HasTextures() && (pack.HasBehaviours() || pack.HasScripts()):
			packType = packet.ResourcePackTypeAddon
		case !pack.HasTextures() && (pack.HasBehaviours() || pack.HasScripts()):
			packType = packet.ResourcePackTypeBehaviour
		case pack.HasTextures():
			packType = packet.ResourcePackTypeResources
		default:
			packType = packet.ResourcePackTypeSkins
		}
		return &packet.ResourcePackDataInfo{
			UUID:          pack.UUID().String() + "_" + pack.Version(),
			DataChunkSize: queue.chunkSize,
			ChunkCount:    uint32(pack.DataChunkCount(int(queue.chunkSize))),
			Size:          uint64(pack.Len()),
			Hash:          checksum[:],
			PackType:      packType,
		}, true
	}
	return nil, false
}

// AllDownloaded checks if all resource packs in the queue are downloaded.
func (queue *resourcePackQueue) AllDownloaded() bool {
	return len(queue.packsToDownload) == 0
}
