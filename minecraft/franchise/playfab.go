package franchise

import (
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/playfab"
	"github.com/sandertv/gophertunnel/playfab/title"
	"golang.org/x/text/language"
)

type PlayFabIdentityProvider struct {
	Environment      *AuthorizationEnvironment
	IdentityProvider playfab.IdentityProvider

	LoginConfig playfab.LoginConfig

	DeviceConfig *DeviceConfig
	UserConfig   *UserConfig
}

func (i PlayFabIdentityProvider) TokenConfig() (*TokenConfig, error) {
	if i.Environment == nil {
		return nil, errors.New("minecraft/franchise: PlayFabIdentityProvider: Environment is nil")
	}
	if i.IdentityProvider == nil {
		return nil, errors.New("minecraft/franchise: PlayFabIdentityProvider: IdentityProvider is nil")
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

	config := i.LoginConfig
	if config.Title == "" {
		config.Title = title.Title(i.Environment.PlayFabTitleID)
	}
	identity, err := i.IdentityProvider.Login(config)
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
