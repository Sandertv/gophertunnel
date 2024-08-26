package discovery

import (
	"context"
	"errors"
	"github.com/sandertv/gophertunnel/minecraft/room"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestListen(t *testing.T) {
	cfg := ListenConfig{
		Log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	l, err := cfg.Listen("udp", ":7551")
	if err != nil {
		t.Fatalf("error listening: %s", err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatalf("error closing: %s", err)
		}
	})

	_ = l.Announce(room.Status{
		HostName:       "Da1z981?",
		WorldName:      "LAN のデバッグ",
		WorldType:      room.WorldTypeCreative,
		MemberCount:    1,
		MaxMemberCount: 30,
		IsEditorWorld:  false,
		TransportLayer: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	for {
		signal, err := l.ReadSignal(ctx.Done())
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("error reading signal: %s", err)
			}
			return
		}
		t.Logf("%#v", signal)
	}
}
