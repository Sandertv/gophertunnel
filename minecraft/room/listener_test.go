package room

import (
	"context"
	"net"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft"
)

func TestListenConfigWrapDefaultsLogger(t *testing.T) {
	t.Parallel()

	l := ListenConfig{
		Announcer: noopAnnouncer{},
	}.Wrap(fakeNetworkListener{addr: stringAddr("unsupported")})

	l.ServerStatus(minecraft.ServerStatus{})
}

type noopAnnouncer struct{}

func (noopAnnouncer) Announce(context.Context, Status) error { return nil }
func (noopAnnouncer) Close() error                           { return nil }

type fakeNetworkListener struct {
	addr net.Addr
}

func (f fakeNetworkListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (f fakeNetworkListener) Close() error              { return nil }
func (f fakeNetworkListener) Addr() net.Addr            { return f.addr }
func (f fakeNetworkListener) ID() int64                 { return 0 }
func (f fakeNetworkListener) PongData([]byte)           {}

type stringAddr string

func (s stringAddr) Network() string { return string(s) }
func (s stringAddr) String() string  { return string(s) }
