package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/df-mc/go-playfab/v2"
)

// TokenSource returns an implementation of TokenSource, which subsequently supplies the token
// by either newly requesting or refreshing an existing, cached token. The given [playfab.Client]
// will be used for logging into Bedrock Edition's network services with the user's PlayFab account.
func (e *AuthorizationEnvironment) TokenSource(client *playfab.Client, config TokenConfig) TokenSource {
	defaultUserConfig(&config.User)
	defaultDeviceConfig(e, &config.Device)

	return &tokenSource{
		client: client,
		env:    e,
		config: config,
	}
}

// tokenSource is an implementation of TokenSource that supplies tokens by
// reusing existing tokens whenever possible.
type tokenSource struct {
	client *playfab.Client
	env    *AuthorizationEnvironment
	config TokenConfig

	token *Token
	mu    sync.Mutex
}

// ServiceToken supplies a token by either re-using an already requested token, or
// requesting or renewing the existing token with a valid PlayFab session ticket.
func (s *tokenSource) ServiceToken(ctx context.Context) (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token != nil && s.token.Valid() {
		return s.token, nil
	}

	// PlayFab Client reuses the cached session ticket from last login if valid.
	// Otherwise, it refreshes the session ticket (approximately 24 hours after login).
	ticket, err := s.client.SessionTicket(ctx)
	if err != nil {
		return nil, fmt.Errorf("request session ticket: %w", err)
	}
	s.config.User.TokenType = TokenTypePlayFab
	s.config.User.Token = ticket

	if s.token == nil {
		s.token, err = s.env.Token(ctx, s.config)
		if err != nil {
			return nil, fmt.Errorf("request: %w", err)
		}
	} else if !s.token.Valid() {
		s.token, err = s.env.Renew(ctx, s.token, s.config.User)
		if err != nil {
			return nil, fmt.Errorf("renew: %w", err)
		}
	}
	return s.token, nil
}
