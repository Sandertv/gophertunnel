package minecraft

import (
	"context"
	"github.com/sandertv/go-raknet"
	"net"
)

// RakNet is an implementation of a RakNet v10 Network.
type RakNet struct{}

// DialContext ...
func (r RakNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return raknet.DialContext(ctx, address)
}

// PingContext ...
func (r RakNet) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return raknet.PingContext(ctx, address)
}

// Listen ...
func (r RakNet) Listen(address string) (NetworkListener, error) {
	return raknet.Listen(address)
}

// init registers the RakNet network.
func init() {
	RegisterNetwork("raknet", RakNet{})
}
