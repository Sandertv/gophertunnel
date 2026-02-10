package minecraft

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/service"
	"golang.org/x/oauth2"
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
// underlying [oauth2.TokenSource] to sign in to the PlayFab account with Xbox Live.
type multiplayerTokenSource struct {
	oauth2.TokenSource
}

// MultiplayerToken issues a multiplayer token using the underlying [oauth2.TokenSource].
func (s *multiplayerTokenSource) MultiplayerToken(ctx context.Context, key *ecdsa.PublicKey) (string, error) {
	env, err := authEnv(ctx)
	if err != nil {
		return "", fmt.Errorf("obtain environment for auth: %w", err)
	}
	return env.MultiplayerToken(ctx, env.TokenSource(ctx, s.TokenSource, service.TokenConfig{}), key)
}

// CachedMultiplayerTokenSource is an implementation of [MultiplayerTokenSource] that reuses a single
// [service.TokenSource] across calls. This allows caching of the Minecraft authorization service token
// (the result of /api/v1.0/session/start) and prevents repeated PlayFab logins when dialing multiple times.
//
// The first context passed to MultiplayerToken is used to initialise internal state and will be used by the
// cached TokenSource for subsequent refreshes. Callers should ensure this context has the desired values set,
// such as oauth2.HTTPClient and auth.WithXBLTokenCache. Cancellation/deadlines are stripped from that context
// so it remains usable after the first dial returns.
type CachedMultiplayerTokenSource struct {
	oauth2.TokenSource

	// TokenConfig is the configuration used when creating the cached [service.TokenSource].
	// If left as the zero value, defaults will be applied by the service package.
	TokenConfig service.TokenConfig

	mu sync.Mutex
	// env is the Minecraft authorization environment (service endpoints + OpenID issuer config).
	// It is used to issue multiplayer tokens.
	env *service.AuthorizationEnvironment
	// src is the cached service.TokenSource created from env. It caches the authorization service session token
	// (/api/v1.0/session/start) and handles renewal, which avoids repeated PlayFab logins across dials.
	//
	// Note: src.Token() has no context parameter, so it captures a context at construction time.
	src service.TokenSource
}

// NewCachedMultiplayerTokenSource returns a [CachedMultiplayerTokenSource] wrapping src.
func NewCachedMultiplayerTokenSource(src oauth2.TokenSource, config service.TokenConfig) *CachedMultiplayerTokenSource {
	return &CachedMultiplayerTokenSource{
		TokenSource: src,
		TokenConfig: config,
	}
}

// MultiplayerToken returns a multiplayer token issued by the Minecraft authorization service.
// The token is still issued per call (it is key-bound), but the underlying service token used to request it
// is cached across calls.
func (s *CachedMultiplayerTokenSource) MultiplayerToken(ctx context.Context, key *ecdsa.PublicKey) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.tryInit(ctx); err != nil {
		return "", err
	}
	return s.env.MultiplayerToken(ctx, s.src, key)
}

// tryInit initialises the CachedMultiplayerTokenSource if it is not already initialised.
// It resolves the Minecraft authorization environment (service endpoints + OpenID issuer config),
// and constructs a cached [service.TokenSource] which will later cache the auth service session token and renewal.
func (s *CachedMultiplayerTokenSource) tryInit(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.src != nil {
		return nil
	}

	env, err := authEnv(ctx)
	if err != nil {
		return fmt.Errorf("obtain environment for auth: %w", err)
	}

	// service.TokenSource.Token() has no context parameter, so it must capture one.
	// Strip cancellation/deadlines so the cached TokenSource remains usable after DialContext returns.
	baseCtx := context.WithoutCancel(ctx)
	s.env = env
	s.src = env.TokenSource(baseCtx, s.TokenSource, s.TokenConfig)
	return nil
}
