package minecraft

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/andreburgaud/crypt2go/padding"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"testing"
)

// TestDiscovery is a messed up test for LAN discovery. Its purpose is to
// debug encoding/decoding packets sent for discovery.
func TestDiscovery(t *testing.T) {
	// Please fill in this constant before running the test.
	const discoveryAddress = ":7551"

	l, err := net.ListenPacket("udp", discoveryAddress)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatalf("error closing discovery conn: %s", err)
		}
	})

	conn := &lan{
		networkID: rand.Uint64(),
		conn:      l,
		signals:   make(chan *nethernet.Signal),
		t:         t,
	}
	var cancel context.CancelCauseFunc
	conn.ctx, cancel = context.WithCancelCause(context.Background())
	go conn.background(cancel)

	RegisterNetwork("nethernet", &network{
		networkID: conn.networkID,
		signaling: conn,
	})

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	listener, err := Listen("nethernet", "")
	if err != nil {
		t.Fatalf("error listening: %s", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("error closing listener: %s", err)
		}
	})

	for {
		netConn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}
		minecraftConn := netConn.(*Conn)
		if err := minecraftConn.StartGame(GameData{
			WorldName:         "NetherNet",
			WorldSeed:         0,
			Difficulty:        0,
			EntityUniqueID:    rand.Int63(),
			EntityRuntimeID:   rand.Uint64(),
			PlayerGameMode:    1,
			PlayerPosition:    mgl32.Vec3{},
			WorldSpawn:        protocol.BlockPos{},
			WorldGameMode:     1,
			Time:              rand.Int63(),
			PlayerPermissions: 2,
		}); err != nil {
			t.Fatalf("error starting game: %s", err)
		}
	}
}

type lan struct {
	networkID uint64
	conn      net.PacketConn
	signals   chan *nethernet.Signal
	ctx       context.Context
	t         testing.TB
}

func (s *lan) WriteSignal(sig *nethernet.Signal) error {
	select {
	case <-s.ctx.Done():
		return context.Cause(s.ctx)
	default:
	}

	msg, err := encodePacket(s.networkID, &messagePacket{
		recipientID: sig.NetworkID,
		data:        sig.String(),
	})
	if err != nil {
		return fmt.Errorf("encode packet: %w", err)
	}
	_, err = s.conn.WriteTo(msg, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 7551,
	})
	return err
}

func (s *lan) ReadSignal() (*nethernet.Signal, error) {
	select {
	case <-s.ctx.Done():
		return nil, context.Cause(s.ctx)
	case signal := <-s.signals:
		return signal, nil
	}
}

func (s *lan) Credentials() (*nethernet.Credentials, error) {
	select {
	case <-s.ctx.Done():
		return nil, context.Cause(s.ctx)
	default:
		return nil, nil
	}
}

func (s *lan) background(cancel context.CancelCauseFunc) {
	for {
		b := make([]byte, 1024)
		n, _, err := s.conn.ReadFrom(b)
		if err != nil {
			cancel(err)
			return
		}
		senderID, pk, err := decodePacket(b[:n])
		if err != nil {
			s.t.Errorf("error decoding packet: %s", err)
			continue
		}
		if senderID == s.networkID {
			continue
		}
		switch pk := pk.(type) {
		case *requestPacket:
			err = s.handleRequest()
		case *messagePacket:
			err = s.handleMessage(senderID, pk)
		default:
			s.t.Logf("unhandled packet: %#v", pk)
		}
		if err != nil {
			s.t.Errorf("error handling packet (%#v): %s", pk, err)
		}
	}
}

func (s *lan) handleRequest() error {
	resp, err := encodePacket(s.networkID, &responsePacket{
		version:        0x2,
		serverName:     "Da1z981?",
		levelName:      "LAN Debugging",
		gameType:       2,
		playerCount:    1,
		maxPlayerCount: 30,
		editorWorld:    false,
		transportLayer: 2,
	})
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	if _, err := s.conn.WriteTo(resp, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 7551,
	}); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

func (s *lan) handleMessage(senderID uint64, pk *messagePacket) error {
	signal := &nethernet.Signal{}
	if err := signal.UnmarshalText([]byte(pk.data)); err != nil {
		return fmt.Errorf("decode signal: %w", err)
	}
	signal.NetworkID = senderID
	s.signals <- signal
	return nil
}

func encodePacket(senderID uint64, pk discoveryPacket) ([]byte, error) {
	buf := &bytes.Buffer{}
	pk.write(buf)

	headerBuf := &bytes.Buffer{}
	h := &packetHeader{
		length:   uint16(20 + buf.Len()),
		packetID: pk.id(),
		senderID: senderID,
	}
	h.write(headerBuf)
	payload := append(headerBuf.Bytes(), buf.Bytes()...)
	data, err := encryptECB(payload)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	hm := hmac.New(sha256.New, key[:])
	hm.Write(payload)
	data = append(append(hm.Sum(nil), data...))
	return data, nil
}

func decodePacket(b []byte) (uint64, discoveryPacket, error) {
	if len(b) < 32 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	data, err := decryptECB(b[32:])
	if err != nil {
		return 0, nil, fmt.Errorf("decrypt: %w", err)
	}

	hm := hmac.New(sha256.New, key[:])
	hm.Write(data)
	if checksum := hm.Sum(nil); !bytes.Equal(b[:32], checksum) {
		return 0, nil, fmt.Errorf("checksum mismatch: %x != %x", b[:32], checksum)
	}
	buf := bytes.NewBuffer(data)

	h := &packetHeader{}
	if err := h.read(buf); err != nil {
		return 0, nil, fmt.Errorf("decode header: %w", err)
	}
	var pk discoveryPacket
	switch h.packetID {
	case idRequest:
		pk = &requestPacket{}
	case idResponse:
		pk = &responsePacket{}
	case idMessage:
		pk = &messagePacket{}
	default:
		return h.senderID, nil, fmt.Errorf("unknown packet ID: %d", h.packetID)
	}
	if err := pk.read(buf); err != nil {
		return h.senderID, nil, fmt.Errorf("read payload: %w", err)
	}
	return h.senderID, pk, nil
}

const (
	idRequest uint16 = iota
	idResponse
	idMessage
)

type discoveryPacket interface {
	id() uint16
	read(buf *bytes.Buffer) error
	write(buf *bytes.Buffer)
}

type requestPacket struct{}

func (*requestPacket) id() uint16               { return idRequest }
func (*requestPacket) read(*bytes.Buffer) error { return nil }
func (*requestPacket) write(*bytes.Buffer)      {}

type responsePacket struct {
	version        uint8
	serverName     string
	levelName      string
	gameType       int32
	playerCount    int32
	maxPlayerCount int32
	editorWorld    bool
	transportLayer int32
}

func (*responsePacket) id() uint16 { return idResponse }
func (pk *responsePacket) read(buf *bytes.Buffer) error {
	var applicationDataLength uint32
	if err := binary.Read(buf, binary.LittleEndian, &applicationDataLength); err != nil {
		return fmt.Errorf("read application data length: %w", err)
	}
	data := buf.Next(int(applicationDataLength))
	n, err := hex.Decode(data, data)
	if err != nil {
		return fmt.Errorf("decode application data: %w", err)
	}

	a := bytes.NewBuffer(data[:n])

	if err := binary.Read(a, binary.LittleEndian, &pk.version); err != nil {
		return fmt.Errorf("read version: %w", err)
	}
	var length uint8
	if err := binary.Read(a, binary.LittleEndian, &length); err != nil {
		return fmt.Errorf("read server name length: %w", err)
	}
	pk.serverName = string(a.Next(int(length)))
	if err := binary.Read(a, binary.LittleEndian, &length); err != nil {
		return fmt.Errorf("read level name length: %w", err)
	}
	pk.levelName = string(a.Next(int(length)))
	if err := binary.Read(a, binary.LittleEndian, &pk.gameType); err != nil {
		return fmt.Errorf("read game type: %w", err)
	}
	if err := binary.Read(a, binary.LittleEndian, &pk.playerCount); err != nil {
		return fmt.Errorf("read player count: %w", err)
	}
	if err := binary.Read(a, binary.LittleEndian, &pk.maxPlayerCount); err != nil {
		return fmt.Errorf("read max player count: %w", err)
	}
	if err := binary.Read(a, binary.LittleEndian, &pk.editorWorld); err != nil {
		return fmt.Errorf("read editor world: %w", err)
	}
	if err := binary.Read(a, binary.LittleEndian, &pk.transportLayer); err != nil {
		return fmt.Errorf("read transport layer: %w", err)
	}

	return nil
}
func (pk *responsePacket) write(buf *bytes.Buffer) {
	a := &bytes.Buffer{}

	_ = binary.Write(a, binary.LittleEndian, pk.version)
	_ = binary.Write(a, binary.LittleEndian, uint8(len(pk.serverName)))
	a.WriteString(pk.serverName)
	_ = binary.Write(a, binary.LittleEndian, uint8(len(pk.levelName)))
	a.WriteString(pk.levelName)
	_ = binary.Write(a, binary.LittleEndian, pk.gameType)
	_ = binary.Write(a, binary.LittleEndian, pk.playerCount)
	_ = binary.Write(a, binary.LittleEndian, pk.maxPlayerCount)
	_ = binary.Write(a, binary.LittleEndian, pk.editorWorld)
	_ = binary.Write(a, binary.LittleEndian, pk.transportLayer)

	applicationData := make([]byte, hex.EncodedLen(a.Len()))
	hex.Encode(applicationData, a.Bytes())
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(applicationData)))
	_, _ = buf.Write(applicationData)
}

type messagePacket struct {
	recipientID uint64
	data        string
}

func (*messagePacket) id() uint16 { return idMessage }
func (pk *messagePacket) read(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.LittleEndian, &pk.recipientID); err != nil {
		return fmt.Errorf("read recipient ID: %w", err)
	}
	var length uint32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return fmt.Errorf("read data length: %w", err)
	}
	pk.data = string(buf.Next(int(length)))
	return nil
}
func (pk *messagePacket) write(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.recipientID)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(pk.data)))
	_, _ = buf.WriteString(pk.data)
}

type packetHeader struct {
	length   uint16
	packetID uint16
	senderID uint64
}

func (h *packetHeader) write(w io.Writer) {
	_ = binary.Write(w, binary.LittleEndian, h.length)
	_ = binary.Write(w, binary.LittleEndian, h.packetID)
	_ = binary.Write(w, binary.LittleEndian, h.senderID)
	_, _ = w.Write(make([]byte, 8))
}

func (h *packetHeader) read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.length); err != nil {
		return fmt.Errorf("read length: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &h.packetID); err != nil {
		return fmt.Errorf("read packet ID: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &h.senderID); err != nil {
		return fmt.Errorf("read sender ID: %w", err)
	}
	if n, err := r.Read(make([]byte, 8)); err != nil || n != 8 {
		return fmt.Errorf("discard padding: %w", err)
	}
	return nil
}

var key = sha256.Sum256(binary.LittleEndian.AppendUint64(nil, 0xdeadbeef))

func encryptECB(src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("make block: %w", err)
	}
	mode := ecb.NewECBEncrypter(block)
	p := padding.NewPkcs7Padding(mode.BlockSize())
	src, err = p.Pad(src)
	if err != nil {
		return nil, fmt.Errorf("pad: %w", err)
	}
	dst := make([]byte, len(src))
	mode.CryptBlocks(dst, src)
	return dst, nil
}

func decryptECB(src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("make block: %w", err)
	}
	mode := ecb.NewECBDecrypter(block)
	dst := make([]byte, len(src))
	mode.CryptBlocks(dst, src)
	p := padding.NewPkcs7Padding(mode.BlockSize())
	dst, err = p.Unpad(dst)
	if err != nil {
		return nil, fmt.Errorf("unpad: %w", err)
	}
	return dst, nil
}
