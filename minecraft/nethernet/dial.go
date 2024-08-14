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
		var (
			setting webrtc.SettingEngine
			factory = logging.NewDefaultLoggerFactory()
		)
		factory.DefaultLogLevel = logging.LogLevelDebug
		setting.LoggerFactory = factory

		d.API = webrtc.NewAPI(webrtc.WithSettingEngine(setting))
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}
	credentials, err := signaling.Credentials()
	if err != nil {
		return nil, wrapSignalError(fmt.Errorf("obtain credentials: %w", err), ErrorCodeFailedToCreatePeerConnection)
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
		return nil, wrapSignalError(fmt.Errorf("create ICE gatherer: %w", err), ErrorCodeFailedToCreatePeerConnection)
	}

	var (
		candidates     []*webrtc.ICECandidate
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
		return nil, wrapSignalError(fmt.Errorf("gather local candidates: %w", err), ErrorCodeFailedToCreatePeerConnection)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-gatherFinished:
		ice := d.API.NewICETransport(gatherer)
		dtls, err := d.API.NewDTLSTransport(ice, nil)
		if err != nil {
			return nil, wrapSignalError(fmt.Errorf("create DTLS transport: %w", err), ErrorCodeFailedToCreatePeerConnection)
		}
		sctp := d.API.NewSCTPTransport(dtls)

		iceParams, err := ice.GetLocalParameters()
		if err != nil {
			return nil, wrapSignalError(fmt.Errorf("obtain local ICE parameters: %w", err), ErrorCodeFailedToCreatePeerConnection)
		}
		dtlsParams, err := dtls.GetLocalParameters()
		if err != nil {
			return nil, wrapSignalError(fmt.Errorf("obtain local DTLS parameters: %w", err), ErrorCodeFailedToCreateAnswer)
		}
		if len(dtlsParams.Fingerprints) == 0 {
			return nil, wrapSignalError(errors.New("local DTLS parameters has no fingerprints"), ErrorCodeFailedToCreateAnswer)
		}
		sctpCapabilities := sctp.GetCapabilities()

		// Encode an offer using the local parameters!
		description := &sdp.SessionDescription{
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
						{Key: "ice-ufrag", Value: iceParams.UsernameFragment},
						{Key: "ice-pwd", Value: iceParams.Password},
						{Key: "ice-options", Value: "trickle"},
						{Key: "fingerprint", Value: fmt.Sprintf("%s %s",
							dtlsParams.Fingerprints[0].Algorithm,
							dtlsParams.Fingerprints[0].Value,
						)},
						{Key: "setup", Value: "actpass"},
						{Key: "mid", Value: "0"},
						{Key: "sctp-port", Value: "5000"},
						{Key: "max-message-size", Value: strconv.FormatUint(uint64(sctpCapabilities.MaxMessageSize), 10)},
					},
				},
			},
		}

		offer, err := description.Marshal()
		if err != nil {
			return nil, wrapSignalError(fmt.Errorf("encode offer: %w", err), ErrorCodeFailedToCreateAnswer)
		}
		if err := signaling.WriteSignal(&Signal{
			Type:         SignalTypeOffer,
			Data:         string(offer),
			ConnectionID: d.ConnectionID,
			NetworkID:    networkID,
		}); err != nil {
			// I don't think the error code will be signaled back to the remote connection, but just in case.
			return nil, wrapSignalError(fmt.Errorf("signal offer: %w", err), ErrorCodeSignalingFailedToSend)
		}
		for i, candidate := range candidates {
			if err := signaling.WriteSignal(&Signal{
				Type:         SignalTypeCandidate,
				Data:         formatICECandidate(i, candidate, iceParams),
				ConnectionID: d.ConnectionID,
				NetworkID:    networkID,
			}); err != nil {
				// I don't think the error code will be signaled back to the remote connection, but just in case.
				return nil, wrapSignalError(fmt.Errorf("signal candidate: %w", err), ErrorCodeSignalingFailedToSend)
			}
		}

		signals := make(chan *Signal)
		go d.notifySignals(ctx, d.ConnectionID, networkID, signaling, signals)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case signal := <-signals:
			if signal.Type != SignalTypeAnswer {
				return nil, fmt.Errorf("received signal for non-answer: %s", signal.String())
			}

			description = &sdp.SessionDescription{}
			if err := description.UnmarshalString(signal.Data); err != nil {
				return nil, fmt.Errorf("decode answer: %w", err)
			}
			desc, err := parseDescription(description)
			if err != nil {
				return nil, fmt.Errorf("parse offer: %w", err)
			}

			c := newConn(ice, dtls, sctp, desc, d.Log, d.ConnectionID, networkID)
			go d.handleConn(ctx, c, signals)

			select {
			case <-ctx.Done():
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
			if signal.Type == SignalTypeCandidate {
				if err := conn.handleSignal(signal); err != nil {
					conn.log.Error("error handling signal", internal.ErrAttr(err))
				}
			}
		}
	}
}

func (d Dialer) notifySignals(ctx context.Context, id, networkID uint64, signaling Signaling, c chan<- *Signal) {
	for {
		if ctx.Err() != nil {
			return
		}
		signal, err := signaling.ReadSignal()
		if err != nil {
			d.Log.Error("error reading signal", internal.ErrAttr(err))
			return
		}
		if signal.ConnectionID != id || signal.NetworkID != networkID {
			d.Log.Error("unexpected connection ID or network ID", slog.Group("signal", signal))
			continue
		}
		c <- signal
	}
}
