package minecraft

import (
	"fmt"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync/atomic"
)

// Listener implements a Minecraft listener on top of an unspecific net.Listener. It abstracts away the
// login sequence of connecting clients and provides the implements the net.Listener interface to provide a
// consistent API.
type Listener struct {
	// ErrorLog is a log.Logger that errors that occur during packet handling of clients are written to. By
	// default, ErrorLog is set to one equal to the global logger.
	ErrorLog *log.Logger

	// AuthenticationDisables specifies if authentication of players that join is disabled. If set to true, no
	// verification will be done to ensure that the player connecting is authenticated using their XBOX Live
	// account.
	AuthenticationDisabled bool

	// ServerName is the server name shown in the in-game menu, above the player list. The name cannot be
	// changed after a player is connected. By default, 'Minecraft Server' will be set.
	ServerName string
	// MaximumPlayers is the maximum amount of players accepted in the server. If non-zero, players that
	// attempt to join while the server is full will be kicked during login. If zero, the maximum player count
	// will be dynamically updated each time a player joins, so that an unlimited amount of players is
	// accepted into the server.
	MaximumPlayers int

	// ResourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	ResourcePacks []*resource.Pack
	// TexturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	TexturePacksRequired bool

	// PacketFunc is called whenever a packet is read from or written to a connection returned when using
	// Listener.Accept. It includes packets that are otherwise covered in the connection sequence, such as the
	// Login packet. The function is called with the header of the packet and its raw payload, the address
	// from which the packet originated, and the destination address.
	PacketFunc func(header packet.Header, payload []byte, src, dst net.Addr)

	// playerCount is the amount of players connected to the server. If MaximumPlayers is non-zero and equal
	// to the playerCount, no more players will be accepted.
	playerCount *int32

	listener net.Listener

	hijackingPong atomic.Value
	incoming      chan *Conn
}

// Listen announces on the local network address. The network is typically "raknet".
// If the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all
// available unicast and anycast IP addresses of the local system.
func (listener *Listener) Listen(network, address string) error {
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
		return err
	}
	var count int32

	if listener.ErrorLog == nil {
		listener.ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
	}
	if listener.ServerName == "" {
		listener.ServerName = "Minecraft Server"
	}
	listener.listener = netListener
	listener.incoming = make(chan *Conn)
	listener.hijackingPong.Store(false)
	listener.playerCount = &count

	// Actually start listening.
	go listener.listen()
	return nil
}

// Listen announces on the local network address. The network must be "tcp", "tcp4", "tcp6", "unix",
// "unixpacket" or "raknet". A Listener is returned which may be used to accept connections.
// If the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all
// available unicast and anycast IP addresses of the local system.
// Listen has the default values for the fields of Listener filled out. To use different values for these
// fields, call &Listener{}.Listen() instead.
func Listen(network, address string) (*Listener, error) {
	l := &Listener{}
	err := l.Listen(network, address)
	return l, err
}

// Accept accepts a fully connected (on Minecraft layer) connection which is ready to receive and send
// packets. It is recommended to cast the net.Conn returned to a *minecraft.Conn so that it is possible to
// use the conn.ReadPacket() and conn.WritePacket() methods.
// Accept returns an error if the listener is closed.
func (listener *Listener) Accept() (net.Conn, error) {
	conn, ok := <-listener.incoming
	if !ok {
		return nil, fmt.Errorf("accept: listener closed")
	}
	return conn, nil
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
	listener.hijackingPong.Store(true)
	return listener.listener.(*raknet.Listener).HijackPong(address)
}

// Addr returns the address of the underlying listener.
func (listener *Listener) Addr() net.Addr {
	return listener.listener.Addr()
}

// Close closes the listener and the underlying net.Listener. Pending calls to Accept will fail immediately.
func (listener *Listener) Close() error {
	close(listener.incoming)
	return listener.listener.Close()
}

// updatePongData updates the pong data of the listener using the current only players, maximum players and
// server name of the listener, provided the listener isn't currently hijacking the pong of another server.
func (listener *Listener) updatePongData() {
	if listener.hijackingPong.Load().(bool) {
		return
	}
	maxCount := int32(listener.MaximumPlayers)
	current := atomic.LoadInt32(listener.playerCount)
	if maxCount == 0 {
		// If the maximum amount of allowed players is 0, we set it to the the current amount of line players
		// plus 1, so that new players can always join.
		maxCount = current + 1
	}

	rakListener := listener.listener.(*raknet.Listener)
	rakListener.PongData([]byte(fmt.Sprintf("MCPE;%v;%v;%v;%v;%v;%v;Minecraft Server;0",
		listener.ServerName, protocol.CurrentProtocol, protocol.CurrentVersion, current, maxCount, rakListener.ID(),
	)))
}

// listen starts listening for incoming connections and packets. When a player is fully connected, it submits
// it to the accepted connections channel so that a call to Accept can pick it up.
func (listener *Listener) listen() {
	listener.updatePongData()
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
		listener.createConn(netConn)
	}
}

// createConn creates a connection for the net.Conn passed and adds it to the listener, so that it may be
// accepted once its login sequence is complete.
func (listener *Listener) createConn(netConn net.Conn) {
	conn := newConn(netConn, nil, listener.ErrorLog)
	conn.packetFunc = listener.PacketFunc
	conn.texturePacksRequired = listener.TexturePacksRequired
	conn.resourcePacks = listener.ResourcePacks
	conn.gameData.WorldName = listener.ServerName
	conn.authEnabled = !listener.AuthenticationDisabled

	if atomic.LoadInt32(listener.playerCount) == int32(listener.MaximumPlayers) && listener.MaximumPlayers != 0 {
		// The server was full. We kick the player immediately and close the connection.
		_ = conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusLoginFailedServerFull})
		_ = conn.Close()
		return
	}
	atomic.AddInt32(listener.playerCount, 1)
	listener.updatePongData()

	go listener.handleConn(conn)
}

// handleConn handles an incoming connection of the Listener. It will first attempt to get the connection to
// log in, after which it will expose packets received to the user.
func (listener *Listener) handleConn(conn *Conn) {
	defer func() {
		_ = conn.Close()
		atomic.AddInt32(listener.playerCount, -1)
		listener.updatePongData()
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
}
