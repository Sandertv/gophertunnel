package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// TokenSource holds an oauth2.TokenSource which uses device auth to get a code. The user authenticates using
// a code. TokenSource prints the authentication code and URL to os.Stdout. To use a different io.Writer, use
// WriterTokenSource. TokenSource automatically refreshes tokens.
var TokenSource oauth2.TokenSource = &tokenSource{w: os.Stdout}

// WriterTokenSource calls [WriterTokenSourceDevice] with the default device info.
func WriterTokenSource(w io.Writer) oauth2.TokenSource {
	return WriterTokenSourceDevice(w, DeviceAndroid)
}

// WriterTokenSourceDevice returns a new oauth2.TokenSource which, like TokenSource, uses device auth to get a code.
// Unlike TokenSource, WriterTokenSourceDevice allows passing an io.Writer to which information on the auth URL and
// code are printed. WriterTokenSourceDevice automatically refreshes tokens.
func WriterTokenSourceDevice(w io.Writer, d Device) oauth2.TokenSource {
	return &tokenSource{w: w, d: d}
}

// tokenSource implements the oauth2.TokenSource interface. It provides a method to get an oauth2.Token using
// device auth through a call to RequestLiveToken.
type tokenSource struct {
	w io.Writer
	t *oauth2.Token
	d Device
}

// Token attempts to return a Live Connect token using the RequestLiveToken function.
func (src *tokenSource) Token() (*oauth2.Token, error) {
	if src.t == nil {
		t, err := RequestLiveTokenWriterDevice(src.w, src.d)
		src.t = t
		return t, err
	}
	tok, err := refreshToken(src.t, src.d)
	if err != nil {
		return nil, err
	}
	// Update the token to use to refresh for the next time Token is called.
	src.t = tok
	return tok, nil
}

// RefreshTokenSource calls [RefreshTokenSourceDevice] with the default device info.
func RefreshTokenSource(t *oauth2.Token) oauth2.TokenSource {
	return RefreshTokenSourceDevice(t, DeviceAndroid)
}

// RefreshTokenSourceDevice returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires. Note that this function must be used over oauth2.ReuseTokenSource
// due to that function not refreshing with the correct scopes.
func RefreshTokenSourceDevice(t *oauth2.Token, d Device) oauth2.TokenSource {
	return RefreshTokenSourceWriterDevice(t, os.Stdout, d)
}

// RefreshTokenSourceWriter calls [RefreshTokenSourceWriterDevice] with the default device info.
func RefreshTokenSourceWriter(t *oauth2.Token, w io.Writer) oauth2.TokenSource {
	return RefreshTokenSourceWriterDevice(t, w, DeviceAndroid)
}

// RefreshTokenSourceWriterDevice returns a new oauth2.TokenSource using the oauth2.Token passed that automatically
// refreshes the token everytime it expires. It requests from io.Writer if the oauth2.Token is invalid.
// Note that this function must be used over oauth2.ReuseTokenSource due to that
// function not refreshing with the correct scopes.
func RefreshTokenSourceWriterDevice(t *oauth2.Token, w io.Writer, d Device) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(t, &tokenSource{w: w, t: t, d: d})
}

// RequestLiveToken calls [RequestLiveTokenDevice] with the default device info.
func RequestLiveToken() (*oauth2.Token, error) {
	return RequestLiveTokenDevice(DeviceAndroid)
}

// RequestLiveTokenDevice does a login request for Microsoft Live Connect using device auth. A login URL will be
// printed to the stdout with a user code which the user must use to submit.
// RequestLiveTokenDevice is the equivalent of RequestLiveTokenWriter(os.Stdout).
func RequestLiveTokenDevice(deviceType Device) (*oauth2.Token, error) {
	return RequestLiveTokenWriterDevice(os.Stdout, deviceType)
}

// RequestLiveTokenWriter calls [RequestLiveTokenWriterDevice] with the default device info.
func RequestLiveTokenWriter(w io.Writer) (*oauth2.Token, error) {
	return RequestLiveTokenWriterDevice(w, DeviceAndroid)
}

// RequestLiveTokenWriterDevice does a login request for Microsoft Live Connect using device auth. A login URL will
// be printed to the io.Writer passed with a user code which the user must use to submit.
// Once fully authenticated, an oauth2 token is returned which may be used to login to XBOX Live.
func RequestLiveTokenWriterDevice(w io.Writer, deviceType Device) (*oauth2.Token, error) {
	d, err := startDeviceAuth(deviceType)
	if err != nil {
		return nil, err
	}
	_, _ = w.Write([]byte(fmt.Sprintf("Authenticate at %v using the code %v.\n", d.VerificationURI, d.UserCode)))
	ticker := time.NewTicker(time.Second * time.Duration(d.Interval))
	defer ticker.Stop()

	for range ticker.C {
		t, err := pollDeviceAuth(d.DeviceCode, deviceType)
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
func startDeviceAuth(deviceType Device) (*deviceAuthConnect, error) {
	resp, err := http.PostForm("https://login.live.com/oauth20_connect.srf", url.Values{
		"client_id":     {deviceType.ClientID},
		"scope":         {"service::user.auth.xboxlive.com::MBI_SSL"},
		"response_type": {"device_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_connect.srf: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_connect.srf: %v", resp.Status)
	}
	data := new(deviceAuthConnect)
	return data, json.NewDecoder(resp.Body).Decode(data)
}

// pollDeviceAuth polls the token endpoint for the device code. A token is returned if the user authenticated
// successfully. If the user has not yet authenticated, err is nil but the token is nil too.
func pollDeviceAuth(deviceCode string, deviceType Device) (t *oauth2.Token, err error) {
	resp, err := http.PostForm(microsoft.LiveConnectEndpoint.TokenURL, url.Values{
		"client_id":   {deviceType.ClientID},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode},
	})
	if err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_token.srf: %w", err)
	}
	defer resp.Body.Close()
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
	return nil, fmt.Errorf("%v: %v", poll.Error, poll.ErrorDescription)
}

// refreshToken refreshes the oauth2.Token passed and returns a new oauth2.Token. An error is returned if
// refreshing was not successful.
func refreshToken(t *oauth2.Token, deviceType Device) (*oauth2.Token, error) {
	// This function unfortunately needs to exist because golang.org/x/oauth2 does not pass the scope to this
	// request, which Microsoft Connect enforces.
	resp, err := http.PostForm(microsoft.LiveConnectEndpoint.TokenURL, url.Values{
		"client_id":     {deviceType.ClientID},
		"scope":         {"service::user.auth.xboxlive.com::MBI_SSL"},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.RefreshToken},
	})
	if err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_token.srf: %w", err)
	}
	defer resp.Body.Close()
	poll := new(deviceAuthPoll)
	if err := json.NewDecoder(resp.Body).Decode(poll); err != nil {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_token.srf: json decode: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST https://login.live.com/oauth20_token.srf: refresh error: %v", poll.Error)
	}
	return &oauth2.Token{
		AccessToken:  poll.AccessToken,
		TokenType:    poll.TokenType,
		RefreshToken: poll.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(poll.ExpiresIn) * time.Second),
	}, nil
}

type deviceAuthConnect struct {
	UserCode        string `json:"user_code"`
	DeviceCode      string `json:"device_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
	ExpiresIn       int    `json:"expires_in"`
}

type deviceAuthPoll struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	UserID           string `json:"user_id"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
}
