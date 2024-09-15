package room

import (
	"errors"
	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/room/internal"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"
)

// ListenConfig holds the configuration for wrapping a [minecraft.NetworkListener] with additional functionality.
// It provides the ability to announce server status and custom the behavior of status reporting.
type ListenConfig struct {
	// Announcer announces the Status of Listener. It is called from [Listener.ServerStatus] to report the status
	// to external services like Xbox Live and LAN discovery. If nil, the Wrap method will panic.
	Announcer Announcer

	// StatusProvider provides the Status for announcing using the Announcer. It will be called by [Listener.ServerStatus]
	// at some intervals. If nil, a default StatusProvider reporting DefaultStatus will be set.
	StatusProvider StatusProvider

	// DisableServerStatusOverride indicates that fields of the Status provided by the StatusProvider should be modified
	// to sync with the [minecraft.ServerStatus] reported from [minecraft.Listener]. It includes fields like [Status.MemberCount],
	// [Status.MaxMemberCount], [Status.WorldName], and [Status.HostName].
	DisableServerStatusOverride bool // TODO: Find a good name

	// Log is used for logging messages at various log levels. If nil, the default [slog.Logger]
	// will be set from [slog.Default].
	Log *slog.Logger
}

// Wrap wraps the [minecraft.NetworkListener] with additional functionality provided by Listener. It returns
// a new [Listener] that hijacks the [minecraft.ServerStatus] of the underlying listener and announces it using
// the [Announcer] and the [Status] provided by the [StatusProvider].
func (conf ListenConfig) Wrap(n minecraft.NetworkListener) *Listener {
	if conf.Announcer == nil {
		panic("minecraft/room: ListenConfig.Wrap: Announcer is nil")
	}
	if conf.StatusProvider == nil {
		conf.StatusProvider = NewStatusProvider(DefaultStatus())
	}

	return &Listener{
		conf: conf,

		n: n,

		closed: make(chan struct{}, 1),
	}
}

// Listener wraps a [minecraft.NetworkListener], allowing it to announce [minecraft.ServerStatus] using
// an Announcer. It can be created using [ListenConfig.Wrap].
type Listener struct {
	conf ListenConfig

	n minecraft.NetworkListener

	closed chan struct{} // Notifies that the Listener has been closed.
	once   sync.Once     // Closes Listener only once.
}

// Accept waits for and returns the next [net.Conn] to the underlying [minecraft.NetworkListener].
// An error may be returned if the Listener has been closed.
func (l *Listener) Accept() (net.Conn, error) { return l.n.Accept() }

// Addr returns the [net.Addr] of the underlying [minecraft.NetworkListener].
func (l *Listener) Addr() net.Addr { return l.n.Addr() }

// ID returns the unique ID of the underlying [minecraft.NetworkListener].
func (l *Listener) ID() int64 { return l.n.ID() }

// PongData updates the pong data on the underlying [minecraft.NetworkListener].
func (l *Listener) PongData(data []byte) { l.n.PongData(data) }

// ServerStatus reports the [minecraft.ServerStatus] to the Announcer with a Status provided by
// the StatusProvider. If [ListenConfig.DisableServerStatusOverride] is false, the fields of the
// Status will be modified to sync with the [minecraft.ServerStatus]. This includes updating member
// counts, world names, host names, and connections based on the address type of the [minecraft.NetworkListener].
func (l *Listener) ServerStatus(server minecraft.ServerStatus) {
	status := l.conf.StatusProvider.RoomStatus()
	if !l.conf.DisableServerStatusOverride {
		status.MemberCount = server.PlayerCount
		status.MaxMemberCount = server.MaxPlayers

		status.WorldName = server.ServerName
		status.HostName = server.ServerSubName

		switch addr := l.n.Addr().(type) {
		case *nethernet.Addr:
			if status.TransportLayer == 0 {
				status.TransportLayer = TransportLayerNetherNet
			}
			if status.WebRTCNetworkID == 0 {
				status.WebRTCNetworkID = addr.NetworkID
			}
			status.SupportedConnections = append(status.SupportedConnections, Connection{
				ConnectionType:  ConnectionTypeWebSocketsWebRTCSignaling,
				NetherNetID:     addr.NetworkID,
				WebRTCNetworkID: addr.NetworkID,
			})
		case *net.UDPAddr:
			if status.TransportLayer == 0 {
				status.TransportLayer = TransportLayerRakNet
			}
			if status.RakNetGUID == "" {
				status.RakNetGUID = strconv.FormatInt(l.n.ID(), 10)
			}
			status.SupportedConnections = append(status.SupportedConnections, Connection{
				ConnectionType: ConnectionTypeUPNP,
				HostIPAddress:  addr.IP.String(),
				HostPort:       uint16(addr.Port),
				RakNetGUID:     strconv.FormatInt(l.n.ID(), 10),
			})
		default:
			l.conf.Log.Debug("unsupported address type", slog.Any("address", addr))
		}
	}

	if err := l.conf.Announcer.Announce(&listenerContext{closed: l.closed}, status); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			l.conf.Log.Error("error announcing status", internal.ErrAttr(err))
		}
	}
}

// Close closes the Listener. Any blocking methods will be canceled through its internal context.
func (l *Listener) Close() (err error) {
	l.once.Do(func() {
		close(l.closed)
		err = errors.Join(
			l.n.Close(),
			l.conf.Announcer.Close(),
		)
	})
	return err
}

// listenerContext implements [context.Context] for a Listener.
type listenerContext struct{ closed <-chan struct{} }

// Deadline returns the zero time and false, indicating that deadlines are not used.
func (*listenerContext) Deadline() (zero time.Time, _ bool) {
	return zero, false
}

// Done returns a channel that is closed when the Listener is closed.
func (ctx *listenerContext) Done() <-chan struct{} { return ctx.closed }

// Err returns net.ErrClosed if the Listener has been closed. Returns nil otherwise.
func (ctx *listenerContext) Err() error {
	select {
	case <-ctx.closed:
		return net.ErrClosed
	default:
		return nil
	}
}

// Value returns nil for any key, as no values are associated with the context.
func (*listenerContext) Value(any) any { return nil }
