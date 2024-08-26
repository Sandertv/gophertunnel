package playfab

import (
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/xsapi"
)

type XBLIdentityProvider struct {
	TokenSource xsapi.TokenSource
}

func (prov XBLIdentityProvider) Login(config LoginConfig) (*Identity, error) {
	if prov.TokenSource == nil {
		return nil, errors.New("playfab: XBLIdentityProvider: TokenSource is nil")
	}

	tok, err := prov.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("request xbox live token: %w", err)
	}

	type loginConfig struct {
		LoginConfig
		XboxToken string `json:"XboxToken"`
	}
	return config.login("/Client/LoginWithXbox", loginConfig{
		LoginConfig: config,
		XboxToken:   tok.String(),
	})
}
