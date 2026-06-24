package minecraft

import (
	"context"
	"log/slog"
	"net"

	"github.com/sandertv/go-raknet"
)

// RakNet is an implementation of a RakNet v10 Network.
type RakNet struct {
	l *slog.Logger
	// Logger overrides the logger used for RakNet dial and listen errors.
	// If nil, the logger passed by RegisterNetwork is used.
	Logger *slog.Logger
	// UpstreamDialer overrides the dialer used for outbound UDP connections.
	// If nil, RakNet uses the default net.Dialer.
	UpstreamDialer raknet.UpstreamDialer
	// ServerID overrides the RakNet server GUID advertised by listeners.
	// If zero, RakNet generates a unique ID for each listener.
	ServerID int64
}

// DialContext ...
func (r RakNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return r.dialer().DialContext(ctx, address)
}

// PingContext ...
func (r RakNet) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return r.dialer().PingContext(ctx, address)
}

// Listen ...
func (r RakNet) Listen(address string) (NetworkListener, error) {
	return raknet.ListenConfig{
		ErrorLog: r.logger().With("net origin", "raknet"),
		ServerID: r.ServerID,
	}.Listen(address)
}

func (r RakNet) dialer() raknet.Dialer {
	return raknet.Dialer{
		ErrorLog:       r.logger().With("net origin", "raknet"),
		UpstreamDialer: r.UpstreamDialer,
	}
}

func (r RakNet) logger() *slog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	if r.l != nil {
		return r.l
	}
	return slog.Default()
}

// init registers the RakNet network.
func init() {
	RegisterNetwork("raknet", func(l *slog.Logger) Network { return RakNet{l: l} })
}
