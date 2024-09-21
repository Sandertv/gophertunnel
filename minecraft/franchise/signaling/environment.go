package signaling

// Environment represents an environment configuration for establishing a Conn with signaling services in Dialer.
// It contains fields necessary for connecting to the appropriate URL, and can be obtained easily from a [franchise.Discovery]
// (which can be also obtained from [franchise.Discover] with a specific version) using [franchise.Discovery.Environment] with
// an *Environment.
//
// Example usage:
//
//	discovery, err := franchise.Discover(protocol.CurrentVersion)
//	if err != nil {
//		panic(err)
//	}
//
//	environment := new(Environment)
//	if err := discovery.Environment(environment); err != nil {
//		panic(err)
//	}
//
// // Use discovery and environment for further uses.
type Environment struct {
	// ServiceURI is the URI of the service where connections should be directed.
	// It is the base URL used for dialing a WebSocket connection of a Conn.
	ServiceURI string `json:"serviceUri"`
	// StunURI is the URI of a STUN server available to connect. It seems unused as it is always
	// provided in a [nethernet.Credentials] received from a Conn.
	StunURI string `json:"stunUri"`
	// TurnURI is the URI of a TURN server available to connect. It seems unused as it is always
	// provided in a credentials received from a Conn.
	TurnURI string `json:"turnUri"`
}

// EnvironmentName implements a [franchise.Environment] so that may be obtained using [franchise.Discovery.Environment].
func (env *Environment) EnvironmentName() string {
	return "signaling"
}
