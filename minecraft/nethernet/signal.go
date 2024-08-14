package nethernet

import (
	"bytes"
	"fmt"
	"github.com/pion/webrtc/v4"
	"strconv"
	"strings"
)

// TODO: Improve documentations written in my poor English ;-;
// TODO: We need an implementation of Signaling that is usable under multiple goroutines.
// I want to use one Signaling connection in both Listen and Dial, because it should work.

type Signaling interface {
	ReadSignal() (*Signal, error)
	WriteSignal(signal *Signal) error

	// Credentials will currently block until a credentials has received from the signaling service. This is usually
	// present in WebSocket signaling connection. A nil *Credentials may be returned if no credentials or
	// the implementation is not capable to do that.
	Credentials() (*Credentials, error)
}

const (
	// SignalTypeOffer is sent by client to request a connection to the remote host. Signals that have
	// SignalTypeOffer usually has a data of local description of its connection.
	SignalTypeOffer = "CONNECTREQUEST"
	// SignalTypeAnswer is sent by server to respond to Signals that have SignalTypeOffer. Signals that
	// have SignalTypeAnswer usually has a data of local description of the host.
	SignalTypeAnswer = "CONNECTRESPONSE"
	// SignalTypeCandidate is sent by both server and client to notify a local candidate to
	// remote connection. This is usually sent after SignalTypeOffer or SignalTypeAnswer by server/client.
	// Signals that have SignalTypeCandidate usually has a data of local candidate gathered with additional
	// credentials received from the Signaling implementation.
	SignalTypeCandidate = "CANDIDATEADD"
	// SignalTypeError is sent by both server and client to notify an error has occurred.
	// Signals that have SignalTypeError has a Data of the code of error occurred, which is listed
	// on the following constants.
	SignalTypeError = "CONNECTERROR"
)

type Signal struct {
	// Type is the type of Signal. It is one of the constants defined above.
	Type string
	// ConnectionID is the unique ID of the connection that has sent the Signal.
	// It is encoded in String as a second segment to identify a connection uniquely.
	ConnectionID uint64
	// Data is the actual data of the Signal.
	Data string

	// NetworkID is used internally by the implementations of Signaling type
	// to reference a remote network with a number.
	NetworkID uint64
}

func (s *Signal) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Signal) UnmarshalText(b []byte) (err error) {
	segments := bytes.SplitN(b, []byte{' '}, 3)
	if len(segments) != 3 {
		return fmt.Errorf("unexpected segmentations: %d", len(segments))
	}
	s.Type = string(segments[0])
	s.ConnectionID, err = strconv.ParseUint(string(segments[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("parse ConnectionID: %w", err)
	}
	s.Data = string(segments[2])
	return nil
}

func (s *Signal) String() string {
	b := &strings.Builder{}
	b.WriteString(s.Type)
	b.WriteByte(' ')
	b.WriteString(strconv.FormatUint(s.ConnectionID, 10))
	b.WriteByte(' ')
	b.WriteString(s.Data)
	return b.String()
}

func formatICECandidate(id int, candidate *webrtc.ICECandidate, iceParams webrtc.ICEParameters) string {
	b := &strings.Builder{}
	b.WriteString("candidate:")
	b.WriteString(candidate.Foundation)
	b.WriteByte(' ')
	b.WriteByte('1')
	b.WriteByte(' ')
	b.WriteString("udp")
	b.WriteByte(' ')
	b.WriteString(strconv.FormatUint(uint64(candidate.Priority), 10))
	b.WriteByte(' ')
	b.WriteString(candidate.Address)
	b.WriteByte(' ')
	b.WriteString(strconv.FormatUint(uint64(candidate.Port), 10))
	b.WriteByte(' ')
	b.WriteString("typ")
	b.WriteByte(' ')
	b.WriteString(candidate.Typ.String())
	b.WriteByte(' ')
	if candidate.Typ == webrtc.ICECandidateTypeRelay || candidate.Typ == webrtc.ICECandidateTypeSrflx {
		b.WriteString("raddr")
		b.WriteByte(' ')
		b.WriteString(candidate.RelatedAddress)
		b.WriteByte(' ')
		b.WriteString("rport")
		b.WriteByte(' ')
		b.WriteString(strconv.FormatUint(uint64(candidate.RelatedPort), 10))
		b.WriteByte(' ')
	}
	b.WriteString("generation")
	b.WriteByte(' ')
	b.WriteByte('0')
	b.WriteByte(' ')
	b.WriteString("ufrag")
	b.WriteByte(' ')
	b.WriteString(iceParams.UsernameFragment)
	b.WriteByte(' ')
	b.WriteString("network-id")
	b.WriteByte(' ')
	b.WriteString(strconv.Itoa(id))
	b.WriteByte(' ')
	b.WriteString("network-cost")
	b.WriteByte(' ')
	b.WriteByte('0')
	return b.String()
}

// These constants are sent as a data of Signal with SignalTypeError, to notify an error to the remote connection.
// TODO: These codes has been extracted from dedicated server (v1.21.2). We need to properly write a documentation for these constants.
const (
	ErrorCodeNone = iota
	ErrorCodeDestinationNotLoggedIn
	ErrorCodeNegotiationTimeout
	ErrorCodeWrongTransportVersion
	ErrorCodeFailedToCreatePeerConnection
	ErrorCodeICE
	ErrorCodeConnectRequest
	ErrorCodeConnectResponse
	ErrorCodeCandidateAdd
	ErrorCodeInactivityTimeout
	ErrorCodeFailedToCreateOffer
	ErrorCodeFailedToCreateAnswer
	ErrorCodeFailedToSetLocalDescription
	ErrorCodeFailedToSetRemoteDescription
	ErrorCodeNegotiationTimeoutWaitingForResponse
	ErrorCodeNegotiationTimeoutWaitingForAccept
	ErrorCodeIncomingConnectionIgnored
	ErrorCodeSignalingParsingFailure
	ErrorCodeSignalingUnknownError
	ErrorCodeSignalingUnicastMessageDeliveryFailed
	ErrorCodeSignalingBroadcastDeliveryFailed
	ErrorCodeSignalingMessageDeliveryFailed
	ErrorCodeSignalingTurnAuthFailed
	ErrorCodeSignalingFallbackToBestEffortDelivery
	ErrorCodeNoSignalingChannel
	ErrorCodeNotLoggedIn
	ErrorCodeSignalingFailedToSend
)
