package nethernet

import (
	"context"
	"errors"
	"fmt"
	"github.com/pion/logging"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"sync"
)

// TODO: Under in construction!

type ListenConfig struct {
	Log *slog.Logger
	API *webrtc.API
}

func (conf ListenConfig) Listen(networkID uint64, signaling Signaling) (*Listener, error) {
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	if conf.API == nil {
		var (
			setting webrtc.SettingEngine
			factory = logging.NewDefaultLoggerFactory()
		)
		factory.DefaultLogLevel = logging.LogLevelDebug
		setting.LoggerFactory = factory

		conf.API = webrtc.NewAPI(webrtc.WithSettingEngine(setting))
	}
	l := &Listener{
		conf:      conf,
		signaling: signaling,
		networkID: networkID,

		incoming: make(chan *Conn),
	}
	var cancel context.CancelCauseFunc
	l.ctx, cancel = context.WithCancelCause(context.Background())
	go l.startListening(cancel)
	return l, nil
}

type Listener struct {
	conf ListenConfig

	ctx       context.Context
	signaling Signaling
	networkID uint64

	connections sync.Map

	incoming chan *Conn
	once     sync.Once
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case <-l.ctx.Done():
		return nil, context.Cause(l.ctx)
	case conn := <-l.incoming:
		return conn, nil
	}
}

// Addr currently returns a dummy address.
// TODO: Return something a valid address.
func (l *Listener) Addr() net.Addr {
	dummy, _ := net.ResolveUDPAddr("udp", ":19132")
	return dummy
}

// ID returns the network ID of listener.
func (l *Listener) ID() int64 { return int64(l.networkID) }

// PongData is currently a stub.
// TODO: Do something.
func (l *Listener) PongData([]byte) {}

func (l *Listener) startListening(cancel context.CancelCauseFunc) {
	for {
		signal, err := l.signaling.ReadSignal()
		if err != nil {
			cancel(err)
			close(l.incoming)
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
		candidates []*webrtc.ICECandidate
		// Notifies that gathering for local candidates has finished.
		gatherFinished = make(chan struct{})
	)
	gatherer.OnLocalCandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			close(gatherFinished)
			return
		}
		candidates = append(candidates, candidate)
	})
	if err := gatherer.Gather(); err != nil {
		return wrapSignalError(fmt.Errorf("gather local candidates: %w", err), ErrorCodeFailedToCreatePeerConnection)
	}

	select {
	case <-l.ctx.Done():
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
		d = &sdp.SessionDescription{
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
						Media: "application",
						Port: sdp.RangedPort{
							Value: 9,
						},
						Protos:  []string{"UDP", "DTLS", "SCTP"},
						Formats: []string{"webrtc-datachannel"},
					},
					ConnectionInformation: &sdp.ConnectionInformation{
						NetworkType: "IN",
						AddressType: "IP4",
						Address: &sdp.Address{
							Address: "0.0.0.0",
						},
					},
					Attributes: []sdp.Attribute{
						{Key: "ice-ufrag", Value: iceParams.UsernameFragment},
						{Key: "ice-pwd", Value: iceParams.Password},
						{Key: "ice-options", Value: "trickle"},
						{Key: "fingerprint", Value: fmt.Sprintf("%s %s",
							dtlsParams.Fingerprints[0].Algorithm,
							dtlsParams.Fingerprints[0].Value,
						)},
						{Key: "setup", Value: "active"},
						{Key: "mid", Value: "0"},
						{Key: "sctp-port", Value: "5000"},
						{Key: "max-message-size", Value: strconv.FormatUint(uint64(sctpCapabilities.MaxMessageSize), 10)},
					},
				},
			},
		}
		answer, err := d.Marshal()
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

		c := newConn(ice, dtls, sctp, desc, l.conf.Log, signal.ConnectionID, signal.NetworkID)

		l.connections.Store(signal.ConnectionID, c)
		go l.handleConn(c)

		return nil
	}
}

func (l *Listener) handleConn(conn *Conn) {
	select {
	case <-l.ctx.Done():
		// Quit the goroutine when the listener closes.
		return
	case <-conn.candidateReceived:
		conn.log.Debug("received first candidate")
		if err := l.startTransports(conn); err != nil {
			conn.log.Error("error starting transports", internal.ErrAttr(err))
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
	case <-l.ctx.Done():
		return l.ctx.Err()
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

func (l *Listener) Close() error {
	l.once.Do(func() {

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
