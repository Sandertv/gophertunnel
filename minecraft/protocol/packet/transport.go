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

// TransportCapabilities is the full set of optional packet transport methods.
//
// Normal transports do not need to implement this interface. If code wraps a
// connection that has any of these methods, the wrapper must expose the same
// methods too. Otherwise packets may be framed, encrypted, or read differently.
type TransportCapabilities interface {
	BatchHeaderer
	EncryptionDisabler
	PacketReader
}
