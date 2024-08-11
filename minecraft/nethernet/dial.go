package nethernet

import (
	"context"
	"errors"
	"fmt"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"math/rand"
	"strconv"
)

type Dialer struct {
	NetworkID, ConnectionID uint64
	API                     *webrtc.API
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
		return nil, fmt.Errorf("gather local ICE candidates: %w", err)
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
		fingerprint := dtlsParams.Fingerprints[0]
		sctpCapabilities := sctp.GetCapabilities()

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
						{Key: "fingerprint", Value: fmt.Sprintf("%s %s", fingerprint.Algorithm, fingerprint.Value)},
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
	}
	return nil, nil // TODO: Implement a way to dial.
}
