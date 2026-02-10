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
//
// Note: The returned TokenSource captures ctx and reuses it for all subsequent refreshes.
// If you intend to keep the returned TokenSource around after the initial operation completes
// (for example, beyond a dial context that may be cancelled), pass a context stripped of
// cancellation/deadlines such as context.WithoutCancel(ctx). This preserves context values
// (like oauth2.HTTPClient or auth token caches) without future refresh failures due to ctx
// cancellation.
func (e *AuthorizationEnvironment) TokenSource(ctx context.Context, src oauth2.TokenSource, config TokenConfig) TokenSource {
	if ctx == nil {
		ctx = context.Background()
	}
	return &tokenSource{
		identity: playfab.XBLIdentityProvider{
			TokenSource: &xblTokenSource{
				TokenSource:  src,
				relyingParty: playfab.RelyingParty,
				ctx:          ctx,
			},
		},
		env:    e,
		config: config,
		ctx:    ctx,
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
	ctx   context.Context
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

	ctx := s.ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*15)
		defer cancel()
	}
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
	ctx          context.Context
}

// Token requests an XSTS token that relies on the party specified in xblTokenSource.relyingParty.
// It uses the underlying [oauth2.TokenSource] to request Windows Live tokens.
func (x xblTokenSource) Token() (xsapi.Token, error) {
	token, err := x.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("request live token: %w", err)
	}
	ctx := x.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	xsts, err := auth.RequestXBLToken(ctx, token, x.relyingParty)
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
