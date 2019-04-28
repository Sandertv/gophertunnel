package protocol

const (
	// CurrentProtocol is the current protocol version for the version below.
	CurrentProtocol = 354
	// CurrentVersion is the current version of Minecraft as supported by the `packet` package.
	CurrentVersion = "1.11.1"
)

const (
	// IDLogin is the identifier for the Login packet.
	IDLogin = iota + 1
	// IDPlayStatus is the identifier for the PlayStatus packet.
	IDPlayStatus
	// IDServerToClientHandshake is the identifier for the ServerToClientHandshake packet.
	IDServerToClientHandshake
	// IDClientToServerHandshake is the identifier for the ClientToServerHandshake packet.
	IDClientToServerHandshake
	// IDDisconnect is the identifier for the Disconnect packet.
	IDDisconnect
)
