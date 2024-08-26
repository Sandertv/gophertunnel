package nethernet

import (
	"context"
	"errors"
	"fmt"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"github.com/sandertv/gophertunnel/minecraft/nethernet/internal"
	"log/slog"
	"math/rand"
	"strconv"
)

type Dialer struct {
	NetworkID, ConnectionID uint64
	API                     *webrtc.API
	Log                     *slog.Logger
}

func (d Dialer) DialContext(ctx context.Context, networkID uint64, signaling Signaling) (*Conn, error) {
	if d.NetworkID == 0 {
		d.NetworkID = rand.Uint64()
	}
	if d.ConnectionID == 0 {
		d.ConnectionID = rand.Uint64()
	}
	if d.API == nil {
		d.API = webrtc.NewAPI()
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}
	credentials, err := signaling.Credentials()
	if err != nil {
		return nil, fmt.Errorf("obtain credentials: %w", err)
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
	gatherer, err := d.API.NewICEGatherer(gatherOptions)
	if err != nil {
		return nil, fmt.Errorf("create ICE gatherer: %w", err)
	}

	var (
		candidates     []webrtc.ICECandidate
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
		return nil, fmt.Errorf("gather local candidates: %w", err)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-gatherFinished:
		ice := d.API.NewICETransport(gatherer)
		dtls, err := d.API.NewDTLSTransport(ice, nil)
		if err != nil {
			return nil, fmt.Errorf("create DTLS transport: %w", err)
		}
		sctp := d.API.NewSCTPTransport(dtls)

		iceParams, err := ice.GetLocalParameters()
		if err != nil {
			return nil, fmt.Errorf("obtain local ICE parameters: %w", err)
		}
		dtlsParams, err := dtls.GetLocalParameters()
		if err != nil {
			return nil, fmt.Errorf("obtain local DTLS parameters: %w", err)
		}
		if len(dtlsParams.Fingerprints) == 0 {
			return nil, errors.New("local DTLS parameters has no fingerprints")
		}
		sctpCapabilities := sctp.GetCapabilities()

		dtlsParams.Role = webrtc.DTLSRoleServer

		// Encode an offer using the local parameters!
		offer, err := description{
			ice:  iceParams,
			dtls: dtlsParams,
			sctp: sctpCapabilities,
		}.encode()
		if err != nil {
			return nil, fmt.Errorf("encode offer: %w", err)
		}
		if err := signaling.WriteSignal(&Signal{
			Type:         SignalTypeOffer,
			Data:         string(offer),
			ConnectionID: d.ConnectionID,
			NetworkID:    networkID,
		}); err != nil {
			return nil, fmt.Errorf("signal offer: %w", err)
		}
		for i, candidate := range candidates {
			if err := signaling.WriteSignal(&Signal{
				Type:         SignalTypeCandidate,
				Data:         formatICECandidate(i, candidate, iceParams),
				ConnectionID: d.ConnectionID,
				NetworkID:    networkID,
			}); err != nil {
				return nil, fmt.Errorf("signal candidate: %w", err)
			}
		}

		signals := make(chan *Signal)
		go d.notifySignals(ctx, d.ConnectionID, networkID, signaling, signals)

		select {
		case <-ctx.Done():
			if errors.Is(err, context.DeadlineExceeded) {
				d.signalError(signaling, networkID, ErrorCodeNegotiationTimeoutWaitingForResponse)
			}
			return nil, ctx.Err()
		case signal := <-signals:
			if signal.Type != SignalTypeAnswer {
				d.signalError(signaling, networkID, ErrorCodeIncomingConnectionIgnored)
				return nil, fmt.Errorf("received signal for non-answer: %s", signal.String())
			}

			s := &sdp.SessionDescription{}
			if err := s.UnmarshalString(signal.Data); err != nil {
				d.signalError(signaling, networkID, ErrorCodeFailedToSetRemoteDescription)
				return nil, fmt.Errorf("decode answer: %w", err)
			}
			desc, err := parseDescription(s)
			if err != nil {
				d.signalError(signaling, networkID, ErrorCodeFailedToSetRemoteDescription)
				return nil, fmt.Errorf("parse offer: %w", err)
			}

			c := newConn(ice, dtls, sctp, desc, d.Log, d.ConnectionID, networkID, d.NetworkID, candidates, nil)
			go d.handleConn(ctx, c, signals)

			select {
			case <-ctx.Done():
				if errors.Is(err, context.DeadlineExceeded) {
					d.signalError(signaling, networkID, ErrorCodeInactivityTimeout)
				}
				return nil, ctx.Err()
			case <-c.candidateReceived:
				c.log.Debug("received first candidate")
				if err := d.startTransports(c); err != nil {
					return nil, fmt.Errorf("start transports: %w", err)
				}
				c.handleTransports()
				return c, nil
			}
		}
	}
}

func (d Dialer) signalError(signaling Signaling, networkID uint64, code int) {
	_ = signaling.WriteSignal(&Signal{
		Type:         SignalTypeError,
		Data:         strconv.Itoa(code),
		ConnectionID: d.ConnectionID,
		NetworkID:    networkID,
	})
}

func (d Dialer) startTransports(conn *Conn) error {
	conn.log.Debug("starting ICE transport as controller")
	iceRole := webrtc.ICERoleControlling
	if err := conn.ice.Start(nil, conn.remote.ice, &iceRole); err != nil {
		return fmt.Errorf("start ICE: %w", err)
	}

	conn.log.Debug("starting DTLS transport as client")
	dtlsParams := conn.remote.dtls
	dtlsParams.Role = webrtc.DTLSRoleClient
	if err := conn.dtls.Start(dtlsParams); err != nil {
		return fmt.Errorf("start DTLS: %w", err)
	}

	conn.log.Debug("starting SCTP transport")
	if err := conn.sctp.Start(conn.remote.sctp); err != nil {
		return fmt.Errorf("start SCTP: %w", err)
	}
	var err error
	conn.reliable, err = d.API.NewDataChannel(conn.sctp, &webrtc.DataChannelParameters{
		Label: "ReliableDataChannel",
	})
	if err != nil {
		return fmt.Errorf("create ReliableDataChannel: %w", err)
	}
	conn.unreliable, err = d.API.NewDataChannel(conn.sctp, &webrtc.DataChannelParameters{
		Label:   "UnreliableDataChannel",
		Ordered: false,
	})
	if err != nil {
		return fmt.Errorf("create UnreliableDataChannel: %w", err)
	}
	return nil
}

func (d Dialer) handleConn(ctx context.Context, conn *Conn, signals <-chan *Signal) {
	for {
		select {
		case <-ctx.Done():
			return
		case signal := <-signals:
			switch signal.Type {
			case SignalTypeCandidate, SignalTypeError:
				if err := conn.handleSignal(signal); err != nil {
					conn.log.Error("error handling signal", internal.ErrAttr(err))
				}
			}
		}
	}
}

func (d Dialer) notifySignals(ctx context.Context, id, networkID uint64, signaling Signaling, c chan<- *Signal) {
	for {
		signal, err := signaling.ReadSignal(ctx.Done())
		if err != nil {
			if !errors.Is(err, ErrSignalingCanceled) {
				d.Log.Error("error reading signal", internal.ErrAttr(err))
			}
			return
		}
		if signal.ConnectionID != id || signal.NetworkID != networkID {
			d.Log.Error("unexpected connection ID or network ID", slog.Group("signal", signal))
			continue
		}
		c <- signal
	}
}
