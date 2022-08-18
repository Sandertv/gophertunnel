package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// NetworkSettings is sent by the server to update a variety of network settings. These settings modify the
// way packets are sent over the network stack.
type NetworkSettings struct {
	// CompressionThreshold is the minimum size of a packet that is compressed when sent. If the size of a
	// packet is under this value, it is not compressed.
	// When set to 0, all packets will be left uncompressed.
	CompressionThreshold uint16
	// CompressionAlgorithm is the algorithm that is used to compress packets.
	CompressionAlgorithm Compression
}

// ID ...
func (*NetworkSettings) ID() uint32 {
	return IDNetworkSettings
}

// Marshal ...
func (pk *NetworkSettings) Marshal(w *protocol.Writer) {
	id := pk.CompressionAlgorithm.EncodeCompression()
	w.Uint16(&pk.CompressionThreshold)
	w.Uint16(&id)
}

// Unmarshal ...
func (pk *NetworkSettings) Unmarshal(r *protocol.Reader) {
	r.Uint16(&pk.CompressionThreshold)

	var id uint16
	r.Uint16(&id)

	pk.CompressionAlgorithm, _ = CompressionByID(id)
}
