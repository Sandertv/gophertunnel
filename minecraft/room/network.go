package room

import (
	"context"
	"github.com/sandertv/gophertunnel/minecraft"
	"net"
)

type Network struct {
	Network minecraft.Network

	Announcer Announcer

	ListenConfig ListenConfig
}

func (n Network) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return n.Network.DialContext(ctx, address)
}

func (n Network) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return n.Network.PingContext(ctx, address)
}

func (n Network) Listen(address string) (minecraft.NetworkListener, error) {
	listener, err := n.Network.Listen(address)
	if err != nil {
		return nil, err
	}

	l, err := n.ListenConfig.Listen(n.Announcer, listener)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (n Network) Encrypted() bool {
	return n.Network.Encrypted()
}

func (n Network) BatchHeader() []byte {
	return n.Network.BatchHeader()
}
