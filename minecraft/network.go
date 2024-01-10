package minecraft

import (
	"context"
	"net"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Network represents an implementation of a supported network layers, such as RakNet.
type Network interface {
	// DialContext attempts to dial a connection to the address passed. The address may be either an IP address or a
	// hostname, combined with a port that is separated with ':'.
	// DialContext will use the deadline (ctx.Deadline) of the context.Context passed for the maximum amount of time that
	// the dialing can take. DialContext will terminate as soon as possible when the context.Context is closed.
	DialContext(ctx context.Context, address string) (net.Conn, error)
	// PingContext sends a ping to an address and returns the response obtained. If successful, a non-nil response byte
	// slice containing the data is returned. If the ping failed, an error is returned describing the failure.
	// Note that the packet sent to the server may be lost due to the nature of UDP. If this is the case, PingContext
	// could last indefinitely, hence a timeout should always be attached to the context passed.
	// PingContext cancels as soon as the deadline expires.
	PingContext(ctx context.Context, address string) (response []byte, err error)
	// Listen listens on the address passed and returns a listener that may be used to accept connections. If not
	// successful, an error is returned.
	// The address follows the same rules as those defined in the net.TCPListen() function.
	// Specific features of the listener may be modified once it is returned, such as the used log and/or the
	// accepted protocol.
	Listen(address string) (NetworkListener, error)

	// Compression returns a new compression instance used by this Protocol.
	Compression(conn net.Conn) packet.Compression
}

// NetworkListener represents a listening connection to a remote server. It is the equivalent of net.Listener, but with extra
// functionality specific to Minecraft.
type NetworkListener interface {
	net.Listener
	// ID returns the unique ID of the listener. This ID is usually used by a client to identify a specific
	// server during a single session.
	ID() int64
	// PongData sets the pong data that is used to respond with when a client sends a ping. It usually holds game
	// specific data that is used to display in a server list.
	// If a data slice is set with a size bigger than math.MaxInt16, the function panics.
	PongData(data []byte)
}

// networks holds a map of id => Network to be used for looking up the network by an ID. It is registered to when calling
// RegisterNetwork.
var networks = map[string]Network{}

// RegisterNetwork registers a network so that it can be used for Gophertunnel.
func RegisterNetwork(id string, n Network) {
	networks[id] = n
}

// networkByID returns the network with the ID passed. If no network is found, the second return value will be false.
func networkByID(id string) (Network, bool) {
	n, ok := networks[id]
	return n, ok
}
