package minecraft

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/df-mc/go-nethernet"
)

// NetherNet is an implementation of a NetherNet network.
type NetherNet struct {
	Signaling nethernet.Signaling
	Log       *slog.Logger
}

// DialContext ...
func (n NetherNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return nethernet.Dialer{Log: n.Log}.DialContext(ctx, address, n.Signaling)
}

// PingContext ...
func (n NetherNet) PingContext(context.Context, string) ([]byte, error) {
	return nil, errors.New("minecraft: NetherNet.PingContext: not supported")
}

// Listen ...
func (n NetherNet) Listen(string) (NetworkListener, error) {
	return nethernet.ListenConfig{Log: n.Log}.Listen(n.Signaling)
}
