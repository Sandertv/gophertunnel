package room

import (
	"context"
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/room/internal"
	"log/slog"
	"net"
	"sync"
)

type ListenConfig struct {
	StatusProvider StatusProvider
	Log            *slog.Logger
}

func (conf ListenConfig) Listen(a Announcer, n minecraft.NetworkListener) (*Listener, error) {
	if conf.StatusProvider == nil {
		conf.StatusProvider = NewStatusProvider(DefaultStatus())
	}
	if conf.Log == nil {
		conf.Log = slog.Default()
	}

	l := &Listener{
		conf: conf,

		announcer: a,
		listener:  n,

		closed: make(chan struct{}),
	}

	return l, nil
}

type Listener struct {
	conf ListenConfig

	announcer Announcer
	listener  minecraft.NetworkListener

	closed chan struct{}
	once   sync.Once
}

func (l *Listener) ID() int64 {
	return l.listener.ID()
}

func (l *Listener) PongData(data []byte) {
	l.listener.PongData(data)
}

func (l *Listener) Accept() (net.Conn, error) {
	return l.listener.Accept()
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *Listener) ServerStatus(serverStatus minecraft.ServerStatus) {
	status := l.conf.StatusProvider.RoomStatus()

	status.HostName = serverStatus.ServerSubName
	status.WorldName = serverStatus.ServerName

	status.MemberCount = serverStatus.PlayerCount
	status.MaxMemberCount = serverStatus.MaxPlayers

	// TODO
	status.SupportedConnections = []Connection{
		{
			ConnectionType:  ConnectionTypeWebSocketsWebRTCSignaling, // ...
			NetherNetID:     uint64(l.listener.ID()),
			WebRTCNetworkID: uint64(l.listener.ID()),
		},
	}
	status.WebRTCNetworkID = uint64(l.listener.ID())

	go l.announce(status)
	return
}

func (l *Listener) announce(status Status) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-l.closed:
			cancel()
		}
	}()

	if err := l.announcer.Announce(ctx, status); err != nil {
		l.conf.Log.Error("error announcing status", internal.ErrAttr(err))
	}
}

func (l *Listener) Close() (err error) {
	l.once.Do(func() {
		close(l.closed)

		fmt.Println("close called")

		errs := []error{l.listener.Close()}
		if err := l.announcer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close announcer: %w", err))
		}
		err = errors.Join(errs...)
	})
	return err
}
