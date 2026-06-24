package minecraft

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestStartGameWritesPropertyData(t *testing.T) {
	t.Parallel()

	client, serverConn := net.Pipe()
	defer client.Close()
	defer serverConn.Close()
	go func() {
		_, _ = io.Copy(io.Discard, serverConn)
	}()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	defer conn.Close()

	var got map[string]any
	conn.packetFunc = func(header packet.Header, payload []byte, _, _ net.Addr) {
		if header.PacketID != packet.IDStartGame {
			return
		}
		var start packet.StartGame
		start.Marshal(protocol.NewReader(bytes.NewBuffer(payload), 0, false))
		got = start.PropertyData
	}

	if err := conn.SendStartGame(GameData{
		PropertyData: map[string]any{
			"gophertunnel:test": int32(1),
		},
	}); err != nil {
		t.Fatalf("SendStartGame: %v", err)
	}

	if got["gophertunnel:test"] != int32(1) {
		t.Fatalf("StartGame.PropertyData = %#v, want gophertunnel:test=1", got)
	}
}

func TestDisconnectWritesDisconnectPacket(t *testing.T) {
	t.Parallel()

	client, serverConn := net.Pipe()
	defer client.Close()

	conn := newConn(serverConn, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.Disconnect("closing")
	}()

	if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	packets, err := packet.NewDecoder(client).Decode()
	if err != nil {
		t.Fatalf("decode disconnect packet: %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("decoded packet count = %d, want 1", len(packets))
	}
	buf := bytes.NewBuffer(packets[0])
	var header packet.Header
	if err := header.Read(buf); err != nil {
		t.Fatalf("read packet header: %v", err)
	}
	if header.PacketID != packet.IDDisconnect {
		t.Fatalf("packet ID = %d, want Disconnect", header.PacketID)
	}
	var disconnect packet.Disconnect
	disconnect.Marshal(protocol.NewReader(buf, 0, false))
	if disconnect.Message != "closing" {
		t.Fatalf("disconnect message = %q, want closing", disconnect.Message)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
}

func TestDisconnectPacketWritesDisconnectReason(t *testing.T) {
	t.Parallel()

	client, serverConn := net.Pipe()
	defer client.Close()

	conn := newConn(serverConn, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.DisconnectPacket(packet.Disconnect{
			Reason:          packet.DisconnectReasonServerFull,
			FilteredMessage: "Server Full",
		})
	}()

	if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	packets, err := packet.NewDecoder(client).Decode()
	if err != nil {
		t.Fatalf("decode disconnect packet: %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("decoded packet count = %d, want 1", len(packets))
	}
	buf := bytes.NewBuffer(packets[0])
	var header packet.Header
	if err := header.Read(buf); err != nil {
		t.Fatalf("read packet header: %v", err)
	}
	if header.PacketID != packet.IDDisconnect {
		t.Fatalf("packet ID = %d, want Disconnect", header.PacketID)
	}
	var disconnect packet.Disconnect
	disconnect.Marshal(protocol.NewReader(buf, 0, false))
	if disconnect.Reason != packet.DisconnectReasonServerFull {
		t.Fatalf("disconnect reason = %d, want %d", disconnect.Reason, packet.DisconnectReasonServerFull)
	}
	if disconnect.FilteredMessage != "Server Full" {
		t.Fatalf("filtered message = %q, want Server Full", disconnect.FilteredMessage)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("DisconnectPacket: %v", err)
	}
}

func TestReceiveDisconnectPreservesPacketReason(t *testing.T) {
	t.Parallel()

	client, serverConn := net.Pipe()
	defer client.Close()
	defer serverConn.Close()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	conn.pool = conn.proto.Packets(false)
	defer conn.Close()

	var buf bytes.Buffer
	header := packet.Header{PacketID: packet.IDDisconnect}
	if err := header.Write(&buf); err != nil {
		t.Fatalf("write header: %v", err)
	}
	(&packet.Disconnect{
		Reason:          packet.DisconnectReasonServerFull,
		Message:         "",
		FilteredMessage: "Server Full",
	}).Marshal(protocol.NewWriter(&buf, 0))

	if err := conn.receive(buf.Bytes()); err != nil {
		t.Fatalf("receive disconnect: %v", err)
	}

	cause := context.Cause(conn.Context())
	var packetErr *DisconnectPacketError
	if !errors.As(cause, &packetErr) {
		t.Fatalf("cause %v does not contain DisconnectPacketError", cause)
	}
	if packetErr.Reason != packet.DisconnectReasonServerFull {
		t.Fatalf("reason = %d, want %d", packetErr.Reason, packet.DisconnectReasonServerFull)
	}
	if packetErr.FilteredMessage != "Server Full" {
		t.Fatalf("filtered message = %q, want Server Full", packetErr.FilteredMessage)
	}
	if packetErr.Error() != "Server Full" {
		t.Fatalf("error = %q, want Server Full", packetErr.Error())
	}

	var legacyErr DisconnectError
	if !errors.As(cause, &legacyErr) {
		t.Fatalf("cause %v does not contain legacy DisconnectError", cause)
	}
	if legacyErr.Error() != "Server Full" {
		t.Fatalf("legacy error = %q, want Server Full", legacyErr.Error())
	}
}

func TestClientToServerHandshakeMarksComplete(t *testing.T) {
	t.Parallel()

	client, serverConn := net.Pipe()
	defer client.Close()
	defer serverConn.Close()
	go func() {
		_, _ = io.Copy(io.Discard, serverConn)
	}()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, -1, false)
	defer conn.Close()

	if conn.handshakeComplete {
		t.Fatal("handshakeComplete was true before ClientToServerHandshake")
	}
	if err := conn.handleClientToServerHandshake(); err != nil {
		t.Fatalf("handleClientToServerHandshake: %v", err)
	}
	if !conn.handshakeComplete {
		t.Fatal("handshakeComplete was false after ClientToServerHandshake")
	}
}

func TestHandleResourcePacksInfoCountsURLDownloadedPacks(t *testing.T) {
	t.Parallel()

	urlPackID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	chunkPackID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	urlPack := testResourcePackArchive(t, urlPackID)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(urlPack)
	}))
	defer server.Close()

	client, serverConn := net.Pipe()
	defer client.Close()
	defer serverConn.Close()
	go func() {
		_, _ = io.Copy(io.Discard, serverConn)
	}()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, time.Second/20, false)
	defer conn.Close()

	err := conn.handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{
			UUID:        urlPackID,
			Version:     "1.0.0",
			Size:        uint64(len(urlPack)),
			DownloadURL: server.URL,
		},
		{
			UUID:    chunkPackID,
			Version: "1.0.0",
			Size:    1,
		},
	}})
	if err != nil {
		t.Fatalf("handleResourcePacksInfo: %v", err)
	}
	if conn.packQueue.packAmount != 1 {
		t.Fatalf("packAmount = %d, want 1", conn.packQueue.packAmount)
	}
	if _, ok := conn.packQueue.downloadingPacks[chunkPackID.String()]; !ok {
		t.Fatalf("chunk pack was not queued for chunk download")
	}
	if len(conn.resourcePacks) != 1 {
		t.Fatalf("resourcePacks length = %d, want 1", len(conn.resourcePacks))
	}
}

func TestHandleResourcePacksInfoFallsBackWhenURLExceedsAdvertisedSize(t *testing.T) {
	t.Parallel()

	urlPackID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	urlPack := testResourcePackArchive(t, urlPackID)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(append(urlPack, 0))
	}))
	defer server.Close()

	client, serverConn := net.Pipe()
	defer client.Close()
	defer serverConn.Close()
	go func() {
		_, _ = io.Copy(io.Discard, serverConn)
	}()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, time.Second/20, false)
	defer conn.Close()

	err := conn.handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{
			UUID:        urlPackID,
			Version:     "1.0.0",
			Size:        uint64(len(urlPack)),
			DownloadURL: server.URL,
		},
	}})
	if err != nil {
		t.Fatalf("handleResourcePacksInfo: %v", err)
	}
	if conn.packQueue.packAmount != 1 {
		t.Fatalf("packAmount = %d, want 1", conn.packQueue.packAmount)
	}
	if _, ok := conn.packQueue.downloadingPacks[urlPackID.String()]; !ok {
		t.Fatalf("oversized URL pack was not queued for chunk download fallback")
	}
	if len(conn.resourcePacks) != 0 {
		t.Fatalf("resourcePacks length = %d, want 0", len(conn.resourcePacks))
	}
}

func TestHandleResourcePacksInfoFallsBackWhenURLPackIdentityMismatch(t *testing.T) {
	t.Parallel()

	advertisedPackID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	urlPack := testResourcePackArchive(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(urlPack)
	}))
	defer server.Close()

	client, serverConn := net.Pipe()
	defer client.Close()
	defer serverConn.Close()
	go func() {
		_, _ = io.Copy(io.Discard, serverConn)
	}()

	conn := newConn(client, nil, slog.New(internal.DiscardHandler{}), DefaultProtocol, time.Second/20, false)
	defer conn.Close()

	err := conn.handleResourcePacksInfo(&packet.ResourcePacksInfo{TexturePacks: []protocol.TexturePackInfo{
		{
			UUID:        advertisedPackID,
			Version:     "1.0.0",
			Size:        uint64(len(urlPack)),
			DownloadURL: server.URL,
		},
	}})
	if err != nil {
		t.Fatalf("handleResourcePacksInfo: %v", err)
	}
	if conn.packQueue.packAmount != 1 {
		t.Fatalf("packAmount = %d, want 1", conn.packQueue.packAmount)
	}
	if _, ok := conn.packQueue.downloadingPacks[advertisedPackID.String()]; !ok {
		t.Fatalf("mismatched URL pack was not queued for chunk download fallback")
	}
	if len(conn.resourcePacks) != 0 {
		t.Fatalf("resourcePacks length = %d, want 0", len(conn.resourcePacks))
	}
}

func testResourcePackArchive(t *testing.T, id uuid.UUID) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	_, _ = w.Write([]byte(`{
		"format_version": 2,
		"header": {
			"name": "test pack",
			"description": "test pack",
			"uuid": "` + id.String() + `",
			"version": [1, 0, 0],
			"min_engine_version": [1, 20, 0]
		},
		"modules": [{
			"description": "test pack",
			"type": "resources",
			"uuid": "550e8400-e29b-41d4-a716-446655440001",
			"version": [1, 0, 0]
		}]
	}`))
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
