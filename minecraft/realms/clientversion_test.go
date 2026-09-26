package realms

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/df-mc/go-xsapi/v2/xal/xsts"
	"github.com/sandertv/gophertunnel/minecraft/auth"
)

// versionTestTransport lets tests inspect requests without accessing Realms.
type versionTestTransport func(*http.Request) (*http.Response, error)

// RoundTrip handles a request using the test callback.
func (f versionTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientVersionRetry(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			var payload []byte
			if method == http.MethodPost {
				payload = []byte(`{"playerOptIn":"OPT_IN"}`)
			}
			attempts, checks := 0, 0
			client := NewClient(nil, &http.Client{Transport: versionTestTransport(func(req *http.Request) (*http.Response, error) {
				status, body := http.StatusOK, "{}"
				if req.URL.Path == compatiblePath {
					checks++
					if req.Method != http.MethodGet || req.Header.Get("Client-Version") != "1.26.59" {
						t.Errorf("unexpected compatibility request: %s, %s", req.Method, req.Header.Get("Client-Version"))
					}
					body = compatibleResponse
				} else {
					attempts++
					data, err := io.ReadAll(req.Body)
					if err != nil {
						t.Fatal(err)
					}
					if req.Method != method || string(data) != string(payload) {
						t.Errorf("request changed on attempt %d: %s, %q", attempts, req.Method, data)
					}
					if method == http.MethodPost && req.Header.Get("Content-Type") != "application/json" {
						t.Error("missing JSON content type")
					}
					if attempts == 1 {
						if req.Header.Get("Client-Version") != "1.26.60" {
							t.Error("initial request did not use preferred version")
						}
						status, body = http.StatusBadRequest, `{"reason":"unknown_client_version"}`
					} else if req.Header.Get("Client-Version") != "1.26.59" {
						t.Error("retry did not use accepted version")
					}
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})})
			client.xblToken = &auth.XBLToken{AuthorizationToken: &xsts.Token{
				Token: "test", NotAfter: time.Now().Add(time.Hour),
				DisplayClaims: xsts.DisplayClaims{UserInfo: []xsts.UserInfo{{}}},
			}}
			client.SetClientVersion("1.26.60")
			_, status, err := client.requestWithMethod(context.Background(), method, realmsBaseURL, "/worlds/1/stories/settings", payload)
			if err != nil || status != http.StatusOK {
				t.Fatalf("request failed: status %d, error %v", status, err)
			}
			if attempts != 2 || checks != 1 || client.clientVersion() != "1.26.59" {
				t.Fatalf("unexpected negotiation: attempts %d, checks %d, version %s", attempts, checks, client.clientVersion())
			}
		})
	}
}
