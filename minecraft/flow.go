package minecraft

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// LoginFlowHandler customises the Minecraft login flow of a connection. Each
// function field is called for the packet received at that step of the login
// flow. A nil field means that step is not handled by the login flow: packets
// of that type are treated as unexpected and are deferred to the user (or
// passed to HandleUnexpectedPacket, if set).
//
// The connection logic advances the login flow between steps based on which
// handlers are set: after a handled step, the packets of the following steps
// whose handler is non-nil are expected next. It is therefore possible to
// replace or short-circuit individual steps of the flow while keeping the rest
// intact.
//
// Returning an error from any handler aborts the login and closes the
// connection.
type LoginFlowHandler struct {
	// HandleUnexpectedPacket is called for a packet received during the login
	// flow that was not expected at the current step. If nil, the packet is
	// deferred and returned by the next call to (*Conn).ReadPacket as if the
	// login had already completed.
	HandleUnexpectedPacket func(*Conn, packet.Packet) error

	// Server handlers are called for packets received by a Conn obtained from
	// a Listener.
	HandleClientCacheStatus           func(*Conn, *packet.ClientCacheStatus) error
	HandleResourcePackClientResponse  func(*Conn, *packet.ResourcePackClientResponse) error
	HandleResourcePackChunkRequest    func(*Conn, *packet.ResourcePackChunkRequest) error
	HandleRequestChunkRadius          func(*Conn, *packet.RequestChunkRadius) error
	HandleSetLocalPlayerAsInitialised func(*Conn, *packet.SetLocalPlayerAsInitialised) error

	// Client handlers are called for packets received by a Conn obtained from
	// a Dialer.
	HandleResourcePacksInfo     func(*Conn, *packet.ResourcePacksInfo) error
	HandleResourcePackDataInfo  func(*Conn, *packet.ResourcePackDataInfo) error
	HandleResourcePackChunkData func(*Conn, *packet.ResourcePackChunkData) error
	HandleResourcePackStack     func(*Conn, *packet.ResourcePackStack) error
	HandleDimensionData         func(*Conn, *packet.DimensionData) error
	HandleStartGame             func(*Conn, *packet.StartGame) error
	HandleItemRegistry          func(*Conn, *packet.ItemRegistry) error
	HandleChunkRadiusUpdated    func(*Conn, *packet.ChunkRadiusUpdated) error
}

// handles checks if the flow has a handler set for the packet with the ID passed.
func (flow *LoginFlowHandler) handles(id uint32) bool {
	switch id {
	case packet.IDRequestNetworkSettings:
		return true
	case packet.IDLogin:
		return true
	case packet.IDClientToServerHandshake:
		return true
	case packet.IDClientCacheStatus:
		return flow.HandleClientCacheStatus != nil
	case packet.IDResourcePackClientResponse:
		return flow.HandleResourcePackClientResponse != nil
	case packet.IDResourcePackChunkRequest:
		return flow.HandleResourcePackChunkRequest != nil
	case packet.IDRequestChunkRadius:
		return flow.HandleRequestChunkRadius != nil
	case packet.IDSetLocalPlayerAsInitialised:
		return flow.HandleSetLocalPlayerAsInitialised != nil
	case packet.IDNetworkSettings:
		return true
	case packet.IDServerToClientHandshake:
		return true
	case packet.IDPlayStatus:
		return true
	case packet.IDResourcePacksInfo:
		return flow.HandleResourcePacksInfo != nil
	case packet.IDResourcePackDataInfo:
		return flow.HandleResourcePackDataInfo != nil
	case packet.IDResourcePackChunkData:
		return flow.HandleResourcePackChunkData != nil
	case packet.IDResourcePackStack:
		return flow.HandleResourcePackStack != nil
	case packet.IDDimensionData:
		return flow.HandleDimensionData != nil
	case packet.IDStartGame:
		return flow.HandleStartGame != nil
	case packet.IDItemRegistry:
		return flow.HandleItemRegistry != nil
	case packet.IDChunkRadiusUpdated:
		return flow.HandleChunkRadiusUpdated != nil
	}
	return false
}
