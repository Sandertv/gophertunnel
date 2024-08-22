package franchise

import (
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/playfab"
	"github.com/sandertv/gophertunnel/playfab/title"
	"github.com/sandertv/gophertunnel/xsapi"
	"golang.org/x/text/language"
)

type PlayFabXBLIdentityProvider struct {
	Environment *AuthorizationEnvironment
	TokenSource xsapi.TokenSource

	DeviceConfig *DeviceConfig
	UserConfig   *UserConfig
}

func (i PlayFabXBLIdentityProvider) TokenConfig() (*TokenConfig, error) {
	if i.Environment == nil {
		return nil, errors.New("minecraft/franchise: PlayFabXBLIdentityProvider: Environment is nil")
	}
	if i.TokenSource == nil {
		return nil, errors.New("minecraft/franchise: PlayFabXBLIdentityProvider: TokenSource is nil")
	}
	if i.DeviceConfig == nil {
		i.DeviceConfig = defaultDeviceConfig(i.Environment)
	}
	if i.UserConfig == nil {
		region, _ := language.English.Region()

		i.UserConfig = &UserConfig{
			Language:     language.English,
			LanguageCode: language.AmericanEnglish,
			RegionCode:   region.String(),
		}
	}

	x, err := i.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("request xbox live token: %w", err)
	}

	cfg := playfab.LoginConfig{
		Title:         title.Title(i.Environment.PlayFabTitleID),
		CreateAccount: true,
	}.WithXbox(x)
	identity, err := cfg.Login()
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	user := *i.UserConfig
	user.Token = identity.SessionTicket
	user.TokenType = TokenTypePlayFab

	return &TokenConfig{
		Device: i.DeviceConfig,
		User:   &user,

		Environment: i.Environment,
	}, nil
}
