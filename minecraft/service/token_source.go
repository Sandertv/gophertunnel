package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/df-mc/go-playfab"
	"github.com/df-mc/go-playfab/title"
	"github.com/df-mc/go-xsapi"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// TokenSource returns an implementation of TokenSource, which subsequently supplies the token
// by either newly requesting or refreshing an existing, cached token.
func (e *AuthorizationEnvironment) TokenSource(src oauth2.TokenSource, config TokenConfig) TokenSource {
	return &tokenSource{
		identity: playfab.XBLIdentityProvider{
			TokenSource: &xblTokenSource{
				TokenSource:  src,
				relyingParty: playfab.RelyingParty,
			},
		},
		env:    e,
		config: config,
	}
}

// tokenSource is an implementation of TokenSource that supplies tokens by
// reusing existing tokens whenever possible.
type tokenSource struct {
	identity playfab.IdentityProvider
	env      *AuthorizationEnvironment
	config   TokenConfig

	token *Token
	mu    sync.Mutex
}

// Token supplies a token by either re-using an already requested token, or
// requesting or renewing the existing token with a valid PlayFab session ticket.
func (s *tokenSource) Token() (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token != nil && s.token.Valid() {
		return s.token, nil
	}

	identity, err := s.identity.Login(playfab.LoginConfig{
		Title:         title.Title(s.env.PlayFabTitleID),
		CreateAccount: true,
	})
	if err != nil {
		return nil, fmt.Errorf("login playfab: %w", err)
	}
	s.config.User.TokenType = TokenTypePlayFab
	s.config.User.Token = identity.SessionTicket

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
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

// xblTokenSource is an implementation of [xsapi.TokenSource].
type xblTokenSource struct {
	oauth2.TokenSource
	relyingParty string
}

// Token requests an XSTS token that relies on the party specified in xblTokenSource.relyingParty.
// It uses the underlying [oauth2.TokenSource] to request Windows Live tokens.
func (x xblTokenSource) Token() (xsapi.Token, error) {
	token, err := x.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("request live token: %w", err)
	}
	xsts, err := auth.RequestXBLToken(context.Background(), token, x.relyingParty)
	if err != nil {
		return nil, fmt.Errorf("request xsts token for %q: %w", x.relyingParty, err)
	}
	return &xstsToken{xsts}, nil
}

// xstsToken wraps an [auth.XBLToken] for use in the xsapi package.
type xstsToken struct {
	*auth.XBLToken
}

// DisplayClaims returns [xsapi.DisplayClaims] from the user info claimed by the token.
func (t *xstsToken) DisplayClaims() xsapi.DisplayClaims {
	return xsapi.DisplayClaims{
		GamerTag: t.AuthorizationToken.DisplayClaims.UserInfo[0].GamerTag,
		XUID:     t.AuthorizationToken.DisplayClaims.UserInfo[0].XUID,
		UserHash: t.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash,
	}
}

// String returns a string representation of the XSTS token in the same format used for Authorization headers.
func (t *xstsToken) String() string {
	return fmt.Sprintf("XBL3.0 x=%s;%s", t.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash, t.AuthorizationToken.Token)
}
