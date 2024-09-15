package franchise

import (
	"errors"
	"fmt"
	"github.com/df-mc/go-playfab"
	"github.com/df-mc/go-playfab/title"
)

// PlayFabIdentityProvider implements IdentityProvider for PlayFab, a primary
// platform used for authentication and authorization with franchise services.
//
// It is implemented to integrate with PlayFab's authentication, providing methods for
// obtaining identity tokens specific to PlayFab. It leverages PlayFab's external identity
// platform to facilitate sign-in and authorization.
type PlayFabIdentityProvider struct {
	// Environment represents the environment used for authorization with franchise services, including various fields
	// such as the base URI for making requests. It is essential for setting up the authorization context and directing
	// requests to the appropriate URL.
	Environment *AuthorizationEnvironment

	// IdentityProvider is an implementation of [playfab.IdentityProvider] from an external platform that supports
	// signing in to PlayFab with its own token, such as Xbox Live using [playfab.XBLIdentityProvider].
	IdentityProvider playfab.IdentityProvider

	// LoginConfig contains the base [playfab.LoginConfig] used to obtain a [playfab.Identity] from the
	// IdentityProvider. If the [playfab.LoginConfig.PlayFabTitleID] field is left nil, it will be set
	// automatically from [AuthorizationEnvironment.PlayFabTitleID]. It includes parameters for signing
	// in to PlayFab, such as options for creating a new account if one does not already exist.
	LoginConfig playfab.LoginConfig

	// DeviceConfig is an optional [DeviceConfig] to be set as [TokenConfig.Device]. If left nil, a default [DeviceConfig]
	// will be set and used. It provides device-specific details that may influence the authorization for franchise services.
	DeviceConfig *DeviceConfig

	// UserConfig is an optional [UserConfig] to be set as [TokenConfig.User]. If left nil, a default [UserConfig]
	// will be set. It provides user-specific details required for authorization, such as language preferences.
	// Note that the [UserConfig.Token] and [UserConfig.TokenType] fields will be overridden to use PlayFab as the
	// identity provider.
	UserConfig *UserConfig
}

// TokenConfig signs in to PlayFab using [PlayFabIdentityProvider.IdentityProvider] to authenticate and authorize
// with franchise services. After a successful sign-in, the [playfab.Identity.SessionTicket] obtained from PlayFab
// will be set into [UserConfig.Token] with [UserConfig.TokenType] set to TokenTypePlayFab.
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
		i.UserConfig = defaultUserConfig()
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
		Device: *i.DeviceConfig,
		User:   user,

		Environment: i.Environment,
	}, nil
}
