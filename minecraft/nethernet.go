package minecraft

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"log/slog"
	"net"

	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NetherNet is an implementation of a NetherNet network, a WebRTC-based transport layer protocol.
// A valid Signaling implementation must be provided before use.
type NetherNet struct {
	// Signaling is the interface used to exchange connection details with the remote peers.
	Signaling nethernet.Signaling

	// Dialer specifies options for establishing a connection with DialContext.
	Dialer nethernet.Dialer
	// ListenConfig specifies options for listening for connections with Listen.
	ListenConfig nethernet.ListenConfig
	// Log is the logger used by default for Dialer and ListenConfig.
	// It is useful when registering this network from RegisterNetwork.
	Log *slog.Logger
}

// Ensure the connection returned by NetherNet.DialContext has the optional
// packet methods used by Encoder and Decoder, even though DialContext returns it
// as a net.Conn.
var _ packet.TransportCapabilities = (*nethernet.Conn)(nil)

// DialContext ...
func (n NetherNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	if n.Signaling == nil {
		return nil, errors.New("minecraft: NetherNet.DialContext: Signaling is nil")
	}
	if n.Dialer.Log == nil && n.Log != nil {
		n.Dialer.Log = n.Log
	}
	return n.Dialer.DialContext(ctx, address, n.Signaling)
}

// DialContextIdentity establishes a connection with the remote NetherNet peer using the
// JWT token issued by the auth service and the private key bound to the Minecraft connection.
// If [nethernet.Dialer.Identity] is already set, that identity is used instead.
func (n NetherNet) DialContextIdentity(ctx context.Context, address string, token string, privateKey *ecdsa.PrivateKey) (net.Conn, error) {
	if n.Dialer.Identity == nil {
		n.Dialer.Identity = &nethernet.Identity{
			PrivateKey: privateKey,
			Token:      token,
		}
	}
	return n.DialContext(ctx, address)
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
	if n.ListenConfig.Log == nil && n.Log != nil {
		n.ListenConfig.Log = n.Log
	}
	return n.ListenConfig.Listen(n.Signaling)
}
