package minecraft

import "time"

const (
	// DefaultResourcePackChunkSize is the size of a single chunk of data from a resource pack sent by a
	// Listener: 128 KiB.
	DefaultResourcePackChunkSize = 1024 * 128
	// DefaultResourcePackChunkSendDelay is the delay a Listener leaves between ResourcePackChunkData
	// packets, so slow clients are not flooded while downloading packs. Clients after 1.26.30 may fail
	// resource pack downloads when pack chunks are sent too aggressively.
	DefaultResourcePackChunkSendDelay = 200 * time.Millisecond
)

// ResourcePackDeliveryConfig controls how a Listener sends resource pack data to clients. The zero value
// keeps the conservative defaults.
type ResourcePackDeliveryConfig struct {
	// ChunkSize is the size of each ResourcePackChunkData payload. Zero uses DefaultResourcePackChunkSize.
	ChunkSize uint32
	// ChunkSendDelay is the delay after each ResourcePackChunkData packet. Zero uses
	// DefaultResourcePackChunkSendDelay; a negative delay disables pacing.
	ChunkSendDelay time.Duration
}

// defaultResourcePackDeliveryConfig returns the default resource pack delivery configuration.
func defaultResourcePackDeliveryConfig() ResourcePackDeliveryConfig {
	return ResourcePackDeliveryConfig{
		ChunkSize:      DefaultResourcePackChunkSize,
		ChunkSendDelay: DefaultResourcePackChunkSendDelay,
	}
}

// normalized returns the configuration with defaults filled in.
func (config ResourcePackDeliveryConfig) normalized() ResourcePackDeliveryConfig {
	defaults := defaultResourcePackDeliveryConfig()
	if config.ChunkSize == 0 {
		config.ChunkSize = defaults.ChunkSize
	}
	if config.ChunkSendDelay == 0 {
		config.ChunkSendDelay = defaults.ChunkSendDelay
	} else if config.ChunkSendDelay < 0 {
		config.ChunkSendDelay = 0
	}
	return config
}
