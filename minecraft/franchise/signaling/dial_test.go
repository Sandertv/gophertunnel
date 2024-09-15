package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
	"math/rand"
	"os"
	"testing"
	"time"
)

// TestDial demonstrates dialing a Conn using [Dialer.DialContext] and ensures that the notification is working correctly.
func TestDial(t *testing.T) {
	tok, err := readToken("../internal/test/auth.tok", auth.TokenSource)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	src := auth.RefreshTokenSource(tok)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var d Dialer
	conn, err := d.DialContext(ctx, src)
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("error closing connection: %s", err)
		}
	})

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	conn.Notify(ctx, testNotifier{t})
	if err := conn.Signal(&nethernet.Signal{
		Type:         nethernet.SignalTypeOffer,
		ConnectionID: rand.Uint64(),
		NetworkID:    100, // Try signaling an offer to invalid network, We hopefully notify an Error.
	}); err != nil {
		t.Fatalf("error signaling offer: %s", err)
	}

	<-ctx.Done()
}

type testNotifier struct {
	testing.TB
}

func (n testNotifier) NotifySignal(signal *nethernet.Signal) {
	n.Logf("NotifySignal(%s)", signal)
}

func (n testNotifier) NotifyError(err error) {
	n.Logf("NotifyError(%s)", err)
}

func readToken(path string, src oauth2.TokenSource) (t *oauth2.Token, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t, err = src.Token()
		if err != nil {
			return nil, fmt.Errorf("obtain token: %w", err)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(t); err != nil {
			return nil, fmt.Errorf("encode: %w", err)
		}
		return t, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return t, nil
}
