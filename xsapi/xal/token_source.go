package xal

import (
	"context"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/xsapi"
	"golang.org/x/oauth2"
	"sync"
)

func RefreshTokenSource(underlying oauth2.TokenSource, relyingParty string) xsapi.TokenSource {
	return &refreshTokenSource{
		underlying: underlying,

		relyingParty: relyingParty,
	}
}

type refreshTokenSource struct {
	underlying oauth2.TokenSource

	relyingParty string

	t  *oauth2.Token
	x  *auth.XBLToken
	mu sync.Mutex
}

func (r *refreshTokenSource) Token() (_ xsapi.Token, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.t == nil || !r.t.Valid() || r.x == nil {
		r.t, err = r.underlying.Token()
		if err != nil {
			return nil, fmt.Errorf("request underlying token: %w", err)
		}
		r.x, err = auth.RequestXBLToken(context.Background(), r.t, r.relyingParty)
		if err != nil {
			return nil, fmt.Errorf("request xbox live token: %w", err)
		}
	}
	return &token{r.x}, nil
}

type token struct {
	*auth.XBLToken
}

func (t *token) DisplayClaims() xsapi.DisplayClaims {
	return t.AuthorizationToken.DisplayClaims.UserInfo[0]
}

func (t *token) String() string {
	return fmt.Sprintf("XBL3.0 x=%s;%s", t.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash, t.AuthorizationToken.Token)
}
