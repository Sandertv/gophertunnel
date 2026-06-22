package realms

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestRealmAddressRequestsImmediately(t *testing.T) {
	requests := make(chan string, 1)
	c := &Client{
		requestFunc: func(_ context.Context, method, path string) ([]byte, int, error) {
			requests <- method + " " + path
			return []byte(`{"address":"127.0.0.1:19132","networkProtocol":"DEFAULT"}`), http.StatusOK, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	addr, err := c.RealmAddress(ctx, 42)
	if err != nil {
		t.Fatalf("RealmAddress: %v", err)
	}
	if addr.Address != "127.0.0.1:19132" || addr.NetworkProtocol != NetworkProtocolDefault {
		t.Fatalf("RealmAddress = %+v", addr)
	}
	select {
	case got := <-requests:
		if got != "GET /worlds/42/join" {
			t.Fatalf("request = %q", got)
		}
	default:
		t.Fatal("RealmAddress did not request before waiting for the poll ticker")
	}
}

func TestRealmAddressPollsAfterServiceUnavailable(t *testing.T) {
	attempts := 0
	c := &Client{
		requestFunc: func(_ context.Context, _, _ string) ([]byte, int, error) {
			attempts++
			return nil, http.StatusServiceUnavailable, errors.New("starting")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if _, err := c.RealmAddress(ctx, 42); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RealmAddress error = %v, want context deadline", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want exactly one immediate attempt before poll wait", attempts)
	}
}
