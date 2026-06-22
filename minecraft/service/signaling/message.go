package signaling

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Message represents a single frame exchanged over a Conn.
// It encapsulates the details of the message including its type, sender,
// recipient, and the actual data contained within the message.
type Message struct {
	// Type is the type of the message. It can be one of the MessageType
	// constants defined in this package.
	Type int `json:"Type"`

	// From indicates the sender of the message. It can be either a fixed string
	// 'Server' or the network ID from which the message was received. It is included
	// only received from the server.
	From string `json:"From,omitempty"`

	// To specifies the recipient of the message, which is the ID of remote network
	// to which the message is being sent. It is included only sent from client.
	To string `json:"To,omitempty"`

	// Data contains the actual payload of the message, which holds the data being transmitted.
	// It is optional and may be omitted if no data is being sent.
	Data string `json:"Message,omitempty"`

	// ID is the unique ID associated to each Message sent over the signaling service connection.
	// It is used to track the status of a single message delivery with its associated ID.
	ID uuid.UUID `json:"MessageId,omitempty"`
}

// Error represents the data included in a Message of MessageTypeError received from the server.
//
// It is notified by Conn to its registered nethernet.Notifier to negotiator to notify an error
// has occurred while sending a signal.
type Error struct {
	// Code is the code of the Error. It indicates the type of error and is may be one
	// of the constants below.
	Code int `json:"Code"`
	// Message represents the Error in a string.
	Message string `json:"Message"`
}

// Error returns a string representing code and message of the Error. It implements an error.
func (err *Error) Error() string {
	b := &strings.Builder{}
	b.WriteString("service/signaling: code ")
	b.WriteString(strconv.Itoa(err.Code))
	if err.Message != "" {
		b.WriteByte(':')
		b.WriteByte(' ')
		b.WriteString(strconv.Quote(err.Message))
	}
	return b.String()
}

const (
	// ErrorCodePlayerNotFound indicates that the remote network ID specified
	// in [Message.To] is not found. It is currently unused, and ErrorCodeDeliveryFailure
	// is instead used for this purpose.
	ErrorCodePlayerNotFound = iota + 1
	// ErrorCodeDeliveryFailure indicates that the signaling service couldn't deliver
	// the message sent by the Conn to a remote network ID. It is also returned when
	// the recipient ID for the remote network is no longer valid.
	ErrorCodeDeliveryFailure
)

// MessageTypeError is sent by server to notify that an error occurred
// in Conn. Messages of MessageTypeError usually contain a JSON string
// represented by Error.
const MessageTypeError = 0

const (
	// MessageTypePing is sent by client to ping the server at some interval.
	// Messages of MessageTypePing usually does not contain data.
	MessageTypePing = iota // RequestType::Ping

	// MessageTypeSignal is sent by both server and client to notify or send
	// a signal to a remote network. Messages of MessageTypeSignal usually
	// contain a data represented in [nethernet.Signal]. A Conn allows sending
	// a [nethernet.Signal] to a network using [Conn.Signal].
	MessageTypeSignal // RequestType::WebRTC

	// MessageTypeCredentials is sent by server to update credentials used for
	// gathering ICE candidates using STUN or TURN server specified on the fields.
	// Messages of MessageTypeCredentials usually contain a JSON string represented
	// in [nethernet.Credentials]. When received from the server, [Message.From] must
	// be 'Server' to ensure correct credentials are used.
	MessageTypeCredentials // RequestType::Credentials

	// MessageTypeAccepted is sent by server to track the delivery status of the message
	// with its associated ID in signaling service. Messages of MessageTypeAccepted usually
	// contain a JSON string represented in MessageStatus with [MessageStatus.AcceptedOn]
	// being the time that the message is accepted on the signaling service. It is mainly
	// used for measuring message performance for telemetry purposes and does not affect
	// the behavior.
	MessageTypeAccepted

	// MessageTypeDelivered is sent by server to track the delivery status of the message
	// with its associated ID in signaling service. Messages of MessageTypeDelivered usually
	// contain a JSON string represented in MessageStatus with [MessageStatus.DeliveredOn]
	// being the time that the message is delivered to the remote network. It is used for
	// measuring message performance and also to indicate a successful message delivery.
	MessageTypeDelivered
)

// MessageStatus tracks the message delivery status with its associated ID in signaling service.
// It is contained as a JSON string in Messages with MessageTypeAccepted or MessageTypeDelivered.
// Only messages with MessageTypeDelivered are used in Conn to complete Conn.Signal calls.
type MessageStatus struct {
	// MessageID is the ID associated to the message.
	MessageID uuid.UUID `json:"MessageId"`
	// SenderID is the network ID that has sent the message.
	// It is only included in Messages with MessageTypeDelivered.
	SenderID string `json:"ToPlayerId"`
	// DeliveredOn is the time that the message has delivered
	// to the remote network. It is only included in Messages
	// with MessageTypeDelivered.
	DeliveredOn time.Time `json:"DeliveredOn"`
	// AcceptedOn is the time that the message has accepted
	// on the signaling service. It is only included in Messages
	// with MessageTypeAccepted.
	AcceptedOn time.Time `json:"AcceptedOn"`
}
