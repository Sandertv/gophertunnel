package realms

import "strings"

// NetworkProtocol is the protocol type returned by the Realms API used to connect to a realm.
type NetworkProtocol string

const (
	NetworkProtocolDefault          NetworkProtocol = "DEFAULT"
	NetworkProtocolNetherNet        NetworkProtocol = "NETHERNET"
	NetworkProtocolNetherNetJSONRPC NetworkProtocol = "NETHERNET_JSONRPC"
)

// ParseNetworkProtocol converts a network protocol string to a NetworkProtocol value.
func ParseNetworkProtocol(protocol string) NetworkProtocol {
	return NetworkProtocol(strings.ToUpper(strings.TrimSpace(protocol)))
}

// Valid reports whether the protocol is one of the known constants.
func (p NetworkProtocol) Valid() bool {
	switch ParseNetworkProtocol(string(p)) {
	case NetworkProtocolDefault, NetworkProtocolNetherNet, NetworkProtocolNetherNetJSONRPC:
		return true
	default:
		return false
	}
}
