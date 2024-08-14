package nethernet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pion/ice/v3"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"log/slog"
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

	closeCandidateReceived sync.Once     // A sync.Once that closes candidateReceived only once.
	candidateReceived      chan struct{} // Notifies that a first candidate is received from the other end, and the Conn is ready to start its transports.

	reliable, unreliable *webrtc.DataChannel // ReliableDataChannel and UnreliableDataChannel

	packets chan []byte

	buf              *bytes.Buffer
	promisedSegments uint8

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

func (c *Conn) Write(b []byte) (n int, err error) {
	select {
	case <-c.closed:
		return n, net.ErrClosed
	default:
		// TODO: Clean up...
		if len(b) > maxMessageSize {
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
					return n, fmt.Errorf("write segment #%d: %w", segments, err)
				}
				n += len(frag)
			}

			// TODO
			if segments != 0 {
				panic("minecraft/nethernet: Conn: segments != 0")
			}
		} else {
			if err := c.reliable.Send(append([]byte{0}, b...)); err != nil {
				return n, err
			}
			n = len(b)
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

// LocalAddr currently returns a dummy address.
// TODO: Return something a valid address.
func (c *Conn) LocalAddr() net.Addr {
	dummy, _ := net.ResolveUDPAddr("udp", ":19132")
	return dummy
}

// RemoteAddr currently returns a dummy address.
// TODO: Return something a valid address.
func (c *Conn) RemoteAddr() net.Addr {
	dummy, _ := net.ResolveUDPAddr("udp", ":19132")
	return dummy
}

func (c *Conn) Close() error {
	errs := make([]error, 0, 3)
	if c.reliable != nil {
		if err := c.reliable.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.unreliable != nil {
		if err := c.unreliable.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(append(errs, c.closeTransports())...)
}

func (c *Conn) handleTransports() {
	c.reliable.OnMessage(func(msg webrtc.DataChannelMessage) {
		if err := c.handleMessage(msg.Data); err != nil {
			c.log.Error("error handling remote message", internal.ErrAttr(err))
		}
	})

	c.ice.OnConnectionStateChange(func(state webrtc.ICETransportState) {
		switch state {
		case webrtc.ICETransportStateClosed, webrtc.ICETransportStateDisconnected, webrtc.ICETransportStateFailed:
			_ = c.closeTransports() // We need to make sure that all transports has been closed
		default:
		}
	})
	c.dtls.OnStateChange(func(state webrtc.DTLSTransportState) {
		switch state {
		case webrtc.DTLSTransportStateClosed, webrtc.DTLSTransportStateFailed:
			_ = c.closeTransports() // We need to make sure that all transports has been closed
		default:
		}
	})
}

func (c *Conn) closeTransports() (err error) {
	c.once.Do(func() {
		errs := make([]error, 0, 3)

		if err := c.sctp.Stop(); err != nil {
			errs = append(errs, err)
		}
		if err := c.dtls.Stop(); err != nil {
			errs = append(errs, err)
		}
		if err := c.ice.Stop(); err != nil {
			errs = append(errs, err)
		}
		err = errors.Join(errs...)
		close(c.closed)
	})
	return err
}

func (c *Conn) handleSignal(signal *Signal) error {
	if signal.Type == SignalTypeCandidate {
		candidate, err := ice.UnmarshalCandidate(signal.Data)
		if err != nil {
			return fmt.Errorf("decode candidate: %w", err)
		}
		protocol, err := webrtc.NewICEProtocol(candidate.NetworkType().NetworkShort())
		if err != nil {
			return fmt.Errorf("parse ICE protocol: %w", err)
		}
		i := &webrtc.ICECandidate{
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

		if err := c.ice.AddRemoteCandidate(i); err != nil {
			return fmt.Errorf("add remote candidate: %w", err)
		}

		c.closeCandidateReceived.Do(func() {
			close(c.candidateReceived)
		})
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

func newConn(ice *webrtc.ICETransport, dtls *webrtc.DTLSTransport, sctp *webrtc.SCTPTransport, d *description, log *slog.Logger, id, networkID uint64) *Conn {
	return &Conn{
		ice:  ice,
		dtls: dtls,
		sctp: sctp,

		remote: d,

		candidateReceived: make(chan struct{}, 1),

		packets: make(chan []byte),
		buf:     bytes.NewBuffer(nil),

		closed: make(chan struct{}, 1),

		log: log.With(slog.Group("connection",
			slog.Uint64("id", id),
			slog.Uint64("networkID", networkID))),

		id:        id,
		networkID: networkID,
	}
}
