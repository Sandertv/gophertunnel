package minecraft

import (
	"fmt"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"io/ioutil"
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

	// ResourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	ResourcePacks []*resource.Pack
	// TexturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	TexturePacksRequired bool

	listener net.Listener

	incoming chan *Conn
	close    chan bool
}

// Listen announces on the local network address. The network must be "tcp", "tcp4", "tcp6", "unix",
// "unixpacket" or "raknet". A Listener is returned which may be used to accept connections.
// If the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all
// available unicast and anycast IP addresses of the local system.
func Listen(network, address string) (*Listener, error) {
	var netListener net.Listener
	var err error
	switch network {
	case "raknet":
		// Listen specifically for the RakNet network type, as the standard library (obviously) doesn't
		// implement that.
		var l *raknet.Listener
		l, err = raknet.Listen(address)
		if err == nil {
			l.ErrorLog = log.New(ioutil.Discard, "", 0)
			netListener = l
		}
	default:
		// Otherwise fall back to the standard net.Listen.
		netListener, err = net.Listen(network, address)
	}
	if err != nil {
		return nil, err
	}

	listener := &Listener{
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
		listener: netListener,
		close:    make(chan bool, 2),
		incoming: make(chan *Conn),
	}

	// Actually start listening.
	go listener.listen()
	return listener, nil
}

// Accept accepts a fully connected (on Minecraft layer) connection which is ready to receive and send
// packets. It is recommended to cast the net.Conn returned to a *minecraft.Conn so that it is possible to
// use the conn.ReadPacket() and conn.WritePacket() methods.
// Accept returns an error if the listener is closed.
func (listener *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-listener.incoming:
		return conn, nil
	case <-listener.close:
		listener.close <- true
		return nil, fmt.Errorf("accept: listener closed")
	}
}

// Disconnect disconnects a Minecraft Conn passed by first sending a disconnect with the message passed, and
// closing the connection after. If the message passed is empty, the client will be immediately sent to the
// player list instead of a disconnect screen.
func (listener *Listener) Disconnect(conn *Conn, message string) error {
	_ = conn.WritePacket(&packet.Disconnect{
		HideDisconnectionScreen: message == "",
		Message:                 message,
	})
	return conn.Close()
}

// HijackPong hijacks the pong response from a server at an address passed. The listener passed will
// continuously update its pong data by hijacking the pong data of the server at the address.
// The hijack will last until the listener is shut down.
// If the address passed could not be resolved, an error is returned.
// Calling HijackPong means that any current and future pong data set using listener.PongData is overwritten
// each update.
func (listener *Listener) HijackPong(address string) error {
	return listener.listener.(*raknet.Listener).HijackPong(address)
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
		conn := newConn(netConn, nil, listener.ErrorLog)
		conn.texturePacksRequired = listener.TexturePacksRequired
		conn.resourcePacks = listener.ResourcePacks

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
						listener.ErrorLog.Printf("error: %v", err)
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
