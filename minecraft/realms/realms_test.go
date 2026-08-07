package realms

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/auth"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func validXBLToken() *auth.XBLToken {
	token := &auth.XBLToken{}
	token.AuthorizationToken.NotAfter = time.Now().Add(time.Hour)
	token.AuthorizationToken.Token = "token"
	token.AuthorizationToken.DisplayClaims.UserInfo = []struct {
		GamerTag string `json:"gtg"`
		XUID     string `json:"xid"`
		UserHash string `json:"uhs"`
	}{{UserHash: "user-hash"}}
	return token
}

func TestClientUpdateStorySettingsSendsRequest(t *testing.T) {
	var received *http.Request
	client := &Client{
		xblToken: validXBLToken(),
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			received = req
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		})},
	}

	settings := StorySettings{
		AutoStories:   true,
		Coordinates:   true,
		Notifications: true,
		PlayerOptIn:   "OPT_IN",
		RealmOptIn:    "OPT_OUT",
		Timeline:      false,
	}
	if err := client.UpdateStorySettings(context.Background(), 123, settings); err != nil {
		t.Fatalf("UpdateStorySettings returned error: %v", err)
	}

	if received == nil {
		t.Fatal("UpdateStorySettings did not send a request")
	}
	if got, want := received.Method, http.MethodPost; got != want {
		t.Fatalf("request method = %q, want %q", got, want)
	}
	if got, want := received.URL.String(), "https://bedrock.frontendlegacy.realms.minecraft-services.net/worlds/123/stories/settings"; got != want {
		t.Fatalf("request URL = %q, want %q", got, want)
	}
	if got, want := received.Header.Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("content type = %q, want %q", got, want)
	}
	if got, want := received.Header.Get("Authorization"), "XBL3.0 x=user-hash;token"; got != want {
		t.Fatalf("authorization = %q, want %q", got, want)
	}

	var body map[string]any
	if err := json.NewDecoder(received.Body).Decode(&body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	wantBody := map[string]any{
		"autostories":   true,
		"coordinates":   true,
		"notifications": true,
		"playerOptIn":   "OPT_IN",
		"realmOptIn":    "OPT_OUT",
		"timeline":      false,
	}
	if !reflect.DeepEqual(body, wantBody) {
		t.Fatalf("request body = %#v, want %#v", body, wantBody)
	}
}
