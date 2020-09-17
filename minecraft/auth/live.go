package auth

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// RequestLiveToken does a login request for Microsoft Live using the login and password passed. If
// successful, a token containing the access token, refresh token, expiry and user ID is returned.
func RequestLiveToken() (*oauth2.Token, error) {
	conf := oauth2.Config{
		ClientID:    "0000000048183522",
		RedirectURL: "https://login.live.com/oauth20_desktop.srf",
		Endpoint:    microsoft.LiveConnectEndpoint,
		Scopes:      []string{"service::user.auth.xboxlive.com::MBI_SSL"},
	}
	url := conf.AuthCodeURL("abc", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	tok, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}
