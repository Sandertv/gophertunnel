package signaling

import "encoding/json"

type Message struct {
	Type uint32 `json:"Type"`
	// From is either a unique ID of remote network, or a string "Server".
	From string      `json:"From,omitempty"`
	To   json.Number `json:"To,omitempty"`
	Data string      `json:"Message,omitempty"`
}

const (
	MessageTypeRequestPing uint32 = iota // RequestType::Ping
	MessageTypeSignal                    // RequestType::Message
	MessageTypeCredentials               // RequestType::TurnAuth
)
