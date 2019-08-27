package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TokenPair holds a pair of an access token and a refresh token, which may be used to refresh the access
// token. It also holds an expiry time to keep track of the time at which a new access token must be
// requested.
type TokenPair struct {
	access     string
	refresh    string
	expiryTime time.Time
}

// NewTokenPair returns a new token pair using an access and refresh token, and their expiry time.
func NewTokenPair(access, refresh string, expiryTime time.Duration) *TokenPair {
	return &TokenPair{
		access:     access,
		refresh:    refresh,
		expiryTime: time.Now().Add(expiryTime * time.Second),
	}
}

// Valid checks if the access token is currently valid, meaning the expiry time has not yet passed.
func (token *TokenPair) Valid() bool {
	return time.Now().Before(token.expiryTime)
}

// AccessToken returns the access token of the token pair.
func (token *TokenPair) AccessToken() string {
	return token.access
}

// RefreshToken returns the refresh token of the token pair.
func (token *TokenPair) RefreshToken() string {
	return token.refresh
}

// ValidOrRefresh checks if the token pair is currently valid, and if it isn't, refreshes it. It is equivalent
// to calling TokenPair.Valid() and TokenPair.Refresh() if it isn't.
func (token *TokenPair) ValidOrRefresh() error {
	if !token.Valid() {
		return token.Refresh()
	}
	return nil
}

// Refresh refreshes the access token of the pair, using the refresh token to refresh it. The access token
// is refreshed when calling this, regardless of the expiry time.
// If successful, the token pair's access and refresh tokens are updated.
func (token *TokenPair) Refresh() error {
	uri := liveTokenURL + "?" + url.Values{
		"grant_type":    []string{"refresh_token"},
		"client_id":     []string{"00000000441cc96b"},
		"scope":         []string{"service::user.auth.xboxlive.com::MBI_SSL"},
		"refresh_token": []string{token.refresh},
	}.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		return fmt.Errorf("GET %v: %v", uri, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %v: %v", uri, resp.Status)
	}
	m := make(map[string]interface{})
	_ = json.NewDecoder(resp.Body).Decode(&m)

	token.access = m["access_token"].(string)
	token.refresh = m["refresh_token"].(string)
	token.expiryTime = time.Now().Add(time.Duration(m["expires_in"].(float64)))
	return nil
}
