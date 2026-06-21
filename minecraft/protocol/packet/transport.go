package packet

// BatchHeaderer may be implemented by transports that need a custom batch prefix.
//
// The default Minecraft packet transport uses the standard batch header. Transports
// such as NetherNet can return nil when packets are already framed by the transport.
type BatchHeaderer interface {
	// BatchHeader returns the bytes expected before each packet batch.
	BatchHeader() []byte
}

// EncryptionDisabler may be implemented by transports that must not use
// Minecraft packet encryption.
//
// This is intended for transports that already provide their own encryption.
type EncryptionDisabler interface {
	// DisableEncryption reports whether Minecraft packet encryption should be disabled.
	DisableEncryption() bool
}

// PacketReader may be implemented by transports that can read complete packet
// payloads directly.
type PacketReader interface {
	// ReadPacket reads one complete packet payload from the transport.
	ReadPacket() ([]byte, error)
}

// TransportCapabilities groups optional packet-layer methods that transport
// wrappers must preserve.
//
// Implementing this full interface is not required for ordinary transports, but
// wrappers around a transport that does implement these methods must keep them
// visible or packet framing and encryption behaviour may change.
type TransportCapabilities interface {
	BatchHeaderer
	EncryptionDisabler
	PacketReader
}
