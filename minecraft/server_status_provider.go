package minecraft

// ServerStatusProvider represents a value that is able to provide the status of a server, in specific its
// MOTD, the amount of online players and the player limit.
// These providers may be used to display different information in the server list. Although they overwrite
// the server name, maximum players and online players maintained by a Listener in the server list, these
// values are not changed and will still be used internally to check if players are able to be connected.
// Players will still be disconnected if the maximum player count as set in the MaximumPlayers field of a
// Listener is reached (unless MaximumPlayers is 0).
type ServerStatusProvider interface {
	// ServerStatus returns the server status which includes the MOTD/server name, amount of online players
	// and the amount of maximum players.
	ServerStatus() (motd string, onlinePlayers, maxPlayers int)
}
