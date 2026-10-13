package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Chat is sent by both the client and the server to send player chat routed through Mojang Managed Chat. It
// carries a Chat Service signature that receiving clients verify.
type Chat struct {
	// SignedChat is the signed chat message. It must be relayed unmodified by the server.
	SignedChat protocol.SignedChat
	// Formatting is optional server decoration rendered around the name and the message. It is not covered
	// by the signature.
	Formatting protocol.Optional[protocol.ChatFormatting]
}

// ID ...
func (*Chat) ID() uint32 {
	return IDChat
}

func (pk *Chat) Marshal(io protocol.IO) {
	protocol.Single(io, &pk.SignedChat)
	protocol.OptionalMarshaler(io, &pk.Formatting)
}
