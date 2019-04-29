package minecraft

import (
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"log"
	"net"
	"os"
)

// Listener implements a Minecraft listener on top of an unspecific net.Listener. It abstracts away the
// login sequence of connecting clients and provides the implements the net.Listener interface to provide a
// consistent API.
type Listener struct {
	// ErrorLog is a log.Logger that errors that occur during packet handling of clients are written to. By
	// default, ErrorLog is set to one equal to the global logger.
	ErrorLog *log.Logger

	// resourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	resourcePacks []*resource.Pack
	// texturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	texturePacksRequired bool

	listener net.Listener

	incoming chan *Conn
	close    chan bool
}

// Listen announces on the local network address. The network must be "tcp", "tcp4", "tcp6", "unix",
// "unixpacket" or "raknet". A Listener is returned which may be used to accept connections.
// If the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all
// available unicast and anycast IP addresses of the local system.
func Listen(network, address string) (*Listener, error) {
	var listener net.Listener
	var err error
	switch network {
	case "raknet":
		// Listen specifically for the RakNet network type, as the standard library (obviously) doesn't
		// implement that.
		listener, err = raknet.Listen(address)
		if err != nil {
			return nil, err
		}
	default:
		// Otherwise fall back to the standard net.Listen.
		listener, err = net.Listen(network, address)
		if err != nil {
			return nil, err
		}
	}

	mcListener := &Listener{
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
		listener: listener,
		close:    make(chan bool, 2),
		incoming: make(chan *Conn),
	}

	// Actually start listening.
	go mcListener.listen()

	return mcListener, nil
}

// Accept accepts a fully connected (on Minecraft layer) connection which is ready to receive and send
// packets. It is recommended to cast the net.Conn returned to a *minecraft.Conn so that it is possible to
// use the conn.ReadPacket() and conn.WritePacket() methods.
func (listener *Listener) Accept() (net.Conn, error) {
	return <-listener.incoming, nil
}

// ResourcePacks sets the resource packs settings of the listener. If texturePacksRequired is set to true,
// clients must accept all resource packs in order to be able to join the server.
// A list of resource packs may be supplied to the method, which the client will have to download when it
// tries to join. These resource packs may be both texture and behaviour packs.
func (listener *Listener) ResourcePacks(texturePacksRequired bool, packs ...*resource.Pack) {
	listener.texturePacksRequired = texturePacksRequired
	listener.resourcePacks = packs
}

// Addr returns the address of the underlying listener.
func (listener *Listener) Addr() net.Addr {
	return listener.listener.Addr()
}

// Close closes the listener and the underlying net.Listener.
func (listener *Listener) Close() error {
	listener.close <- true
	return listener.listener.Close()
}

// listen starts listening for incoming connections and packets. When a player is fully connected, it submits
// it to the accepted connections channel so that a call to Accept can pick it up.
func (listener *Listener) listen() {
	defer func() {
		_ = listener.Close()
	}()
	for {
		netConn, err := listener.listener.Accept()
		if err != nil {
			// The underlying listener was closed, meaning we should return immediately so this listener can
			// close too.
			return
		}
		conn := newConn(netConn)
		conn.texturePacksRequired = listener.texturePacksRequired
		conn.resourcePacks = listener.resourcePacks

		go func() {
			defer func() {
				_ = conn.Close()
			}()
			for {
				// We finally arrived at the packet decoding loop. We constantly decode packets that arrive
				// and push them to the Conn so that they may be processed.
				packets, err := conn.decoder.Decode()
				if err != nil {
					if !raknet.ErrConnectionClosed(err) {
						listener.ErrorLog.Printf("error reading from client connection: %v\n", err)
					}
					return
				}
				for _, data := range packets {
					loggedInBefore := conn.loggedIn
					if err := conn.handleIncoming(data); err != nil {
						listener.ErrorLog.Printf("%v", err)
						return
					}
					if !loggedInBefore && conn.loggedIn {
						// The connection was previously not logged in, but was after receiving this packet,
						// meaning the connection is fully completely now. We add it to the channel so that
						// a call to Accept() can receive it.
						listener.incoming <- conn
					}
				}
			}
		}()
	}
}
