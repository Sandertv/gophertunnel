package minecraft

import (
	"github.com/sandertv/go-raknet"
	"net"
	"strconv"
	"sync"
	"time"
)

// ServerStatusProvider represents a type that is able to provide the visual status of a server, in specific
// its MOTD, the amount of online players and the player limit displayed in the server list.
// These providers may be used to display different information in the server list. Although they overwrite
// the maximum players and online players maintained by a Listener in the server list, these values are not
// changed and will still be used internally to check if players are able to be connected. Players will still
// be disconnected if the maximum player count as set in the ListenConfig.MaximumPlayers field of a Listener
// is reached (unless ListenConfig.MaximumPlayers is 0).
type ServerStatusProvider interface {
	// ServerStatus returns the server status which includes the MOTD/server name, amount of online players
	// and the amount of maximum players.
	// The player count and max players of the minecraft.Listener that calls this method is passed.
	ServerStatus(playerCount, maxPlayers int) ServerStatus
}

// ServerStatus holds the information shown in the Minecraft server list. They have no impact on the listener
// functionality-wise.
type ServerStatus struct {
	// ServerName is the name or MOTD of the server, as shown in the server list.
	ServerName string
	// ServerName is the sub-name or sub-MOTD of the server, as shown in the friend list.
	ServerSubName string
	// PlayerCount is the current amount of players displayed in the list.
	PlayerCount int
	// MaxPlayers is the maximum amount of players in the server. If set to 0, MaxPlayers is set to
	// PlayerCount + 1.
	MaxPlayers int
}

// ListenerStatusProvider is the default ServerStatusProvider of a Listener. It displays a static server name/
// MOTD and displays the player count and maximum amount of players of the server.
type ListenerStatusProvider struct {
	// name is the name of the server, or the MOTD, that is displayed in the server list.
	name string
	// subName is the sub-name of the server, or the MOTD, that is displayed in the friend list.
	subName string
}

// NewStatusProvider creates a ListenerStatusProvider that displays the server name passed.
func NewStatusProvider(serverName, serverSubName string) ListenerStatusProvider {
	return ListenerStatusProvider{name: serverName, subName: serverSubName}
}

// ServerStatus ...
func (l ListenerStatusProvider) ServerStatus(playerCount, maxPlayers int) ServerStatus {
	return ServerStatus{
		ServerName:    l.name,
		ServerSubName: l.subName,
		PlayerCount:   playerCount,
		MaxPlayers:    maxPlayers,
	}
}

// ForeignStatusProvider is a ServerStatusProvider that provides the status of a target server to the Listener
// so that the MOTD, player count etc. is copied.
type ForeignStatusProvider struct {
	addr string

	mu     sync.Mutex
	status ServerStatus

	once   sync.Once
	closed chan struct{}
}

// NewForeignStatusProvider creates a ForeignStatusProvider that uses the status of the server running at the
// target address. An error is returned if the address is invalid.
// Close must be called if the ForeignStatusProvider is discarded.
func NewForeignStatusProvider(addr string) (*ForeignStatusProvider, error) {
	if _, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return nil, err
	}
	f := &ForeignStatusProvider{addr: addr, closed: make(chan struct{})}
	go f.update()
	return f, nil
}

// ServerStatus returns the ServerStatus of the target server.
func (f *ForeignStatusProvider) ServerStatus(int, int) ServerStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.status
}

// Close closes the ForeignStatusProvider and stops polling it. Close always returns nil.
func (f *ForeignStatusProvider) Close() error {
	f.once.Do(func() {
		close(f.closed)
	})
	return nil
}

// update updates the status every second and cancels when Close is called.
func (f *ForeignStatusProvider) update() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			data, err := raknet.Ping(f.addr)
			if err != nil {
				continue
			}
			f.mu.Lock()
			f.status = parsePongData(data)
			f.mu.Unlock()
		case <-f.closed:
			return
		}
	}
}

// parsePongData parses the unconnected pong data passed into the relevant fields of a ServerStatus struct.
func parsePongData(pong []byte) ServerStatus {
	frag := splitPong(string(pong))
	if len(frag) < 7 {
		return ServerStatus{ServerName: "Invalid pong data"}
	}
	serverName := frag[1]
	serverSubName := frag[6]
	online, err := strconv.Atoi(frag[4])
	if err != nil {
		return ServerStatus{ServerName: "Invalid player count"}
	}
	max, err := strconv.Atoi(frag[5])
	if err != nil {
		return ServerStatus{ServerName: "Invalid max player count"}
	}
	return ServerStatus{
		ServerName:    serverName,
		ServerSubName: serverSubName,
		PlayerCount:   online,
		MaxPlayers:    max,
	}
}
