package minecraft_test

import (
	"io"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TestPacketsReadInOrder writes a long burst of packets in large batches and
// fails on the first packet read out of order.
func TestPacketsReadInOrder(t *testing.T) {
	const (
		burst = 5000
		batch = 512
	)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	l, err := minecraft.ListenConfig{AuthenticationDisabled: true, FlushRate: -1, ErrorLog: log}.Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	dialed := make(chan *minecraft.Conn, 1)
	go func() {
		defer close(dialed)
		c, err := minecraft.Dialer{IdentityData: login.IdentityData{DisplayName: "Order"}, ErrorLog: log}.DialTimeout("raknet", l.Addr().String(), 10*time.Second)
		if err != nil {
			return
		}
		if err := c.DoSpawnTimeout(10 * time.Second); err != nil {
			_ = c.Close()
			return
		}
		dialed <- c
	}()
	accepted, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	server := accepted.(*minecraft.Conn)
	defer server.Close()
	if err := server.StartGameTimeout(minecraft.GameData{EntityRuntimeID: 1, EntityUniqueID: 1}, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	client, ok := <-dialed
	if !ok {
		t.Fatal("client did not spawn")
	}
	defer client.Close()

	go func() {
		for i := range burst {
			if err := server.WritePacket(&packet.Text{TextType: packet.TextTypeRaw, Message: strconv.Itoa(i)}); err != nil {
				return
			}
			if i%batch == batch-1 {
				if err := server.Flush(); err != nil {
					return
				}
			}
		}
		_ = server.Flush()
	}()

	if err := client.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	for next := 0; next < burst; {
		pk, err := client.ReadPacket()
		if err != nil {
			t.Fatalf("read %d of %d packets: %v", next, burst, err)
		}
		text, ok := pk.(*packet.Text)
		if !ok {
			continue
		}
		if text.Message != strconv.Itoa(next) {
			t.Fatalf("read packet %s where %d was next", text.Message, next)
		}
		next++
	}
}
