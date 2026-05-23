package signaling

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Environment represents an environment for the signaling service.
type Environment struct {
	// ServiceURI is the base endpoint URL for the signaling service.
	// The scheme is typically 'wss://'.
	ServiceURI *url.URL `json:"serviceUri"`
	// TurnURI is the 'turn://' URI for Microsoft's TURN server. NetherNet
	// connections use the TURN server embedded in
	// [github.com/df-mc/go-nethernet.Credentials]
	// for actual WebRTC negotiation, so the purpose of this field is unknown.
	TurnURI string `json:"turnUri"`
	// StunURI is the 'stun://' URI for Microsoft's STUN server. NetherNet
	// connections use the STUN server embedded in
	// [github.com/df-mc/go-nethernet.Credentials]
	// for actual WebRTC negotiation, so the purpose of this field is unknown.
	StunURI string `json:"stunUri"`
}

// ServiceName always returns 'signaling' as the name of the service environment.
// It implements [service.Environment] so it can be derived using [service.Discovery.Environment].
func (e *Environment) ServiceName() string {
	return "signaling"
}

// UnmarshalJSON decodes the ServiceURI field to string then parses as URL.
// Other URI fields such as TurnURI is not validated as it is not used in
// the actual WebRTC negotiation for NetherNet connections.
func (e *Environment) UnmarshalJSON(b []byte) (err error) {
	type Alias Environment
	data := struct {
		*Alias
		ServiceURI string `json:"serviceUri"`
	}{Alias: (*Alias)(e)}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	e.ServiceURI, err = url.Parse(data.ServiceURI)
	if err != nil {
		return fmt.Errorf("parse service URI: %w", err)
	}
	return nil
}
