package minecraft

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

// DefaultResourcePackChunkSize is the size of a single chunk of data from a resource pack sent by a
// Listener: 100 KiB, the size a vanilla host serves packs at. A client rejects a ResourcePackChunkData
// larger than 10 MiB outright, so any larger framing would have to stay under that.
const DefaultResourcePackChunkSize = 1024 * 100

// DefaultResourcePackMaxInFlightChunks is how many chunk requests a download keeps outstanding,
// matching the vanilla client request window.
const DefaultResourcePackMaxInFlightChunks = 100

// maxResourcePackPrealloc bounds the buffer reserved up front for a pack. The size comes from
// ResourcePacksInfo and is only a claim at that point, so a server would otherwise turn one small packet
// into an allocation of any size it names. Larger packs still grow their buffer as chunks arrive.
const maxResourcePackPrealloc = 32 << 20

// resourcePackQueue is used to aid in the handling of resource pack queueing and downloading. Only one
// resource pack is downloaded at a time.
type resourcePackQueue struct {
	packs           []*resource.Pack
	packsToDownload map[string]*resource.Pack
	currentPack     *resource.Pack
	currentOffset   uint64

	packAmount       int
	downloadingPacks map[string]*downloadingPack
	awaitingPacks    map[string]*downloadingPack
}

// downloadingPack is a resource pack that is being downloaded by a client connection.
type downloadingPack struct {
	buf        *bytes.Buffer
	chunkSize  uint32
	chunkCount uint32
	size       uint64
	newFrag    chan resourcePackChunk
	contentKey string
	cacheKey   ResourcePackCacheKey

	// mu guards requested, which tracks the chunk indices requested but not yet received.
	mu        sync.Mutex
	requested map[uint32]struct{}
}

// resourcePackChunk is a single received chunk of a resource pack, tagged with its index.
type resourcePackChunk struct {
	index uint32
	data  []byte
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
func (queue *resourcePackQueue) NextPack() (pk *packet.ResourcePackDataInfo, ok bool, err error) {
	for index, pack := range queue.packsToDownload {
		delete(queue.packsToDownload, index)
		chunkCount, ok := resourcePackChunkCount(uint64(pack.Size()), DefaultResourcePackChunkSize)
		if !ok {
			return nil, false, fmt.Errorf("resource pack %v has too many chunks", pack.UUID())
		}

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
			DataChunkSize: DefaultResourcePackChunkSize,
			ChunkCount:    chunkCount,
			Size:          uint64(pack.Size()),
			Hash:          checksum[:],
			PackType:      packType,
		}, true, nil
	}
	return nil, false, nil
}

// AllDownloaded checks if all resource packs in the queue are downloaded.
func (queue *resourcePackQueue) AllDownloaded() bool {
	return len(queue.packsToDownload) == 0
}

// resourcePackChunkCount returns the number of chunks a pack of size bytes is split into. It reports false if
// chunkSize is zero or if the resulting indices cannot be represented by ResourcePackChunkRequest.ChunkIndex,
// which is a signed int32.
func resourcePackChunkCount(size uint64, chunkSize uint32) (uint32, bool) {
	if chunkSize == 0 {
		return 0, false
	}
	count := size / uint64(chunkSize)
	if size%uint64(chunkSize) != 0 {
		count++
	}
	if count > uint64(1)<<31 {
		return 0, false
	}
	return uint32(count), true
}
