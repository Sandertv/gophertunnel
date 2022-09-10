package minecraft

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"go.uber.org/atomic"
	"log"
	"net"
	"os"
	"time"
)

// ListenConfig holds settings that may be edited to change behaviour of a Listener.
type ListenConfig struct {
	// ErrorLog is a log.Logger that errors that occur during packet handling of clients are written to. By
	// default, ErrorLog is set to one equal to the global logger.
	ErrorLog *log.Logger

	// AuthenticationDisabled specifies if authentication of players that join is disabled. If set to true, no
	// verification will be done to ensure that the player connecting is authenticated using their XBOX Live
	// account.
	AuthenticationDisabled bool

	// MaximumPlayers is the maximum amount of players accepted in the server. If non-zero, players that
	// attempt to join while the server is full will be kicked during login. If zero, the maximum player count
	// will be dynamically updated each time a player joins, so that an unlimited amount of players is
	// accepted into the server.
	MaximumPlayers int

	// StatusProvider is the ServerStatusProvider of the Listener. When set to nil, the default provider,
	// ListenerStatusProvider, is used as provider.
	StatusProvider ServerStatusProvider

	// AcceptedProtocols is a slice of Protocol accepted by a Listener created with this ListenConfig. The current
	// Protocol is always added to this slice. Clients with a protocol version that is not present in this slice will
	// be disconnected.
	AcceptedProtocols []Protocol

	// ResourcePacks is a slice of resource packs that the listener may hold. Each client will be asked to
	// download these resource packs upon joining.
	// This field should not be edited during runtime of the Listener to avoid race conditions. Use
	// Listener.AddResourcePack() to add a resource pack after having called Listener.Listen().
	ResourcePacks []*resource.Pack
	// Biomes contains information about all biomes that the server has registered, which the client can use
	// to render the world more effectively. If these are nil, the default biome definitions will be used.
	Biomes map[string]any
	// TexturePacksRequired specifies if clients that join must accept the texture pack in order for them to
	// be able to join the server. If they don't accept, they can only leave the server.
	TexturePacksRequired bool

	// PacketFunc is called whenever a packet is read from or written to a connection returned when using
	// Listener.Accept. It includes packets that are otherwise covered in the connection sequence, such as the
	// Login packet. The function is called with the header of the packet and its raw payload, the address
	// from which the packet originated, and the destination address.
	PacketFunc func(header packet.Header, payload []byte, src, dst net.Addr)
}

// Listener implements a Minecraft listener on top of an unspecific net.Listener. It abstracts away the
// login sequence of connecting clients and provides the implements the net.Listener interface to provide a
// consistent API.
type Listener struct {
	cfg      ListenConfig
	listener NetworkListener

	// playerCount is the amount of players connected to the server. If MaximumPlayers is non-zero and equal
	// to the playerCount, no more players will be accepted.
	playerCount atomic.Int32

	incoming chan *Conn
	close    chan struct{}

	key *ecdsa.PrivateKey
}

// Listen announces on the local network address. The network is typically "raknet".
// If the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all
// available unicast and anycast IP addresses of the local system.
func (cfg ListenConfig) Listen(network string, address string) (*Listener, error) {
	n, ok := networkByID(network)
	if !ok {
		return nil, fmt.Errorf("listen: no network under id: %v", network)
	}

	netListener, err := n.Listen(address)
	if err != nil {
		return nil, err
	}

	if cfg.ErrorLog == nil {
		cfg.ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
	}
	if cfg.StatusProvider == nil {
		cfg.StatusProvider = NewStatusProvider("Minecraft Server")
	}
	key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	listener := &Listener{
		cfg:      cfg,
		listener: netListener,
		incoming: make(chan *Conn),
		close:    make(chan struct{}),
		key:      key,
	}

	// Actually start listening.
	go listener.listen()
	return listener, nil
}

// Listen announces on the local network address. The network must be "tcp", "tcp4", "tcp6", "unix",
// "unixpacket" or "raknet". A Listener is returned which may be used to accept connections.
// If the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all
// available unicast and anycast IP addresses of the local system.
// Listen has the default values for the fields of Listener filled out. To use different values for these
// fields, call &Listener{}.Listen() instead.
func Listen(network, address string) (*Listener, error) {
	var lc ListenConfig
	return lc.Listen(network, address)
}

// Accept accepts a fully connected (on Minecraft layer) connection which is ready to receive and send
// packets. It is recommended to cast the net.Conn returned to a *minecraft.Conn so that it is possible to
// use the Conn.ReadPacket() and Conn.WritePacket() methods.
// Accept returns an error if the listener is closed.
func (listener *Listener) Accept() (net.Conn, error) {
	conn, ok := <-listener.incoming
	if !ok {
		return nil, &net.OpError{Op: "accept", Net: "minecraft", Addr: listener.Addr(), Err: errListenerClosed}
	}
	return conn, nil
}

// Disconnect disconnects a Minecraft Conn passed by first sending a disconnect with the message passed, and
// closing the connection after. If the message passed is empty, the client will be immediately sent to the
// server list instead of a disconnect screen.
func (listener *Listener) Disconnect(conn *Conn, message string) error {
	_ = conn.WritePacket(&packet.Disconnect{
		HideDisconnectionScreen: message == "",
		Message:                 message,
	})
	return conn.Close()
}

// Addr returns the address of the underlying listener.
func (listener *Listener) Addr() net.Addr {
	return listener.listener.Addr()
}

// Close closes the listener and the underlying net.Listener. Pending calls to Accept will fail immediately.
func (listener *Listener) Close() error {
	return listener.listener.Close()
}

// updatePongData updates the pong data of the listener using the current only players, maximum players and
// server name of the listener, provided the listener isn't currently hijacking the pong of another server.
func (listener *Listener) updatePongData() {
	s := listener.status()
	listener.listener.PongData([]byte(fmt.Sprintf("MCPE;%v;%v;%v;%v;%v;%v;Gophertunnel;%v;%v;%v;%v;",
		s.ServerName, protocol.CurrentProtocol, protocol.CurrentVersion, s.PlayerCount, s.MaxPlayers,
		listener.listener.ID(), "Creative", 1, listener.Addr().(*net.UDPAddr).Port, listener.Addr().(*net.UDPAddr).Port,
	)))
}

// listen starts listening for incoming connections and packets. When a player is fully connected, it submits
// it to the accepted connections channel so that a call to Accept can pick it up.
func (listener *Listener) listen() {
	listener.updatePongData()
	go func() {
		ticker := time.NewTicker(time.Second * 4)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				listener.updatePongData()
			case <-listener.close:
				return
			}
		}
	}()
	defer func() {
		close(listener.incoming)
		close(listener.close)
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
	conn := newConn(netConn, listener.key, listener.cfg.ErrorLog)
	conn.acceptedProto = append(listener.cfg.AcceptedProtocols, proto{})
	// Temporarily set the protocol to the latest: We don't know the actual protocol until we read the Login packet.
	conn.proto = proto{}
	conn.pool = conn.proto.Packets()

	conn.packetFunc = listener.cfg.PacketFunc
	conn.texturePacksRequired = listener.cfg.TexturePacksRequired
	conn.resourcePacks = listener.cfg.ResourcePacks
	conn.biomes = listener.cfg.Biomes
	conn.gameData.WorldName = listener.status().ServerName
	conn.authEnabled = !listener.cfg.AuthenticationDisabled

	if listener.playerCount.Load() == int32(listener.cfg.MaximumPlayers) && listener.cfg.MaximumPlayers != 0 {
		// The server was full. We kick the player immediately and close the connection.
		_ = conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusLoginFailedServerFull})
		_ = conn.Close()
		return
	}
	listener.playerCount.Add(1)
	listener.updatePongData()

	go listener.handleConn(conn)
}

// status returns the current ServerStatus of the Listener.
func (listener *Listener) status() ServerStatus {
	status := listener.cfg.StatusProvider.ServerStatus(int(listener.playerCount.Load()), listener.cfg.MaximumPlayers)
	if status.MaxPlayers == 0 {
		status.MaxPlayers = status.PlayerCount + 1
	}
	return status
}

// handleConn handles an incoming connection of the Listener. It will first attempt to get the connection to
// log in, after which it will expose packets received to the user.
func (listener *Listener) handleConn(conn *Conn) {
	defer func() {
		_ = conn.Close()
		listener.playerCount.Add(-1)
		listener.updatePongData()
	}()
	for {
		// We finally arrived at the packet decoding loop. We constantly decode packets that arrive
		// and push them to the Conn so that they may be processed.
		packets, err := conn.dec.Decode()
		if err != nil {
			if !raknet.ErrConnectionClosed(err) {
				listener.cfg.ErrorLog.Printf("error reading from listener connection: %v\n", err)
			}
			return
		}
		for _, data := range packets {
			loggedInBefore := conn.loggedIn
			if err := conn.receive(data); err != nil {
				listener.cfg.ErrorLog.Printf("error: %v", err)
				return
			}
			if !loggedInBefore && conn.loggedIn {
				select {
				case <-listener.close:
					// The listener was closed while this one was logged in, so the incoming channel will be
					// closed. Just return so the connection is closed and cleaned up.
					return
				case listener.incoming <- conn:
					// The connection was previously not logged in, but was after receiving this packet,
					// meaning the connection is fully completely now. We add it to the channel so that
					// a call to Accept() can receive it.
				}
			}
		}
	}
}
