package minecraft

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
)

func (conn *Conn) Logger() *slog.Logger {
	return conn.log
}

func (conn *Conn) SetGameData(data GameData) {
	conn.gameData = data
}

func (conn *Conn) SetShieldID(id int32) {
	conn.shieldID.Store(id)
}

func (conn *Conn) ShieldID() int32 {
	return conn.shieldID.Load()
}

// MarkLoggedIn marks the connection as logged in, used by the login flow once the entire login sequence has
// completed.
func (conn *Conn) MarkLoggedIn() {
	conn.loggedIn = true
}
func (conn *Conn) CloseWithCause(cause error) error {
	return conn.close(cause)
}

func (conn *Conn) CloseErr(op string) error {
	return conn.closeErr(op)
}

// GameDataReceived marks the connection's game data as received and attempts to finalise the client
// connection if both the ChunkRadiusUpdated and PlayStatus packets have been seen. It is only relevant for
// client sided connections.
func (conn *Conn) GameDataReceived() {
	conn.gameDataReceived.Store(true)
	conn.tryFinaliseClientConn()
}

func (conn *Conn) Spawned() bool {
	if conn.waitingForSpawn.CompareAndSwap(true, false) {
		close(conn.spawn)
		return true
	}
	return false
}

func (conn *Conn) ResetResourcePackQueue(totalPacks int) {
	conn.packQueue = &resourcePackQueue{
		packAmount:       totalPacks,
		downloadingPacks: make(map[string]downloadingPack),
		awaitingPacks:    make(map[string]*downloadingPack),
	}
}

func (conn *Conn) HandleResourcePacksInfo(pk *packet.ResourcePacksInfo) error {
	totalPacks := len(pk.TexturePacks)
	conn.ResetResourcePackQueue(totalPacks)
	packsToDownload := make([]string, 0, totalPacks)

	for index, pack := range pk.TexturePacks {
		id := pack.UUID.String()
		if conn.PackAlreadyDownloading(id) {
			conn.Logger().Warn("handle ResourcePacksInfo: duplicate texture pack", "UUID", pack.UUID)
			conn.DecrementPackAmount()
			continue
		}
		if !conn.ShouldDownloadResourcePack(uuid.MustParse(id), pack.Version, index, totalPacks) {
			conn.IgnoreResourcePack(id, pack.Version)
			conn.DecrementPackAmount()
			continue
		}

		cacheKey := ResourcePackCacheKey{
			UUID:    pack.UUID,
			Version: pack.Version,
			Size:    pack.Size,
		}
		if conn.HasResourcePackCache() {
			cachedPack, err := conn.LoadResourcePackFromCache(cacheKey)
			switch {
			case err != nil:
				conn.Logger().Warn("handle ResourcePacksInfo: failed to load resource pack from cache", "UUID", pack.UUID, "version", pack.Version, "err", err)
			case cachedPack != nil && !cacheKey.Matches(cachedPack):
				conn.Logger().Warn("handle ResourcePacksInfo: cached resource pack did not match advertised pack", "UUID", pack.UUID, "version", pack.Version, "cached_UUID", cachedPack.UUID(), "cached_version", cachedPack.Version(), "cached_size", cachedPack.Len())
			case cachedPack != nil:
				conn.AppendResourcePacks(cachedPack.WithContentKey(pack.ContentKey))
				conn.DecrementPackAmount()
				continue
			}
		}

		// This UUID_Version is a hack Mojang put in place.
		packsToDownload = append(packsToDownload, id+"_"+pack.Version)
		conn.RegisterPackDownload(id, pack.Size, pack.ContentKey, cacheKey, make(chan []byte))
	}

	if conn.PendingPackDownloads() != 0 {
		_ = conn.WritePacket(&packet.ResourcePackClientResponse{
			Response:        packet.PackResponseSendPacks,
			PacksToDownload: packsToDownload,
		})
		return nil
	}
	_ = conn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded})

	if conn.PendingPackDownloads() != 0 {
		// We told the server which packs we need downloaded, so we expect the first ResourcePackDataInfo and
		// ResourcePackChunkData packets next.
		conn.expect(packet.IDResourcePackDataInfo, packet.IDResourcePackChunkData)
		return nil
	}
	conn.expect(packet.IDResourcePackStack)
	return nil
}

// PackAlreadyDownloading reports whether a resource pack with the given id is already registered in the
// download queue.
func (conn *Conn) PackAlreadyDownloading(id string) bool {
	if conn.packQueue == nil {
		return false
	}
	_, ok := conn.packQueue.downloadingPacks[id]
	return ok
}

func (conn *Conn) DecrementPackAmount() {
	if conn.packQueue == nil {
		return
	}
	conn.packQueue.packAmount--
}

func (conn *Conn) PendingPackDownloads() int {
	if conn.packQueue == nil {
		return 0
	}
	return len(conn.packQueue.downloadingPacks)
}

// RegisterPackDownload registers a resource pack to be downloaded. id is the pack's "uuid_version" form, or
// a bare UUID. The pack is added to the download queue with the given size, content key, cache key and a
// fragment channel that receives each downloaded chunk.
func (conn *Conn) RegisterPackDownload(id string, size uint64, contentKey string, cacheKey ResourcePackCacheKey, newFrag chan []byte) {
	if conn.packQueue == nil {
		conn.ResetResourcePackQueue(0)
	}
	conn.packQueue.downloadingPacks[id] = downloadingPack{
		size:       size,
		buf:        bytes.NewBuffer(make([]byte, 0, size)),
		newFrag:    newFrag,
		contentKey: contentKey,
		cacheKey:   cacheKey,
	}
}

// ShouldDownloadResourcePack reports whether the resource pack with the given UUID and version should be
// downloaded, using the connection's download filter if one is configured. When no filter is set, the pack
// is always downloaded.
func (conn *Conn) ShouldDownloadResourcePack(id uuid.UUID, version string, currentPack, totalPacks int) bool {
	if conn.downloadResourcePack == nil {
		return true
	}
	return conn.downloadResourcePack(id, version, currentPack, totalPacks)
}

// IgnoreResourcePack adds a resource pack to the set of packs that are exempt from download because the
// connection's download filter rejected them. Such packs may still be applied in the ResourcePackStack.
func (conn *Conn) IgnoreResourcePack(id, version string) {
	conn.ignoredResourcePacks = append(conn.ignoredResourcePacks, exemptedResourcePack{uuid: id, version: version})
}

// HasResourcePackCache reports whether the connection has a resource pack cache configured.
func (conn *Conn) HasResourcePackCache() bool {
	return conn.resourcePackCache != nil
}

// LoadResourcePackFromCache looks up a resource pack in the connection's resource pack cache. A nil pack and
// nil error are returned when the cache is not configured or does not hold the pack.
func (conn *Conn) LoadResourcePackFromCache(key ResourcePackCacheKey) (*resource.Pack, error) {
	if conn.resourcePackCache == nil {
		return nil, nil
	}
	return conn.resourcePackCache.Load(conn.ctx, key)
}

func (conn *Conn) CacheResourcePack(key ResourcePackCacheKey, pack *resource.Pack) {
	conn.storeResourcePack(key, pack)
}

func (conn *Conn) AppendResourcePacks(packs ...*resource.Pack) {
	conn.resourcePacks = append(conn.resourcePacks, packs...)
}

func (conn *Conn) ResourcePackDelivery() ResourcePackDeliveryConfig {
	return conn.resourcePackDelivery
}

func (conn *Conn) HasResourcePack(uuidStr, version string, hasBehaviours bool) bool {
	return conn.hasPack(uuidStr, version, hasBehaviours)
}

func (conn *Conn) SendNextResourcePack() error {
	return conn.nextResourcePackDownload()
}

func (conn *Conn) HandleResourcePackClientResponse(pk *packet.ResourcePackClientResponse) error {
	switch pk.Response {
	case packet.PackResponseRefused:
		return conn.CloseWithCause(conn.CloseErr("resource pack refused"))
	case packet.PackResponseSendPacks:
		if err := conn.serveRequestedResourcePacks(pk.PacksToDownload); err != nil {
			return err
		}
		conn.expect(packet.IDResourcePackChunkRequest)
	case packet.PackResponseAllPacksDownloaded:
		if err := conn.WritePacket(conn.TexturePackStack()); err != nil {
			return fmt.Errorf("send ResourcePackStack: %w", err)
		}
	case packet.PackResponseCompleted:
		conn.MarkLoggedIn()
	default:
		return fmt.Errorf("unknown ResourcePackClientResponse response type %v", pk.Response)
	}
	return nil
}

func (conn *Conn) serveRequestedResourcePacks(packs []string) error {
	conn.packQueue = &resourcePackQueue{
		packs:     conn.resourcePacks,
		chunkSize: conn.resourcePackDelivery.ChunkSize,
	}
	if err := conn.packQueue.Request(packs); err != nil {
		return fmt.Errorf("lookup resource packs by UUID: %w", err)
	}
	return conn.nextResourcePackDownload()
}

// TexturePackStack returns the ResourcePackStack packet that should be sent once all resource packs have
// been downloaded, containing the packs the connection holds and the exempted packs.
func (conn *Conn) TexturePackStack() *packet.ResourcePackStack {
	pk := &packet.ResourcePackStack{BaseGameVersion: protocol.CurrentVersion, Experiments: []protocol.ExperimentData{{Name: "cameras", Enabled: true}}}
	for _, pack := range conn.resourcePacks {
		pk.TexturePacks = append(pk.TexturePacks, protocol.StackResourcePack{UUID: pack.UUID().String(), Version: pack.Version()})
	}
	for _, exempted := range exemptedPacks {
		pk.TexturePacks = append(pk.TexturePacks, protocol.StackResourcePack{UUID: exempted.uuid, Version: exempted.version})
	}
	return pk
}

func (conn *Conn) HandleResourcePackStack(pk *packet.ResourcePackStack) error {
	// We currently don't apply resource packs in any way, so instead we just check if all resource packs in
	// the stacks are also downloaded.
	for _, pack := range pk.TexturePacks {
		if !conn.HasResourcePack(pack.UUID, pack.Version, false) {
			return fmt.Errorf("texture pack (UUID=%v, version=%v) not downloaded", pack.UUID, pack.Version)
		}
	}
	return conn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseCompleted})
}

func (conn *Conn) BeginPackDataInfo(pk *packet.ResourcePackDataInfo) error {
	id, _, _ := strings.Cut(pk.UUID, "_")

	pack, ok := conn.packQueue.downloadingPacks[id]
	if !ok {
		// We either already downloaded the pack or we got sent an invalid UUID, that did not match any pack
		// sent in the ResourcePacksInfo packet.
		return fmt.Errorf("handle ResourcePackDataInfo: unknown pack (UUID=%v)", id)
	}
	if pack.size != pk.Size {
		// Size mismatch: the ResourcePacksInfo packet had a size for the pack that did not match with the
		// size sent here.
		conn.log.Warn("handle ResourcePackDataInfo: pack had a different size in ResourcePacksInfo than in ResourcePackDataInfo", "UUID", id)
		pack.size = pk.Size
	}

	// Remove the resource pack from the downloading packs and add it to the awaiting packets.
	delete(conn.packQueue.downloadingPacks, id)
	conn.packQueue.awaitingPacks[id] = &pack

	pack.chunkSize = pk.DataChunkSize

	// The client calculates the chunk count by itself: You could in theory send a chunk count of 0 even
	// though there's data, and the client will still download normally.
	chunkCount := int32(pk.Size / uint64(pk.DataChunkSize))
	if pk.Size%uint64(pk.DataChunkSize) != 0 {
		chunkCount++
	}

	idCopy := pk.UUID
	go func() {
		for i := int32(0); i < chunkCount; i++ {
			_ = conn.WritePacket(&packet.ResourcePackChunkRequest{
				UUID:       idCopy,
				ChunkIndex: i,
			})
			select {
			case <-conn.ctx.Done():
				return
			case frag := <-pack.newFrag:
				// Write the fragment to the full buffer of the downloading resource pack.
				_, _ = pack.buf.Write(frag)
			}
		}
		conn.packMu.Lock()
		if pack.buf.Len() != int(pack.size) {
			conn.log.Error(fmt.Sprintf("download resource pack: incorrect resource pack size: expected %v, got %v", pack.size, pack.buf.Len()), "UUID", id)
			conn.packMu.Unlock()
			return
		}
		// First parse the resource pack from the total byte buffer we obtained.
		newPack, err := resource.Read(pack.buf)
		if err != nil {
			conn.log.Error("download resource pack: invalid full resource pack data: "+err.Error(), "UUID", id)
			conn.packMu.Unlock()
			return
		}
		newPack = newPack.WithContentKey(pack.contentKey)
		conn.packQueue.packAmount--
		// Finally we add the resource to the resource packs slice.
		conn.resourcePacks = append(conn.resourcePacks, newPack)
		packAmount := conn.packQueue.packAmount
		conn.packMu.Unlock()

		if packAmount == 0 {
			conn.expect(packet.IDResourcePackStack)
			_ = conn.WritePacket(&packet.ResourcePackClientResponse{Response: packet.PackResponseAllPacksDownloaded})
		}
		conn.storeResourcePack(pack.cacheKey, newPack)
	}()
	return nil
}

func (conn *Conn) HandlePackChunkData(pk *packet.ResourcePackChunkData) error {
	pk.UUID = strings.Split(pk.UUID, "_")[0]
	pack, ok := conn.packQueue.awaitingPacks[pk.UUID]
	if !ok {
		return fmt.Errorf("chunk data for resource pack that was not being downloaded")
	}
	lastData := pack.buf.Len()+int(pack.chunkSize) >= int(pack.size)
	if !lastData && uint32(len(pk.Data)) != pack.chunkSize {
		return fmt.Errorf("expected chunk size %v, got %v", pack.chunkSize, len(pk.Data))
	}
	if pk.ChunkIndex != pack.expectedIndex {
		return fmt.Errorf("expected chunk index %v, got %v", pack.expectedIndex, pk.ChunkIndex)
	}
	pack.expectedIndex++
	pack.newFrag <- pk.Data
	return nil
}

func (conn *Conn) HandlePackChunkRequest(pk *packet.ResourcePackChunkRequest) error {
	current := conn.packQueue.currentPack
	chunkSize := uint64(conn.packQueue.chunkSize)
	uuid, _, _ := strings.Cut(pk.UUID, "_")
	if current.UUID().String() != uuid {
		return fmt.Errorf("expected pack UUID %v, but got %v", current.UUID(), pk.UUID)
	}
	if conn.packQueue.currentOffset != uint64(pk.ChunkIndex)*chunkSize {
		return fmt.Errorf("expected chunk index %v, but got %v", conn.packQueue.currentOffset/chunkSize, pk.ChunkIndex)
	}
	response := &packet.ResourcePackChunkData{
		UUID:       pk.UUID,
		ChunkIndex: uint32(pk.ChunkIndex),
		DataOffset: conn.packQueue.currentOffset,
		Data:       make([]byte, chunkSize),
	}
	conn.packQueue.currentOffset += chunkSize
	// We read the data directly into the response's data.
	if n, err := current.ReadAt(response.Data, int64(response.DataOffset)); err != nil {
		// If we hit an EOF, we don't need to return an error, as we've simply reached the end of the content
		// AKA the last chunk.
		if err != io.EOF {
			return fmt.Errorf("read resource pack chunk: %w", err)
		}
		response.Data = response.Data[:n]
	}
	if err := conn.WritePacket(response); err != nil {
		return fmt.Errorf("send ResourcePackChunkData: %w", err)
	}
	if err := conn.Flush(); err != nil {
		return fmt.Errorf("flush ResourcePackChunkData: %w", err)
	}
	if err := waitResourcePackChunkSendDelay(conn.ctx, conn.resourcePackDelivery.ChunkSendDelay); err != nil {
		return err
	}

	lastChunk := uint64(pk.ChunkIndex)*chunkSize+chunkSize >= uint64(current.Len())
	if lastChunk {
		if !conn.packQueue.AllDownloaded() {
			_ = conn.nextResourcePackDownload()
			// We send the next pack's data info and continue expecting its chunk requests.
			conn.expect(packet.IDResourcePackChunkRequest)
		} else {
			conn.expect(packet.IDResourcePackClientResponse)
		}
	}
	return nil
}

// waitResourcePackChunkSendDelay waits for the configured delay before processing the next resource pack
// chunk request.
func waitResourcePackChunkSendDelay(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
