package minecraft

import (
	"context"
	"github.com/sandertv/go-raknet"
	"log/slog"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
)

// RakNet is an implementation of a RakNet v10 Network.
type RakNet struct{
	l *slog.Logger
}

// DialContext ...
func (RakNet) DialContext(ctx context.Context, a string) (net.Conn, error) {
	return raknet.DialContext(ctx, a)
}

// PingContext ...
func (RakNet) PingContext(ctx context.Context, a string) ([]byte, error) {
	return raknet.PingContext(ctx, a)
}

// Listen ...
func (RakNet) Listen(address string) (NetworkListener, error) { return raknet.Listen(address) }

// Compression ...
func (RakNet) Compression(net.Conn) packet.Compression { return packet.FlateCompression }

// init registers the RakNet network.
func init() {
	RegisterNetwork("raknet", func(l *slog.Logger) Network { return RakNet{l: l} })
}