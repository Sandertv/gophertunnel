package minecraft

// DefaultResourcePackMaxInFlightChunks matches the vanilla client request window.
const DefaultResourcePackMaxInFlightChunks = 100

// maxResourcePackPrealloc bounds the buffer reserved up front for a pack. The size comes from
// ResourcePacksInfo and is only a claim at that point, so a server would otherwise turn one small packet
// into an allocation of any size it names. Larger packs still grow their buffer as chunks arrive.
const maxResourcePackPrealloc = 32 << 20

// ResourcePackDownloadConfig controls resource pack downloads performed by a Dialer.
type ResourcePackDownloadConfig struct {
	// MaxInFlightChunks is the maximum number of outstanding chunk requests. Values below one use the
	// vanilla default.
	MaxInFlightChunks int
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

// normalized returns the configuration with defaults filled in.
func (config ResourcePackDownloadConfig) normalized() ResourcePackDownloadConfig {
	if config.MaxInFlightChunks < 1 {
		config.MaxInFlightChunks = DefaultResourcePackMaxInFlightChunks
	}
	return config
}
