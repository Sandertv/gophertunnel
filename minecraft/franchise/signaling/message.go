package signaling

import (
	"strconv"
	"strings"
)

// Message represents a message sent or received over a Conn.
// It encapsulates the details of the message including its type, sender,
// recipient, and the actual data contained within the message.
type Message struct {
	// Type indicates the type of the Message. It corresponds to one of the
	// constants defined below.
	Type int `json:"Type"`

	// From indicates the sender of the message. It can be either a fixed string
	// 'Server' or the network ID from which the message was received. It is included
	// only received from the server.
	From string `json:"From,omitempty"`

	// To specifies the recipient of the message, which is the ID of remote network
	// to which the message is being sent. It is included only sent from client.
	To uint64 `json:"To,omitempty"`

	// Data contains the actual payload of the message, which holds the data being transmitted.
	// It is optional and may be omitted if no data is being sent.
	Data string `json:"Message,omitempty"`
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
	b.WriteString("franchise/signaling: code ")
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
	// in [Message.To] is not found. Meaning that a [nethernet.Signal] signaled
	// to the server is no longer valid.
	ErrorCodePlayerNotFound = 1
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
)
