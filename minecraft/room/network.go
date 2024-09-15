package room

import (
	"context"
	"github.com/sandertv/gophertunnel/minecraft"
	"net"
)

// Network wraps the [minecraft.Network], extending its functionality by hijacking some methods
// to add some features that room package would provide. It provides the ability to customize
// listener through [ListenConfig].
//
// It must be registered manually using [minecraft.RegisterNetwork] with an ID.
type Network struct {
	// Network is the underlying [minecraft.Network] that will be extended.
	Network minecraft.Network

	// ListenConfig specifies the configuration used to customize listeners.
	ListenConfig ListenConfig
}

// DialContext ...
func (n Network) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return n.Network.DialContext(ctx, address)
}

// PingContext ...
func (n Network) PingContext(ctx context.Context, address string) ([]byte, error) {
	return n.Network.PingContext(ctx, address)
}

// Listen listens on the specified address using the underlying [minecraft.Network.Listen] method
// and wraps the returned [minecraft.NetworkListener] to extend its functionality with [ListenConfig.Wrap].
// This wrapping allows for additioanl features, such as hijacking the [minecraft.ServerStatus] for reporting
// it with Status on Announcer.
func (n Network) Listen(address string) (minecraft.NetworkListener, error) {
	l, err := n.Network.Listen(address)
	if err != nil {
		return nil, err
	}
	return n.ListenConfig.Wrap(l), nil
}

// DisableEncryption ...
func (n Network) DisableEncryption() bool { return n.Network.DisableEncryption() }

// BatchHeader ...
func (n Network) BatchHeader() []byte { return n.Network.BatchHeader() }
