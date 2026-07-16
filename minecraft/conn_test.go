package minecraft

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestReceiveDisconnectReturnsTypedError(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	conn.pool = conn.proto.Packets(false)
	defer conn.Close()

	const message = "You are not whitelisted on this server"
	var buf bytes.Buffer
	header := packet.Header{PacketID: packet.IDDisconnect}
	if err := header.Write(&buf); err != nil {
		t.Fatal(err)
	}
	(&packet.Disconnect{Message: message}).Marshal(protocol.NewWriter(&buf, 0))
	if err := conn.receive(buf.Bytes()); err != nil {
		t.Fatal(err)
	}

	var disconnect DisconnectError
	if cause := context.Cause(conn.Context()); !errors.As(cause, &disconnect) {
		t.Fatalf("disconnect cause %v does not contain DisconnectError", cause)
	} else if disconnect.Error() != message {
		t.Fatalf("disconnect message = %q, want %q", disconnect.Error(), message)
	}
}
