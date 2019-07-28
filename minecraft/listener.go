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
)

// Listener implements a Minecraft listener on top of an unspecific net.Listener. It abstracts away the
// login sequence of connecting clients and provides the implements the net.Listener interface to provide a
// consistent API.
type Listener struct {
	// ErrorLog is a log.Logger that errors that occur during packet handling of clients are written to. By
	// default, ErrorLog is set to one equal to the global logger.
	ErrorLog *log.Logger

	// ServerName is the server name shown in the in-game menu, above the player list. The name cannot be
	// changed after a player is connected. By default, 'Minecraft Server' will be set.
	ServerName string
	// MaximumPlayers is the maximum amount of players accepted in the server. If non-zero, players that
	// attempt to join while the server is full will be kicked during login. If zero, the maximum player count
	// will be dynamically updated each time a player joins, so that an unlimited amount of players is
	// accepted into the server.
	MaximumPlayers int
	// InitialGameData is the game data that joining players should obtain. Some of this information may not
	// be changed over the course of the game, so they must be set correctly at the start.
	// A function must be set which returns a new object each time.
	InitialGameData func(conn *Conn) GameData

	// MaximumChunkRadius is the maximum initial requested chunk radius of the connection. The client will
	// obtain this initial chunk radius if it requests a radius higher than this.
	// A default maximum chunk radius of 16 is set.
	MaximumChunkRadius int

	// ResourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	ResourcePacks []*resource.Pack
	// TexturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	TexturePacksRequired bool

	// OnlyLogin specifies if the Conn returned will first handle packets to establish a fully spawned client,
	// or if it should only handle packets to login.
	// If OnlyLogin is set to true, the first packet the returned Conn when calling Accept will expect is a
	// ResourcePackInfo packet.
	OnlyLogin bool

	// playerCount is the amount of players connected to the server. If MaximumPlayers is non-zero and equal
	// to the playerCount, no more players will be accepted.
	playerCount int

	listener net.Listener

	hijackingPong bool
	incoming      chan *Conn
	close         chan bool
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
		ErrorLog:           log.New(os.Stderr, "", log.LstdFlags),
		MaximumChunkRadius: 16,
		InitialGameData: func(*Conn) GameData {
			return GameData{}
		},
		ServerName: "Minecraft Server",
		listener:   netListener,
		close:      make(chan bool, 2),
		incoming:   make(chan *Conn),
	}

	// Actually start listening.
	go listener.listen()
	return listener, nil
}

// Accept accepts a fully connected (on Minecraft layer) connection which is ready to receive and send
// packets. It is recommended to cast the net.Conn returned to a *minecraft.Conn so that it is possible to
// use the conn.ReadPacket() and conn.WritePacket() methods.
func (listener *Listener) Accept() (net.Conn, error) {
	return <-listener.incoming, nil
}

// HijackPong hijacks the pong response from a server at an address passed. The listener passed will
// continuously update its pong data by hijacking the pong data of the server at the address.
// The hijack will last until the listener is shut down.
// If the address passed could not be resolved, an error is returned.
// Calling HijackPong means that any current and future pong data set using listener.PongData is overwritten
// each update.
func (listener *Listener) HijackPong(address string) error {
	listener.hijackingPong = true
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

// updatePongData updates the pong data of the listener using the current only players, maximum players and
// server name of the listener, provided the listener isn't currently hijacking the pong of another server.
func (listener *Listener) updatePongData() {
	if listener.hijackingPong {
		return
	}
	maxCount := listener.MaximumPlayers
	if maxCount == 0 {
		// If the maximum amount of allowed players is 0, we set it to the the current amount of line players
		// plus 1, so that new players can always join.
		maxCount = listener.playerCount + 1
	}

	rakListener := listener.listener.(*raknet.Listener)
	rakListener.PongData([]byte(fmt.Sprintf("MCPE;%v;%v;%v;%v;%v;%v;Minecraft Server;;",
		listener.ServerName, protocol.CurrentProtocol, protocol.CurrentVersion, listener.playerCount, maxCount, rakListener.ID(),
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
		conn := newConn(netConn, nil, listener.ErrorLog)
		conn.texturePacksRequired = listener.TexturePacksRequired
		conn.resourcePacks = listener.ResourcePacks
		conn.serverName = listener.ServerName
		conn.initialChunkRadius = int32(listener.MaximumChunkRadius)
		conn.onlyLogin = listener.OnlyLogin

		if listener.playerCount == listener.MaximumPlayers && listener.MaximumPlayers != 0 {
			// The server was full. We kick the player immediately and close the connection.
			_ = conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusLoginFailedServerFull})
			_ = conn.Close()
			continue
		}
		listener.playerCount++
		listener.updatePongData()

		hasGameData := false

		go func() {
			defer func() {
				_ = conn.Close()
				listener.playerCount--
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
				if !hasGameData {
					// We wait with this until the Conn receives and handles its first packet. The Login packet
					// contains information which may be relevant for the user to create game data for.
					conn.gameData = listener.InitialGameData(conn)
					hasGameData = true
				}
			}
		}()
	}
}
