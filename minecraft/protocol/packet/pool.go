package packet

// Pool is a map holding packets indexed by a packet ID.
type Pool map[uint32]Packet

// NewPool returns a new pool with all supported packets sent. Packets may be retrieved from it simply by
// indexing it with the packet ID.
func NewPool() Pool {
	return Pool{
		IDLogin:                      &Login{},
		IDPlayStatus:                 &PlayStatus{},
		IDServerToClientHandshake:    &ServerToClientHandshake{},
		IDClientToServerHandshake:    &ClientToServerHandshake{},
		IDDisconnect:                 &Disconnect{},
		IDResourcePacksInfo:          &ResourcePacksInfo{},
		IDResourcePackStack:          &ResourcePackStack{},
		IDResourcePackClientResponse: &ResourcePackClientResponse{},
		IDText:                       &Text{},
		// ...
		IDStartGame: &StartGame{},
		// ...
		IDRequestChunkRadius: &RequestChunkRadius{},
		IDChunkRadiusUpdated: &ChunkRadiusUpdated{},
		// ...
		IDResourcePackDataInfo:     &ResourcePackDataInfo{},
		IDResourcePackChunkData:    &ResourcePackChunkData{},
		IDResourcePackChunkRequest: &ResourcePackChunkRequest{},
		// ...
		IDSetLocalPlayerAsInitialised: &SetLocalPlayerAsInitialised{},
	}
}
