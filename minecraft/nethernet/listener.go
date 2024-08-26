package nethernet

import (
	"errors"
	"fmt"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
)

type ListenConfig struct {
	Log *slog.Logger
	API *webrtc.API
}

func (conf ListenConfig) Listen(networkID uint64, signaling Signaling) (*Listener, error) {
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	if conf.API == nil {
		conf.API = webrtc.NewAPI()
	}
	l := &Listener{
		conf:      conf,
		signaling: signaling,
		networkID: networkID,

		incoming: make(chan *Conn),

		closed: make(chan struct{}),
	}
	go l.listen()
	return l, nil
}

type Listener struct {
	conf ListenConfig

	signaling Signaling
	networkID uint64

	connections sync.Map

	incoming chan *Conn

	closed chan struct{}
	once   sync.Once
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	case conn := <-l.incoming:
		return conn, nil
	}
}

func (l *Listener) Addr() net.Addr {
	return &Addr{NetworkID: l.networkID}
}

type Addr struct {
	ConnectionID uint64
	NetworkID    uint64
	Candidates   []webrtc.ICECandidate
}

func (addr *Addr) String() string {
	b := &strings.Builder{}
	b.WriteString(strconv.FormatUint(addr.NetworkID, 10))
	b.WriteByte(' ')
	if addr.ConnectionID != 0 {
		b.WriteByte('(')
		b.WriteString(strconv.FormatUint(addr.ConnectionID, 10))
		b.WriteByte(')')
	}
	return b.String()
}

func (addr *Addr) Network() string { return "nethernet" }

// ID returns the network ID of listener.
func (l *Listener) ID() int64 { return int64(l.networkID) }

// PongData is a stub.
func (l *Listener) PongData([]byte) {}

func (l *Listener) listen() {
	for {
		signal, err := l.signaling.ReadSignal(l.closed)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				l.conf.Log.Error("error reading signal", internal.ErrAttr(err))
			}
			_ = l.Close()
			return
		}

		// Um... It seems the game has a bug that doesn't even send an offer if joining worlds too many times.
		// This is not a bug of this code because you may not join any worlds if the bug has occurred.
		// Once the bug has occurred, you need to restart the game.
		l.conf.Log.Debug(signal.String())
		switch signal.Type {
		case SignalTypeOffer:
			err = l.handleOffer(signal)
		case SignalTypeCandidate:
			err = l.handleCandidate(signal)
		case SignalTypeError:
			err = l.handleError(signal)
		default:
			l.conf.Log.Debug("received signal for unknown type", "signal", signal)
		}
		if err != nil {
			var s *signalError
			if errors.As(err, &s) {
				// Additionally, we write a Signal back with SignalTypeError using the code wrapped on it.
				if err := l.signaling.WriteSignal(&Signal{
					Type:         SignalTypeError,
					ConnectionID: signal.ConnectionID,
					Data:         strconv.FormatUint(uint64(s.code), 10),
					NetworkID:    signal.NetworkID,
				}); err != nil {
					l.conf.Log.Error("error signaling error", internal.ErrAttr(err))
				}
			}
			l.conf.Log.Error("error handling signal", "signal", signal, internal.ErrAttr(err))
		}
	}
}

// handleOffer handles an incoming Signal of SignalTypeOffer. An answer will be
// encoded and the listener will prepare a connection for handling the signals incoming that has the same ID.
func (l *Listener) handleOffer(signal *Signal) error {
	d := &sdp.SessionDescription{}
	if err := d.UnmarshalString(signal.Data); err != nil {
		return wrapSignalError(fmt.Errorf("decode offer: %w", err), ErrorCodeFailedToSetRemoteDescription)
	}
	desc, err := parseDescription(d)
	if err != nil {
		return wrapSignalError(fmt.Errorf("parse offer: %w", err), ErrorCodeFailedToSetRemoteDescription)
	}

	credentials, err := l.signaling.Credentials()
	if err != nil {
		return wrapSignalError(fmt.Errorf("obtain credentials: %w", err), ErrorCodeSignalingTurnAuthFailed)
	}

	var gatherOptions webrtc.ICEGatherOptions
	if credentials != nil && len(credentials.ICEServers) > 0 {
		gatherOptions.ICEServers = make([]webrtc.ICEServer, len(credentials.ICEServers))
		for i, server := range credentials.ICEServers {
			gatherOptions.ICEServers[i] = webrtc.ICEServer{
				Username:       server.Username,
				Credential:     server.Password,
				CredentialType: webrtc.ICECredentialTypePassword,
				URLs:           server.URLs,
			}
		}
	}

	gatherer, err := l.conf.API.NewICEGatherer(gatherOptions)
	if err != nil {
		return wrapSignalError(fmt.Errorf("create ICE gatherer: %w", err), ErrorCodeFailedToCreatePeerConnection)
	}

	var (
		// Local candidates gathered by webrtc.ICEGatherer
		candidates []webrtc.ICECandidate
		// Notifies that gathering for local candidates has finished.
		gatherFinished = make(chan struct{})
	)
	gatherer.OnLocalCandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			close(gatherFinished)
			return
		}
		candidates = append(candidates, *candidate)
	})
	if err := gatherer.Gather(); err != nil {
		return wrapSignalError(fmt.Errorf("gather local candidates: %w", err), ErrorCodeFailedToCreatePeerConnection)
	}

	select {
	case <-l.closed:
		return nil
	case <-gatherFinished:
		ice := l.conf.API.NewICETransport(gatherer)
		dtls, err := l.conf.API.NewDTLSTransport(ice, nil)
		if err != nil {
			return wrapSignalError(fmt.Errorf("create DTLS transport: %w", err), ErrorCodeFailedToCreatePeerConnection)
		}
		sctp := l.conf.API.NewSCTPTransport(dtls)

		iceParams, err := ice.GetLocalParameters()
		if err != nil {
			return wrapSignalError(fmt.Errorf("obtain local ICE parameters: %w", err), ErrorCodeFailedToCreateAnswer)
		}
		dtlsParams, err := dtls.GetLocalParameters()
		if err != nil {
			return wrapSignalError(fmt.Errorf("obtain local DTLS parameters: %w", err), ErrorCodeFailedToCreateAnswer)
		}
		if len(dtlsParams.Fingerprints) == 0 {
			return wrapSignalError(errors.New("local DTLS parameters has no fingerprints"), ErrorCodeFailedToCreateAnswer)
		}
		sctpCapabilities := sctp.GetCapabilities()

		// Encode an answer using the local parameters!
		answer, err := description{
			ice:  iceParams,
			dtls: dtlsParams,
			sctp: sctpCapabilities,
		}.encode()
		if err != nil {
			return wrapSignalError(fmt.Errorf("encode answer: %w", err), ErrorCodeFailedToCreateAnswer)
		}

		if err := l.signaling.WriteSignal(&Signal{
			Type:         SignalTypeAnswer,
			ConnectionID: signal.ConnectionID,
			Data:         string(answer),
			NetworkID:    signal.NetworkID,
		}); err != nil {
			// I don't think the error code will be signaled back to the remote connection, but just in case.
			return wrapSignalError(fmt.Errorf("signal answer: %w", err), ErrorCodeSignalingFailedToSend)
		}
		for i, candidate := range candidates {
			if err := l.signaling.WriteSignal(&Signal{
				Type:         SignalTypeCandidate,
				ConnectionID: signal.ConnectionID,
				Data:         formatICECandidate(i, candidate, iceParams),
				NetworkID:    signal.NetworkID,
			}); err != nil {
				// I don't think the error code will be signaled back to the remote connection, but just in case.
				return wrapSignalError(fmt.Errorf("signal candidate: %w", err), ErrorCodeSignalingFailedToSend)
			}
		}

		c := newConn(ice, dtls, sctp, desc, l.conf.Log, signal.ConnectionID, signal.NetworkID, l.networkID, candidates, l)

		l.connections.Store(signal.ConnectionID, c)
		go l.handleConn(c)

		return nil
	}
}

func (l *Listener) handleClose(conn *Conn) {
	l.connections.Delete(conn.id)
}

func (l *Listener) handleConn(conn *Conn) {
	select {
	case <-l.closed:
		// Quit the goroutine when the listener closes.
		return
	case <-conn.candidateReceived:
		conn.log.Debug("received first candidate")
		if err := l.startTransports(conn); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				conn.log.Error("error starting transports", internal.ErrAttr(err))
			}
			return
		}
		conn.handleTransports()
		l.incoming <- conn
	}
}

func (l *Listener) startTransports(conn *Conn) error {
	conn.log.Debug("starting ICE transport as controlled")
	iceRole := webrtc.ICERoleControlled
	if err := conn.ice.Start(nil, conn.remote.ice, &iceRole); err != nil {
		return fmt.Errorf("start ICE: %w", err)
	}

	conn.log.Debug("starting DTLS transport as server")
	dtlsParams := conn.remote.dtls
	dtlsParams.Role = webrtc.DTLSRoleServer
	if err := conn.dtls.Start(dtlsParams); err != nil {
		return fmt.Errorf("start DTLS: %w", err)
	}

	conn.log.Debug("starting SCTP transport")
	var (
		once     = new(sync.Once)
		bothOpen = make(chan struct{}, 1)
	)
	conn.sctp.OnDataChannelOpened(func(channel *webrtc.DataChannel) {
		switch channel.Label() {
		case "ReliableDataChannel":
			conn.reliable = channel
		case "UnreliableDataChannel":
			conn.unreliable = channel
		}
		if conn.reliable != nil && conn.unreliable != nil {
			once.Do(func() {
				close(bothOpen)
			})
		}
	})
	if err := conn.sctp.Start(conn.remote.sctp); err != nil {
		return fmt.Errorf("start SCTP: %w", err)
	}

	select {
	case <-l.closed:
		return net.ErrClosed
	case <-bothOpen:
		return nil
	}
}

// handleCandidate handles an incoming Signal of SignalTypeCandidate. It looks up for a connection that has the same ID, and
// call the [Conn.handleSignal] method, which adds a remote candidate into its ICE transport.
func (l *Listener) handleCandidate(signal *Signal) error {
	conn, ok := l.connections.Load(signal.ConnectionID)
	if !ok {
		return fmt.Errorf("no connection found for ID %d", signal.ConnectionID)
	}
	return conn.(*Conn).handleSignal(signal)
}

// handleError handles an incoming Signal of SignalTypeError. It looks up for a connection that has the same ID, and
// call the [Conn.handleSignal] method, which parses the data into error code and closes the connection as failed.
func (l *Listener) handleError(signal *Signal) error {
	conn, ok := l.connections.Load(signal.ConnectionID)
	if !ok {
		return fmt.Errorf("no connection found for ID %d", signal.ConnectionID)
	}
	return conn.(*Conn).handleSignal(signal)
}

func (l *Listener) Close() error {
	l.once.Do(func() {
		close(l.closed)
		close(l.incoming)
	})
	return nil
}

type signalError struct {
	code       uint32
	underlying error
}

func (e *signalError) Error() string {
	return fmt.Sprintf("minecraft/nethernet: %s [signaling with code %d]", e.underlying, e.code)
}

func (e *signalError) Unwrap() error { return e.underlying }

func wrapSignalError(err error, code uint32) *signalError {
	return &signalError{code: code, underlying: err}
}
