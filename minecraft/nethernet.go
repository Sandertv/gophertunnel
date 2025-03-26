package minecraft

import (
	"context"
	"fmt"
	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-nethernet/discovery"
	"github.com/sandertv/go-raknet"
	"log/slog"
	"math/rand/v2"
	"net"
)

// NetherNet is an implementation of a NetherNet Network.
type NetherNet struct {
	l *slog.Logger
}

// DialContext ...
func (r NetherNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return raknet.Dialer{ErrorLog: r.l.With("net origin", "nethernet")}.DialContext(ctx, address)
}

// PingContext ...
func (r NetherNet) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return raknet.Dialer{ErrorLog: r.l.With("net origin", "nethernet")}.PingContext(ctx, address)
}

// Listen ...
func (r NetherNet) Listen(address string) (NetworkListener, error) {
	log := r.l.With("net origin", "nethernet")
	d, err := discovery.ListenConfig{
		Log:       log,
		NetworkID: rand.Uint64(),
	}.Listen("udp", address)
	if err != nil {
		return nil, fmt.Errorf("error listening on discovery: %w", err)
	}
	l, err := nethernet.ListenConfig{Log: log}.Listen(d)
	if err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("error listening: %w", err)
	}
	return l, nil
}

// init registers the RakNet network.
func init() {
	RegisterNetwork("nethernet", func(l *slog.Logger) Network { return NetherNet{l: l} })
}
