package nethernet

import (
	"context"
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"net"
	"strconv"
)

type Network struct {
	Signaling Signaling
}

func (n Network) DialContext(ctx context.Context, address string) (net.Conn, error) {
	if n.Signaling == nil {
		return nil, errors.New("minecraft/nethernet: Network.DialContext: Signaling is nil")
	}
	networkID, err := strconv.ParseUint(address, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse network ID: %w", err)
	}
	var d Dialer
	return d.DialContext(ctx, networkID, n.Signaling)
}

func (n Network) PingContext(context.Context, string) ([]byte, error) {
	return nil, errors.New("minecraft/nethernet: Network.PingContext: not supported")
}

func (n Network) Listen(address string) (minecraft.NetworkListener, error) {
	if n.Signaling == nil {
		return nil, errors.New("minecraft/nethernet: Network.Listen: Signaling is nil")
	}
	networkID, err := strconv.ParseUint(address, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse network ID: %w", err)
	}
	var cfg ListenConfig
	return cfg.Listen(networkID, n.Signaling)
}

func (Network) Encrypted() bool { return true }

func (Network) BatchHeader() []byte { return nil }

func NetworkAddress(networkID uint64) string {
	return strconv.FormatUint(networkID, 10)
}
