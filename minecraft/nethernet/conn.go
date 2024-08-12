package nethernet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pion/ice/v3"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	ice  *webrtc.ICETransport
	dtls *webrtc.DTLSTransport
	sctp *webrtc.SCTPTransport

	// Remote parameters for starting ICE, DTLS, and SCTP transport.
	iceParams        webrtc.ICEParameters
	dtlsParams       webrtc.DTLSParameters
	sctpCapabilities webrtc.SCTPCapabilities

	candidatesReceived atomic.Uint32 // Total amount of candidates received.
	api                *webrtc.API   // WebRTC API to create new data channels on SCTP transport.

	reliable, unreliable *webrtc.DataChannel // ReliableDataChannel and UnreliableDataChannel
	ready                chan struct{}       // Notifies when reliable and unreliable are ready.

	packets chan []byte

	buf              *bytes.Buffer
	promisedSegments uint8

	log *slog.Logger

	id, networkID uint64
}

func (c *Conn) Read(b []byte) (n int, err error) {
	select {
	//case <-c.closed:
	//	return n, net.ErrClosed
	case pk := <-c.packets:
		return copy(b, pk), nil
	}
}

func (c *Conn) Write(b []byte) (n int, err error) {
	select {
	//case <-c.closed:
	//	return n, net.ErrClosed
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
	errs := make([]error, 0, 5)
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

	if err := c.sctp.Stop(); err != nil {
		errs = append(errs, err)
	}
	if err := c.dtls.Stop(); err != nil {
		errs = append(errs, err)
	}
	if err := c.ice.Stop(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (c *Conn) startTransports() error {
	c.log.Debug("starting ICE transport")
	iceRole := webrtc.ICERoleControlled
	if err := c.ice.Start(nil, c.iceParams, &iceRole); err != nil {
		return fmt.Errorf("start ICE transport: %w", err)
	}

	c.log.Debug("starting DTLS transport")
	c.dtlsParams.Role = webrtc.DTLSRoleServer
	if err := c.dtls.Start(c.dtlsParams); err != nil {
		return fmt.Errorf("start DTLS transport: %w", err)
	}
	c.log.Debug("starting SCTP transport")

	var once sync.Once
	c.sctp.OnDataChannelOpened(func(channel *webrtc.DataChannel) {
		switch channel.Label() {
		case "ReliableDataChannel":
			c.reliable = channel
		case "UnreliableDataChannel":
			c.unreliable = channel
		}
		if c.reliable != nil && c.unreliable != nil {
			once.Do(func() {
				close(c.ready)
			})
		}
	})
	if err := c.sctp.Start(c.sctpCapabilities); err != nil {
		return fmt.Errorf("start SCTP transport: %w", err)
	}

	<-c.ready
	c.reliable.OnMessage(c.handleRemoteMessage)
	return nil
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

		if c.candidatesReceived.Add(1) == 1 {
			c.log.Debug("received first candidate, starting transports")
			go func() {
				if err := c.startTransports(); err != nil {
					c.log.Error("error starting transports", internal.ErrAttr(err))
				}
			}()
		}
	}
	return nil
}

const maxMessageSize = 10000
