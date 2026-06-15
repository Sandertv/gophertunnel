package minecraft

import (
	"context"
	"crypto/ecdsa"

	"github.com/sandertv/gophertunnel/minecraft/service"
)

// MultiplayerTokenSource supplies a multiplayer token issued by the Minecraft authorization
// service, which is newly introduced in 1.21.100.
//
// The token is key-bound (it includes the client's public key in the 'cpk' claim), so callers should
// expect to request it per connection key.
type MultiplayerTokenSource interface {
	// MultiplayerToken issues a JWT token to be used for OpenID authentication with
	// multiplayer servers. The token must contain the public key in the 'cpk' claim in
	// order for the server to verify client data with the same key.
	MultiplayerToken(ctx context.Context, key *ecdsa.PublicKey) (jwt string, err error)
}

// multiplayerTokenSource is an implementation of MultiplayerTokenSource used by default, which uses the
// underlying [service.TokenSource] to log in to the Minecraft: Bedrock Edition's network services.
type multiplayerTokenSource struct {
	// env is the environment used for requesting multiplayer tokens.
	env *service.AuthorizationEnvironment
	// src is the [service.TokenSource] used to log in to the network services.
	// It is typically created from [service.AuthorizationEnvironment.TokenSource].
	src service.TokenSource
}

// MultiplayerToken issues a multiplayer token using the underlying [service.TokenSource].
func (s *multiplayerTokenSource) MultiplayerToken(ctx context.Context, key *ecdsa.PublicKey) (string, error) {
	return s.env.MultiplayerToken(ctx, s.src, key)
}
