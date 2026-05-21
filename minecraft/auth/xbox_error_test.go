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
	for name, want := range map[string]string{
		"did":      "0xdevice-id",
		"existing": "1",
		"redirect": "https://sisu.xboxlive.com/sisu_desktop.srf",
		"sid":      "session-id",
	} {
		if got := query.Get(name); got != want {
			t.Fatalf("unexpected %q query parameter: got %q, want %q", name, got, want)
		}
	}
	if got := query.Get("sig"); got == "" {
		t.Fatal("expected signature query parameter")
	}
}
