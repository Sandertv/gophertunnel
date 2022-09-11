package minecraft

import (
	"context"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
)

// RakNet is an implementation of a RakNet v10 Network.
type RakNet struct{}

func (RakNet) DialContext(ctx context.Context, a string) (net.Conn, error) {
	return raknet.DialContext(ctx, a)
}
func (RakNet) PingContext(ctx context.Context, a string) ([]byte, error) {
	return raknet.PingContext(ctx, a)
}
func (RakNet) Listen(address string) (NetworkListener, error) { return raknet.Listen(address) }

func (RakNet) Compression(net.Conn) packet.Compression { return packet.FlateCompression{} }

// init registers the RakNet network.
func init() {
	RegisterNetwork("raknet", RakNet{})
}
