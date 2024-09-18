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

	// Dialer specifies options for establishing a connection with DialContext.
	Dialer nethernet.Dialer
	// ListenConfig specifies options for listening for connections with Listen.
	ListenConfig nethernet.ListenConfig
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
	return n.Dialer.DialContext(ctx, networkID, n.Signaling)
}

// PingContext ...
func (n NetherNet) PingContext(context.Context, string) ([]byte, error) {
	return nil, errors.New("minecraft: NetherNet.PingContext: not supported")
}

// Listen ...
func (n NetherNet) Listen(string) (NetworkListener, error) {
	if n.Signaling == nil {
		return nil, errors.New("minecraft: NetherNet.Listen: Signaling is nil")
	}
	return n.ListenConfig.Listen(n.Signaling)
}

// DisableEncryption ...
func (NetherNet) DisableEncryption() bool { return true }

// BatchHeader ...
func (NetherNet) BatchHeader() []byte { return nil }
