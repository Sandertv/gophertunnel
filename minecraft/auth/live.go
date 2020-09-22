package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// TokenSource holds an oauth2.TokenSource which uses device auth to get a code. The user authenticates using
// a code. TokenSource prints the authentication code and URL to os.Stdout. To use a different io.Writer, use
// WriterTokenSource.
var TokenSource oauth2.TokenSource = tokenSource{w: os.Stdout}

// WriterTokenSource returns a new oauth2.TokenSource which, like TokenSource, uses device auth to get a code.
// Unlike TokenSource, WriterTokenSource allows passing an io.Writer to which information on the auth URL and
// code are printed.
func WriterTokenSource(w io.Writer) oauth2.TokenSource {
	return tokenSource{w: w}
}

// RefreshTokenSource returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires.
func RefreshTokenSource(t *oauth2.Token) oauth2.TokenSource {
	c := oauth2.Config{
		ClientID: "0000000048183522",
		Endpoint: microsoft.LiveConnectEndpoint,
		Scopes:   []string{"service::user.auth.xboxlive.com::MBI_SSL"},
	}
	return c.TokenSource(context.Background(), t)
}

// tokenSource implements the oauth2.TokenSource interface. It provides a method to get an oauth2.Token using
// device auth through a call to RequestLiveToken.
type tokenSource struct{ w io.Writer }

// Token attempts to return a Live Connect token using the RequestLiveToken function.
func (t tokenSource) Token() (*oauth2.Token, error) {
	return RequestLiveTokenWriter(t.w)
}

// RequestLiveToken does a login request for Microsoft Live Connect using device auth. A login URL will be
// printed to the stdout with a user code which the user must use to submit.
// RequestLiveToken is the equivalent of RequestLiveTokenWriter(os.Stdout).
func RequestLiveToken() (*oauth2.Token, error) {
	return RequestLiveTokenWriter(os.Stdout)
}

// RequestLiveTokenWriter does a login request for Microsoft Live Connect using device auth. A login URL will
// be printed to the io.Writer passed with a user code which the user must use to submit.
// Once fully authenticated, an oauth2 token is returned which may be used to login to XBOX Live.
func RequestLiveTokenWriter(w io.Writer) (*oauth2.Token, error) {
	d, err := startDeviceAuth()
	if err != nil {
		return nil, err
	}
	_, _ = w.Write([]byte(fmt.Sprintf("Authenticate at %v using the code %v.\n", d.VerificationURI, d.UserCode)))
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
			_, _ = w.Write([]byte("Authentication successful.\n"))
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
