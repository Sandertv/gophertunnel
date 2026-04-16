package realms

import "strings"

// NetworkProtocol is the protocol hint returned by the Realms API for a join target.
type NetworkProtocol string

const (
	NetworkProtocolDefault          NetworkProtocol = "DEFAULT"
	NetworkProtocolNetherNet        NetworkProtocol = "NETHERNET"
	NetworkProtocolNetherNetJSONRPC NetworkProtocol = "NETHERNET_JSONRPC"
)

// ParseNetworkProtocol normalizes a Realm network protocol string.
func ParseNetworkProtocol(protocol string) NetworkProtocol {
	return NetworkProtocol(strings.ToUpper(strings.TrimSpace(protocol)))
}

// Valid reports whether the protocol is one of the known Realms values.
func (p NetworkProtocol) Valid() bool {
	switch ParseNetworkProtocol(string(p)) {
	case NetworkProtocolDefault, NetworkProtocolNetherNet, NetworkProtocolNetherNetJSONRPC:
		return true
	default:
		return false
	}
}
