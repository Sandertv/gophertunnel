package minecraft

import (
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
	// uuid is the identifier of the pack as sent in ResourcePackDataInfo, used to request its chunks.
	uuid string
	// chunks holds every chunk received so far by its index. Memory grows only with the data that actually
	// arrives, as the sizes in ResourcePacksInfo and ResourcePackDataInfo are only claims by the server.
	chunks     map[uint32]packChunk
	size       uint64
	chunkCount uint32
	// requested is the index of the next chunk to request.
	requested  uint32
	contentKey string
	cacheKey   ResourcePackCacheKey
}

// packChunk is one chunk of a downloading resource pack, placed at its offset.
type packChunk struct {
	offset uint64
	data   []byte
}

// maxPackChunksInFlight is the maximum number of chunks of a resource pack requested but not yet received,
// the same as the vanilla client. Some proxies hold back a chunk until the 99 after it were requested, so it
// must not be lower or shared between packs.
const maxPackChunksInFlight = 100

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
