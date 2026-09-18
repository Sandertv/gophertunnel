package minecraft

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-nethernet/endpoint"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NetherNet is an implementation of a NetherNet network, a WebRTC-based transport layer protocol.
// A valid Signaling implementation must be provided before use.
type NetherNet struct {
	// Signaling is the interface used to exchange connection details with the remote peers.
	// When set to nil, either [endpoint.Client] or [endpoint.Handler] will be created and used.
	// If [endpoint.Client] is specified, the network cannot listen on an address.
	// If [endpoint.Handler] is specified, the network cannot dial or ping a remote server.
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

// DialContext establishes a connection with the remote NetherNet network using the address.
func (n NetherNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	if n.Signaling == nil {
		n.Signaling = endpoint.NewClient()
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
		env, err := authEnv(ctx)
		if err != nil {
			return nil, fmt.Errorf("request authorization environment: %w", err)
		}
		n.Dialer.Identity = &nethernet.Identity{
			PrivateKey: privateKey,
			Token:      token,
			// We need to append '/' on the URL if not present.
			Domain: env.Issuer.JoinPath().String(),
		}
	}
	return n.DialContext(ctx, address)
}

// PingContext sends a ping to the specified address if the [NetherNet.Signaling]
// supports sending a ping request to the remote server. Otherwise, it returns an error.
// If [NetherNet.Signaling] is nil, [endpoint.Client] will be created and used instead.
// The address must be an HTTP/HTTPS URL with a port.
func (n NetherNet) PingContext(ctx context.Context, address string) ([]byte, error) {
	if n.Signaling == nil {
		n.Signaling = endpoint.NewClient()
	}
	if p, ok := n.Signaling.(pinger); ok {
		return p.PingContext(ctx, address)
	}
	return nil, errors.New("minecraft: NetherNet.PingContext: not supported")
}

// Ensure that the default Signaling Client implementation used by NetherNet satisfies
// the pinger interface.
var _ pinger = (*endpoint.Client)(nil)

// pinger is implemented by [nethernet.Signaling] that support sending
// ping requests to remote servers.
type pinger interface {
	PingContext(ctx context.Context, address string) ([]byte, error)
}

// Listen creates a NetherNet listener using the [NetherNet.Signaling] implementation.
// If [NetherNet.Signaling] is nil, an HTTP server will be started at the specified address.
// Unlike [PingContext] and [DialContext], the address should be formatted as 'ip:port'.
func (n NetherNet) Listen(address string) (_ NetworkListener, err error) {
	var closeFunc func() error
	if n.Signaling == nil {
		server, err := endpoint.Serve(address)
		if err != nil {
			return nil, fmt.Errorf("serve HTTP: %w", err)
		}
		n.Signaling = server
		closeFunc = server.Close
	}
	if n.ListenConfig.Log == nil && n.Log != nil {
		n.ListenConfig.Log = n.Log
	}
	l, err := n.ListenConfig.Listen(n.Signaling)
	if err != nil {
		if closeFunc != nil {
			if err2 := closeFunc(); err2 != nil {
				err = errors.Join(err, fmt.Errorf("close HTTP server: %w", err2))
			}
		}
		return nil, err
	}
	return netherNetListener{NetworkListener: l, close: closeFunc}, nil
}

// netherNetListener implements [NetworkListener] with additional handling
// to close the HTTP server together with the underlying NetworkListener.
type netherNetListener struct {
	NetworkListener
	// close is a function for closing the HTTP server used to listen.
	// It is only set if the caller did not specify [NetherNet.Signaling]
	// when calling [NetherNet.Listen].
	close func() error
}

// Close closes the underlying NetworkListener along with the HTTP server.
func (n netherNetListener) Close() (err error) {
	if err2 := n.NetworkListener.Close(); err2 != nil {
		err = errors.Join(err, err2)
	}
	if n.close != nil {
		if err2 := n.close(); err2 != nil {
			err = errors.Join(err, fmt.Errorf("close HTTP server: %w", err2))
		}
	}
	return err
}

// init registers the NetherNet network.
func init() {
	RegisterNetwork("nethernet", func(l *slog.Logger) Network {
		return NetherNet{Log: l}
	})
}
