package franchise

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
)

type MultiplayerToken struct {
	SignedToken string    `json:"signedToken"`
	ValidUntil  time.Time `json:"validUntil"`
	IssuedAt    time.Time `json:"issuedAt"`
}

// RequestMultiplayerToken requests a token for use with multiplayer servers
func RequestMultiplayerToken(ctx context.Context, env AuthorizationEnvironment, mcToken *Token, key *ecdsa.PrivateKey) (tok *MultiplayerToken, err error) {
	u, err := url.Parse(env.ServiceURI)
	if err != nil {
		return nil, fmt.Errorf("parse service URI: %w", err)
	}
	u = u.JoinPath("/api/v1.0/multiplayer/session/start")

	encodedKey, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	body := `{"publicKey":"` + base64.StdEncoding.EncodeToString(encodedKey) + `"}`

	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(body))
	req.Header.Set("Authorization", mcToken.AuthorizationHeader)
	req.Header.Set("Content-Type", "application/json")

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
		},
	}
	defer c.CloseIdleConnections()

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request multiplayer token: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST %v: %v", u, resp.Status)
	}

	var result internal.Result[*MultiplayerToken]
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode multiplayer token: %w", err)
	}

	return result.Data, nil
}
