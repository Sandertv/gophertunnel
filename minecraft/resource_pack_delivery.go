package minecraft

const (
	// DefaultResourcePackChunkSize is the size of a single chunk of data from a resource pack sent by a
	// Listener: 100 KiB, the size a vanilla host serves packs at.
	DefaultResourcePackChunkSize = 1024 * 100
	// MaxResourcePackChunkSize is the largest ResourcePackChunkData a vanilla client accepts. A bigger
	// chunk is rejected as a malformed packet, so a Listener must stay under it.
	MaxResourcePackChunkSize = 10 * 1024 * 1024
)

// ResourcePackDeliveryConfig controls how a Listener sends resource pack data to clients. The zero value
// uses the default chunk size.
type ResourcePackDeliveryConfig struct {
	// ChunkSize is the size of each ResourcePackChunkData payload. Zero uses DefaultResourcePackChunkSize,
	// and anything above MaxResourcePackChunkSize is lowered to it.
	ChunkSize uint32
}

// defaultResourcePackDeliveryConfig returns the default resource pack delivery configuration.
func defaultResourcePackDeliveryConfig() ResourcePackDeliveryConfig {
	return ResourcePackDeliveryConfig{ChunkSize: DefaultResourcePackChunkSize}
}

// normalized returns the configuration with defaults filled in.
func (config ResourcePackDeliveryConfig) normalized() ResourcePackDeliveryConfig {
	if config.ChunkSize == 0 {
		config.ChunkSize = defaultResourcePackDeliveryConfig().ChunkSize
	}
	if config.ChunkSize > MaxResourcePackChunkSize {
		config.ChunkSize = MaxResourcePackChunkSize
	}
	return config
}
