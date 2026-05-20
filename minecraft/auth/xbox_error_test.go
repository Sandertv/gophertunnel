package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"testing"
)

func TestNewAccountCreationRequiredError(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate proof key: %v", err)
	}

	accountErr, err := newAccountCreationRequiredError(
		&XboxError{XboxErrorCode: "2148916233"},
		http.Header{"X-Sessionid": []string{"session-id"}},
		[]byte(`{"WebPage":"https://sisu.xboxlive.com/sisu_create_account.srf?existing=1"}`),
		&deviceToken{
			DisplayClaims: DeviceDisplayClaims{
				DeviceInfo: DeviceInfo{DeviceID: "device-id"},
			},
			proofKey: key,
		},
	)
	if err != nil {
		t.Fatalf("new account creation error: %v", err)
	}
	if accountErr.SignupURL == nil {
		t.Fatal("expected signup URL")
	}

	query := accountErr.SignupURL.Query()
	if query.Get("existing") != "1" {
		t.Fatalf("expected existing query parameter to be preserved, got %q", query.Get("existing"))
	}
	if query.Get("sig") == "" {
		t.Fatal("expected signature query parameter")
	}
	if query.Get("did") != "0xdevice-id" {
		t.Fatalf("unexpected device ID query parameter: %q", query.Get("did"))
	}
	if query.Get("sid") != "session-id" {
		t.Fatalf("unexpected session ID query parameter: %q", query.Get("sid"))
	}
	if query.Get("redirect") != "https://sisu.xboxlive.com/sisu_desktop.srf" {
		t.Fatalf("unexpected redirect query parameter: %q", query.Get("redirect"))
	}
}
