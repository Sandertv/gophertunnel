package minecraft

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync/atomic"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/internal"
)

func TestRakNetPingContextUsesUpstreamDialer(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	network := RakNet{
		l: slog.New(internal.DiscardHandler{}),
		UpstreamDialer: upstreamDialerFunc(func(ctx context.Context, network, address string) (net.Conn, error) {
			calls.Add(1)
			if network != "udp" {
				t.Fatalf("network = %q, want udp", network)
			}
			if address != "127.0.0.1:19132" {
				t.Fatalf("address = %q, want 127.0.0.1:19132", address)
			}
			return nil, ctx.Err()
		}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := network.PingContext(ctx, "127.0.0.1:19132")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("PingContext error = %v, want context.Canceled", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("upstream dial calls = %d, want 1", calls.Load())
	}
}

func TestRakNetDialContextUsesUpstreamDialer(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	network := RakNet{
		l: slog.New(internal.DiscardHandler{}),
		UpstreamDialer: upstreamDialerFunc(func(ctx context.Context, network, address string) (net.Conn, error) {
			calls.Add(1)
			if network != "udp" {
				t.Fatalf("network = %q, want udp", network)
			}
			if address != "127.0.0.1:19132" {
				t.Fatalf("address = %q, want 127.0.0.1:19132", address)
			}
			return nil, ctx.Err()
		}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := network.DialContext(ctx, "127.0.0.1:19132")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DialContext error = %v, want context.Canceled", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("upstream dial calls = %d, want 1", calls.Load())
	}
}

func TestRakNetPingContextAllowsNilLogger(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	network := RakNet{
		UpstreamDialer: upstreamDialerFunc(func(ctx context.Context, network, address string) (net.Conn, error) {
			calls.Add(1)
			return nil, ctx.Err()
		}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := network.PingContext(ctx, "127.0.0.1:19132")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("PingContext error = %v, want context.Canceled", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("upstream dial calls = %d, want 1", calls.Load())
	}
}

type upstreamDialerFunc func(context.Context, string, string) (net.Conn, error)

func (f upstreamDialerFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}
