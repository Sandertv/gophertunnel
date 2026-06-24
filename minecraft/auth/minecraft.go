package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/df-mc/go-xsapi/v2"
	"github.com/df-mc/go-xsapi/v2/xal/nsal"
	"github.com/df-mc/go-xsapi/v2/xal/xsts"
	"github.com/sandertv/gophertunnel/minecraft/auth/authclient"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// minecraftAuthURL is the URL that an authentication request is made to to get an encoded JWT claim chain.
var minecraftAuthURL = &url.URL{
	Scheme: "https",
	Host:   "multiplayer.minecraft.net",
	Path:   "/authentication",
} // https://multiplayer.minecraft.net/authentication

// MinecraftRelyingParty is the Xbox Live relying party used to request the
// XSTS token sent to the Minecraft authentication endpoint.
const MinecraftRelyingParty = "https://multiplayer.minecraft.net/"

// MinecraftTokenSigner adapts an XSTS token source to the TokenAndSignaturer
// interface for the Minecraft authentication endpoint. The endpoint does not
// need NSAL title-data resolution because its relying party is fixed and the
// returned signature policy is not used by [RequestMinecraftChain].
type MinecraftTokenSigner struct {
	Source xsts.TokenSource
}

// TokenAndSignature requests an XSTS token for [MinecraftRelyingParty].
func (s MinecraftTokenSigner) TokenAndSignature(ctx context.Context, _ *url.URL) (*xsts.Token, nsal.SignaturePolicy, error) {
	if s.Source == nil {
		return nil, nsal.SignaturePolicy{}, errors.New("minecraft/auth: nil XSTS token source")
	}
	token, err := s.Source.XSTSToken(ctx, MinecraftRelyingParty)
	return token, nsal.AuthPolicy, err
}

// RequestMinecraftChain requests a fully processed Minecraft JWT chain using
// signer and the ECDSA private key passed. The key will later be used to
// initialise encryption, and must be saved for when packets need to be
// decrypted/encrypted.
func RequestMinecraftChain(ctx context.Context, signer xsapi.TokenAndSignaturer, client *http.Client, key *ecdsa.PrivateKey) (string, error) {
	data, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshal public key: %w", err)
	}
	token, _, err := signer.TokenAndSignature(ctx, minecraftAuthURL)
	if err != nil {
		return "", fmt.Errorf("request XSTS token: %w", err)
	}

	// The body of the requests holds a JSON object with one key in it, the 'identityPublicKey', which holds
	// the public key data of the private key passed.
	body := `{"identityPublicKey":"` + base64.StdEncoding.EncodeToString(data) + `"}`
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, minecraftAuthURL.String(), strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST %v: %w", minecraftAuthURL, err)
	}

	request.Header.Set("User-Agent", "MCPE/Android")
	request.Header.Set("Client-Version", protocol.CurrentVersion)
	request.Header.Set("Content-Type", "application/json")
	token.SetAuthHeader(request)

	resp, err := authclient.SendRequestWithRetries(ctx, client, request, authclient.RetryOptions{Attempts: 5})
	if err != nil {
		return "", fmt.Errorf("POST %v: %w", minecraftAuthURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body []byte
		if resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
		}
		return "", fmt.Errorf("POST %v: %v, body: %s", minecraftAuthURL, resp.Status, string(body))
	}
	data, err = io.ReadAll(resp.Body)
	return string(data), err
}
