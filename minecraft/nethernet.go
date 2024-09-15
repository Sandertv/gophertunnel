package minecraft

import (
	"context"
	"errors"
	"fmt"
	"github.com/df-mc/go-nethernet"
	"net"
	"strconv"
)

// NetherNet is an implementation of NetherNet network. Unlike RakNet, it needs to be registered manually with a Signaling.
type NetherNet struct {
	Signaling nethernet.Signaling
}

// DialContext ...
func (n NetherNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	if n.Signaling == nil {
		return nil, errors.New("minecraft: NetherNet.DialContext: Signaling is nil")
	}
	networkID, err := strconv.ParseUint(address, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse network ID: %w", err)
	}
	var d nethernet.Dialer
	return d.DialContext(ctx, networkID, n.Signaling)
}

// PingContext ...
func (n NetherNet) PingContext(context.Context, string) ([]byte, error) {
	return nil, errors.New("minecraft: NetherNet.PingContext: not supported")
}

// Listen ...
func (n NetherNet) Listen(address string) (NetworkListener, error) {
	if n.Signaling == nil {
		return nil, errors.New("minecraft: NetherNet.Listen: Signaling is nil")
	}
	networkID, err := strconv.ParseUint(address, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse network ID: %w", err)
	}
	var cfg nethernet.ListenConfig
	return cfg.Listen(networkID, n.Signaling)
}

// DisableEncryption ...
func (NetherNet) DisableEncryption() bool { return true }

// BatchHeader ...
func (NetherNet) BatchHeader() []byte { return nil }
