package discovery

import (
	"context"
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"log/slog"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ListenConfig struct {
	NetworkID        uint64
	BroadcastAddress net.Addr
	Log              *slog.Logger
}

func (conf ListenConfig) Listen(network string, addr string) (*Listener, error) {
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	if conf.NetworkID == 0 {
		conf.NetworkID = rand.Uint64()
	}
	conn, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		conn: conn,

		conf: conf,

		signals: make(chan *nethernet.Signal),

		addresses: make(map[uint64]net.Addr),

		closed: make(chan struct{}),
	}
	go l.listen()

	if conf.BroadcastAddress == nil {
		conf.BroadcastAddress, err = broadcastAddress(conn.LocalAddr())
		if err != nil {
			conf.Log.Error("error resolving address for broadcast: local rooms may not be returned")
		}
	}
	if conf.BroadcastAddress != nil {
		go l.broadcast(conf.BroadcastAddress)
	}

	return l, nil
}

func broadcastAddress(addr net.Addr) (net.Addr, error) {
	switch addr := addr.(type) {
	case *net.UDPAddr:
		ip := addr.IP.To4()
		if ip == nil {
			return nil, fmt.Errorf("address %q is not an IPv4 address; broadcasting on non-IPv4 address is currently not supported", addr)
		}
		return &net.UDPAddr{
			IP:   broadcastIP4(ip),
			Port: addr.Port,
		}, nil
	case *net.TCPAddr:
		ip := addr.IP.To4()
		if ip == nil {
			return nil, fmt.Errorf("address %q is not an IPv4 address; broadcasting on non-IPv4 address is currently not supported", addr)
		}
		return &net.TCPAddr{
			IP:   broadcastIP4(ip),
			Port: addr.Port,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported address type %T", addr)
	}
}

func broadcastIP4(ip net.IP) net.IP {
	mask := ip.DefaultMask()
	bcast := make(net.IP, len(ip))
	for i := 0; i < len(bcast); i++ {
		bcast[i] = ip[i] | ^mask[i]
	}
	return bcast
}

type Listener struct {
	conn net.PacketConn

	conf ListenConfig

	pongData atomic.Pointer[[]byte]

	signals chan *nethernet.Signal

	addressesMu sync.RWMutex
	addresses   map[uint64]net.Addr

	responsesMu sync.RWMutex
	responses   map[uint64][]byte

	closed chan struct{}
	once   sync.Once
}

func (l *Listener) ReadSignal(cancel <-chan struct{}) (*nethernet.Signal, error) {
	select {
	case <-cancel:
		return nil, context.Canceled
	case <-l.closed:
		return nil, net.ErrClosed
	case signal := <-l.signals:
		return signal, nil
	}
}

func (l *Listener) WriteSignal(signal *nethernet.Signal) error {
	select {
	case <-l.closed:
		return net.ErrClosed
	default:
		l.addressesMu.RLock()
		addr, ok := l.addresses[signal.NetworkID]
		l.addressesMu.RUnlock()

		if !ok {
			return fmt.Errorf("no address found for network ID: %d", signal.NetworkID)
		}

		_, err := l.write(Marshal(&MessagePacket{
			RecipientID: signal.NetworkID,
			Data:        signal.String(),
		}, l.conf.NetworkID), addr)
		return err
	}
}

func (l *Listener) Credentials() (*nethernet.Credentials, error) {
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
		return nil, nil
	}
}

func (l *Listener) listen() {
	for {
		b := make([]byte, 1024)
		n, addr, err := l.conn.ReadFrom(b)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				l.conf.Log.Error("error reading from conn", internal.ErrAttr(err))
			}
			close(l.closed)
			return
		}
		if err := l.handlePacket(b[:n], addr); err != nil {
			l.conf.Log.Error("error handling packet", internal.ErrAttr(err), "from", addr)
		}
	}
}

func (l *Listener) handlePacket(data []byte, addr net.Addr) error {
	pk, senderID, err := Unmarshal(data)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	if senderID == l.conf.NetworkID {
		return nil
	}

	l.addressesMu.Lock()
	l.addresses[senderID] = addr
	l.addressesMu.Unlock()

	switch pk := pk.(type) {
	case *RequestPacket:
		err = l.handleRequest(addr)
	case *ResponsePacket:
		err = l.handleResponse(pk, senderID)
	case *MessagePacket:
		err = l.handleMessage(pk, senderID, addr)
	default:
		err = fmt.Errorf("unknown packet: %T", pk)
	}

	return err
}

func (l *Listener) handleRequest(addr net.Addr) error {
	data := l.pongData.Load()
	if data == nil {
		return errors.New("application data not set yet")
	}
	if _, err := l.write(Marshal(&ResponsePacket{
		ApplicationData: *data,
	}, l.conf.NetworkID), addr); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

func (l *Listener) handleResponse(pk *ResponsePacket, senderID uint64) error {
	l.responsesMu.Lock()
	l.responses[senderID] = pk.ApplicationData
	l.responsesMu.Unlock()

	return nil
}

func (l *Listener) handleMessage(pk *MessagePacket, senderID uint64, addr net.Addr) error {
	if pk.Data == "Ping" {
		return nil
	}

	signal := &nethernet.Signal{}
	if err := signal.UnmarshalText([]byte(pk.Data)); err != nil {
		return fmt.Errorf("decode signal: %w", err)
	}
	signal.NetworkID = senderID
	l.signals <- signal

	return nil
}

func (l *Listener) ServerData(d *ServerData) {
	b, _ := d.MarshalBinary()
	l.PongData(b)
}

func (l *Listener) PongData(b []byte) { l.pongData.Store(&b) }

func (l *Listener) Close() (err error) {
	l.once.Do(func() {
		err = l.conn.Close()
	})
	return err
}

func (l *Listener) broadcast(addr net.Addr) {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	request := Marshal(&RequestPacket{}, l.conf.NetworkID)

	for {
		select {
		case <-l.closed:
			return
		case <-ticker.C:
			if _, err := l.conn.WriteTo(request, addr); err != nil {
				if !errors.Is(err, net.ErrClosed) {
					l.conf.Log.Error("error broadcasting request", internal.ErrAttr(err))
				}
				return
			}
		}
	}
}

func (l *Listener) write(b []byte, addr net.Addr) (n int, err error) {
	localIP, remoteIP := l.ip(addr), l.ip(l.conn.LocalAddr())
	if localIP != nil && remoteIP != nil && localIP.Equal(remoteIP) {
		bcast, err := broadcastAddress(addr)
		if err != nil {
			l.conf.Log.Error("error resolving broadcast address", slog.Any("addr", addr), internal.ErrAttr(err))
		} else {
			addr = bcast
		}
	}
	return l.conn.WriteTo(b, addr)
}

func (l *Listener) ip(addr net.Addr) net.IP {
	switch addr := addr.(type) {
	case *net.UDPAddr:
		return addr.IP
	case *net.TCPAddr:
		return addr.IP
	default:
		return nil
	}
}
