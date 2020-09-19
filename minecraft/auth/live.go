package auth

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"time"
)

// TokenSource holds an oauth2.TokenSource which uses device auth to get a code. The user authenticates using
// a code.
var TokenSource oauth2.TokenSource = tokenSource{}

// tokenSource implements the oauth2.TokenSource interface. It provides a method to get an oauth2.Token using
// device auth through a call to RequestLiveToken.
type tokenSource struct{}

// Token attempts to return a Live Connect token using the RequestLiveToken function.
func (t tokenSource) Token() (*oauth2.Token, error) {
	return RequestLiveToken()
}

// RequestLiveToken does a login request for Microsoft Live Connect using device auth. A login URL will be
// printed together with a user code which the user must use to submit.
// Once fully authenticated, an oauth2 token is returned which may be used to login to XBOX Live.
func RequestLiveToken() (*oauth2.Token, error) {
	d, err := startDeviceAuth()
	if err != nil {
		return nil, err
	}
	fmt.Println("Authenticate at", d.VerificationURI, "using the code", d.UserCode+".")
	ticker := time.NewTicker(time.Second * time.Duration(d.Interval))
	defer ticker.Stop()

	for range ticker.C {
		t, err := pollDeviceAuth(d.DeviceCode)
		if err != nil {
			return nil, fmt.Errorf("error polling for device auth: %w", err)
		}
		// If the token could not be obtained yet (authentication wasn't finished yet), the token is nil.
		// We just retry if this is the case.
		if t != nil {
			fmt.Println("Authentication successful.")
			return t, nil
		}
	}
	panic("unreachable")
}

// startDeviceAuth starts the device auth, retrieving a login URI for the user and a code the user needs to
// enter.
func startDeviceAuth() (*deviceAuthConnect, error) {
	resp, err := http.PostForm("https://login.live.com/oauth20_connect.srf", url.Values{
		"client_id":     []string{"0000000048183522"},
		"scope":         []string{"service::user.auth.xboxlive.com::MBI_SSL"},
		"response_type": []string{"device_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_connect.srf: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_connect.srf: %v", resp.Status)
	}
	data := new(deviceAuthConnect)
	return data, json.NewDecoder(resp.Body).Decode(data)
}

// pollDeviceAuth polls the token endpoint for the device code. A token is returned if the user authenticated
// successfully. If the user has not yet authenticated, err is nil but the token is nil too.
func pollDeviceAuth(deviceCode string) (t *oauth2.Token, err error) {
	resp, err := http.PostForm("https://login.live.com/oauth20_token.srf", url.Values{
		"client_id":   []string{"0000000048183522"},
		"grant_type":  []string{"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": []string{deviceCode},
	})
	if err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_token.srf: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	poll := new(deviceAuthPoll)
	if err := json.NewDecoder(resp.Body).Decode(poll); err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_token.srf: json decode: %w", err)
	}
	if poll.Error == "authorization_pending" {
		return nil, nil
	} else if poll.Error == "" {
		return &oauth2.Token{
			AccessToken:  poll.AccessToken,
			TokenType:    poll.TokenType,
			RefreshToken: poll.RefreshToken,
			Expiry:       time.Now().Add(time.Duration(poll.ExpiresIn) * time.Second),
		}, nil
	}
	return nil, fmt.Errorf("non-empty unknown poll error: %v", poll.Error)
}

type deviceAuthConnect struct {
	UserCode        string `json:"user_code"`
	DeviceCode      string `json:"device_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
	ExpiresIn       int    `json:"expiresIn"`
}

type deviceAuthPoll struct {
	Error        string `json:"error"`
	UserID       string `json:"user_id"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
