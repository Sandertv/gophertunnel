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
	// data holds chunks [0, next), each run of them placed at its first chunk's DataOffset.
	data []byte
	// early holds chunks that arrived before an earlier one, by index.
	early map[uint32][]byte
	// receivedBytes is the total length of every new chunk received.
	receivedBytes uint64
	// heldBytes is the memory counted for data and early: their lengths plus packChunkOverhead per early chunk.
	heldBytes  uint64
	size       uint64
	chunkCount uint32
	// next is the index of the first chunk not yet placed into data.
	next uint32
	// requested is the index of the next chunk to request.
	requested uint32
	// complete is set once every chunk was placed and the pack was handed off for reading.
	complete   bool
	contentKey string
	cacheKey   ResourcePackCacheKey
}

const (
	// initialPackChunkRequests is how many chunks of a resource pack are requested up front, the same as the
	// vanilla client. After that, one more is requested for every chunk packet received.
	initialPackChunkRequests = 100
	// packChunkOverhead is the memory counted for an early chunk on top of its data.
	packChunkOverhead = 128
	// packHeldSlack is the memory a download may hold beyond twice the data received, see packHeldLimit.
	packHeldSlack = 1 << 20
)

// packHeldLimit returns how much counted memory a download may hold: twice the data received so far plus
// packHeldSlack. The vanilla client has no such limit, as it writes to a file while this keeps the pack in
// memory. Only chunks that hold far less data than the memory they take reach it: many empty chunks waiting
// for earlier ones, or offsets far past the data received.
func packHeldLimit(pack *downloadingPack) uint64 {
	return 2*pack.receivedBytes + packHeldSlack
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
