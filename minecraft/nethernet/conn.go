package nethernet

import (
	"errors"
	"fmt"
	"github.com/pion/ice/v3"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Conn struct {
	ice  *webrtc.ICETransport
	dtls *webrtc.DTLSTransport
	sctp *webrtc.SCTPTransport

	remote *description

	candidateReceived chan struct{} // Notifies that a first candidate is received from the other end, and the Conn is ready to start its transports.
	candidates        []webrtc.ICECandidate
	candidatesMu      sync.Mutex

	localCandidates []webrtc.ICECandidate
	localNetworkID  uint64

	reliable, unreliable *webrtc.DataChannel // ReliableDataChannel and UnreliableDataChannel

	packets chan []byte

	message *message

	once   sync.Once
	closed chan struct{}

	log *slog.Logger

	id, networkID uint64
}

func (c *Conn) Read(b []byte) (n int, err error) {
	select {
	case <-c.closed:
		return n, net.ErrClosed
	case pk := <-c.packets:
		return copy(b, pk), nil
	}
}

func (c *Conn) ReadPacket() ([]byte, error) {
	select {
	case <-c.closed:
		return nil, net.ErrClosed
	case pk := <-c.packets:
		return pk, nil
	}
}

func (c *Conn) Write(b []byte) (n int, err error) {
	select {
	case <-c.closed:
		return n, net.ErrClosed
	default:
		// TODO: Clean up...
		segments := uint8(len(b) / maxMessageSize)
		if len(b)%maxMessageSize != 0 {
			segments++ // If there's a remainder, we need an additional segment.
		}

		for i := 0; i < len(b); i += maxMessageSize {
			segments--

			end := i + maxMessageSize
			if end > len(b) {
				end = len(b)
			}
			frag := b[i:end]
			if err := c.reliable.Send(append([]byte{segments}, frag...)); err != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					return n, net.ErrClosed
				}
				return n, fmt.Errorf("write segment #%d: %w", segments, err)
			}
			n += len(frag)
		}

		// TODO
		if segments != 0 {
			panic("minecraft/nethernet: Conn: segments != 0")
		}
		return n, nil
	}
}

func (*Conn) SetDeadline(time.Time) error {
	return errors.New("minecraft/nethernet: Conn: not implemented (yet)")
}

func (*Conn) SetReadDeadline(time.Time) error {
	return errors.New("minecraft/nethernet: Conn: not implemented (yet)")
}

func (*Conn) SetWriteDeadline(time.Time) error {
	return errors.New("minecraft/nethernet: Conn: not implemented (yet)")
}

func (c *Conn) LocalAddr() net.Addr {
	return &Addr{
		NetworkID:    c.localNetworkID,
		ConnectionID: c.id,
		Candidates:   c.localCandidates,
	}
}

func (c *Conn) RemoteAddr() net.Addr {
	c.candidatesMu.Lock()
	defer c.candidatesMu.Unlock()

	return &Addr{
		NetworkID:    c.networkID,
		ConnectionID: c.id,
		Candidates:   c.candidates,
	}
}

func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		close(c.closed)

		errs := make([]error, 0, 5)
		errs = append(errs, c.reliable.Close())
		errs = append(errs, c.unreliable.Close())
		errs = append(errs, c.sctp.Stop())
		errs = append(errs, c.dtls.Stop())
		errs = append(errs, c.ice.Stop())
		err = errors.Join(errs...)
	})
	return err
}

func (c *Conn) handleTransports() {
	c.reliable.OnMessage(func(msg webrtc.DataChannelMessage) {
		if err := c.handleMessage(msg.Data); err != nil {
			c.log.Error("error handling remote message", internal.ErrAttr(err))
		}
	})

	c.reliable.OnClose(func() {
		c.Close()
	})

	c.unreliable.OnClose(func() {
		_ = c.Close()
	})

	c.ice.OnConnectionStateChange(func(state webrtc.ICETransportState) {
		switch state {
		case webrtc.ICETransportStateClosed, webrtc.ICETransportStateDisconnected, webrtc.ICETransportStateFailed:
			// This handler function itself is holding the lock, call Close in a goroutine.
			go c.Close() // We need to make sure that all transports has been closed
		default:
		}
	})
	c.dtls.OnStateChange(func(state webrtc.DTLSTransportState) {
		switch state {
		case webrtc.DTLSTransportStateClosed, webrtc.DTLSTransportStateFailed:
			// This handler function itself is holding the lock, call Close in a goroutine.
			go c.Close() // We need to make sure that all transports has been closed
		default:
		}
	})
}

func (c *Conn) handleSignal(signal *Signal) error {
	switch signal.Type {
	case SignalTypeCandidate:
		candidate, err := ice.UnmarshalCandidate(signal.Data)
		if err != nil {
			return fmt.Errorf("decode candidate: %w", err)
		}
		protocol, err := webrtc.NewICEProtocol(candidate.NetworkType().NetworkShort())
		if err != nil {
			return fmt.Errorf("parse ICE protocol: %w", err)
		}
		i := webrtc.ICECandidate{
			Foundation: candidate.Foundation(),
			Priority:   candidate.Priority(),
			Address:    candidate.Address(),
			Protocol:   protocol,
			Port:       uint16(candidate.Port()),
			Component:  candidate.Component(),
			Typ:        webrtc.ICECandidateType(candidate.Type()),
			TCPType:    candidate.TCPType().String(),
		}

		if r := candidate.RelatedAddress(); r != nil {
			i.RelatedAddress, i.RelatedPort = r.Address, uint16(r.Port)
		}

		if err := c.ice.AddRemoteCandidate(&i); err != nil {
			return fmt.Errorf("add remote candidate: %w", err)
		}

		c.candidatesMu.Lock()
		if len(c.candidates) == 0 {
			close(c.candidateReceived)
		}
		c.candidates = append(c.candidates, i)
		c.candidatesMu.Unlock()
	case SignalTypeError:
		code, err := strconv.ParseUint(signal.Data, 10, 32)
		if err != nil {
			return fmt.Errorf("parse error code: %w", err)
		}
		c.log.Error("connection failed with error", slog.Uint64("code", code))
		if err := c.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}
	}
	return nil
}

const maxMessageSize = 10000

func parseDescription(d *sdp.SessionDescription) (*description, error) {
	if len(d.MediaDescriptions) != 1 {
		return nil, fmt.Errorf("unexpected number of media descriptions: %d, expected 1", len(d.MediaDescriptions))
	}
	m := d.MediaDescriptions[0]

	ufrag, ok := m.Attribute("ice-ufrag")
	if !ok {
		return nil, errors.New("missing ice-ufrag attribute")
	}
	pwd, ok := m.Attribute("ice-pwd")
	if !ok {
		return nil, errors.New("missing ice-pwd attribute")
	}

	attr, ok := m.Attribute("fingerprint")
	if !ok {
		return nil, errors.New("missing fingerprint attribute")
	}
	fingerprint := strings.Split(attr, " ")
	if len(fingerprint) != 2 {
		return nil, fmt.Errorf("invalid fingerprint: %s", attr)
	}
	fingerprintAlgorithm, fingerprintValue := fingerprint[0], fingerprint[1]

	attr, ok = m.Attribute("max-message-size")
	if !ok {
		return nil, errors.New("missing max-message-size attribute")
	}
	maxMessageSize, err := strconv.ParseUint(attr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse max-message-size attribute as uint32: %w", err)
	}

	return &description{
		ice: webrtc.ICEParameters{
			UsernameFragment: ufrag,
			Password:         pwd,
		},
		dtls: webrtc.DTLSParameters{
			Fingerprints: []webrtc.DTLSFingerprint{
				{
					Algorithm: fingerprintAlgorithm,
					Value:     fingerprintValue,
				},
			},
		},
		sctp: webrtc.SCTPCapabilities{
			MaxMessageSize: uint32(maxMessageSize),
		},
	}, nil
}

type description struct {
	ice  webrtc.ICEParameters
	dtls webrtc.DTLSParameters
	sctp webrtc.SCTPCapabilities
}

func (desc description) encode() ([]byte, error) {
	d := &sdp.SessionDescription{
		Version: 0x0,
		Origin: sdp.Origin{
			Username:       "-",
			SessionID:      rand.Uint64(),
			SessionVersion: 0x2,
			NetworkType:    "IN",
			AddressType:    "IP4",
			UnicastAddress: "127.0.0.1",
		},
		SessionName: "-",
		TimeDescriptions: []sdp.TimeDescription{
			{},
		},
		Attributes: []sdp.Attribute{
			{Key: "group", Value: "BUNDLE 0"},
			{Key: "extmap-allow-mixed", Value: ""},
			{Key: "msid-semantic", Value: " WMS"},
		},
		MediaDescriptions: []*sdp.MediaDescription{
			{
				MediaName: sdp.MediaName{
					Media:   "application",
					Port:    sdp.RangedPort{Value: 9},
					Protos:  []string{"UDP", "DTLS", "SCTP"},
					Formats: []string{"webrtc-datachannel"},
				},
				ConnectionInformation: &sdp.ConnectionInformation{
					NetworkType: "IN",
					AddressType: "IP4",
					Address:     &sdp.Address{Address: "0.0.0.0"},
				},
				Attributes: []sdp.Attribute{
					{Key: "ice-ufrag", Value: desc.ice.UsernameFragment},
					{Key: "ice-pwd", Value: desc.ice.Password},
					{Key: "ice-options", Value: "trickle"},
					{Key: "fingerprint", Value: fmt.Sprintf("%s %s",
						desc.dtls.Fingerprints[0].Algorithm,
						desc.dtls.Fingerprints[0].Value,
					)},
					desc.setupAttribute(),
					{Key: "mid", Value: "0"},
					{Key: "sctp-port", Value: "5000"},
					{Key: "max-message-size", Value: strconv.FormatUint(uint64(desc.sctp.MaxMessageSize), 10)},
				},
			},
		},
	}
	return d.Marshal()
}

func (desc description) setupAttribute() sdp.Attribute {
	attr := sdp.Attribute{Key: "setup"}
	if desc.dtls.Role == webrtc.DTLSRoleServer {
		attr.Value = "actpass"
	} else {
		attr.Value = "active"
	}
	return attr
}

func newConn(ice *webrtc.ICETransport, dtls *webrtc.DTLSTransport, sctp *webrtc.SCTPTransport, d *description, log *slog.Logger, id, networkID, localNetworkID uint64, candidates []webrtc.ICECandidate) *Conn {
	return &Conn{
		ice:  ice,
		dtls: dtls,
		sctp: sctp,

		remote: d,

		candidateReceived: make(chan struct{}, 1),

		localNetworkID:  localNetworkID,
		localCandidates: candidates,

		packets: make(chan []byte),

		message: &message{},

		closed: make(chan struct{}, 1),

		log: log.With(slog.Group("connection",
			slog.Uint64("id", id),
			slog.Uint64("networkID", networkID))),

		id:        id,
		networkID: networkID,
	}
}
