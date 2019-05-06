package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io/ioutil"
	"net/http"
	"strings"
)

// minecraftAuthURL is the URL that an authentication request is made to to get an encoded JWT claim chain.
const minecraftAuthURL = `https://multiplayer.minecraft.net/authentication`

// RequestMinecraftChain requests a fully processed Minecraft JWT chain using the XSTS token passed, and the
// ECDSA private key of the client. This key will later be used to initialise encryption, and must be saved
// for when packets need to be decrypted/encrypted.
func RequestMinecraftChain(token *XSTSToken, key *ecdsa.PrivateKey) (string, error) {
	data, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubKeyData := base64.StdEncoding.EncodeToString(data)

	// The body of the requests holds a JSON object with one key in it, the 'identityPublicKey', which holds
	// the public key data of the private key passed.
	body := fmt.Sprintf(`{"identityPublicKey":"%v"}`, pubKeyData)
	request, _ := http.NewRequest("POST", minecraftAuthURL, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	// The Authorization header is important in particular. It is composed of the 'uhs' found in the XSTS
	// token, and the Token it holds itself.
	request.Header.Set("Authorization", fmt.Sprintf("XBL3.0 x=%v;%v", token.DisplayClaims.XUI[0].UserHash, token.Token))
	request.Header.Set("User-Agent", "MCPE/UWP")
	request.Header.Set("Client-Version", protocol.CurrentVersion)

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return "", fmt.Errorf("POST %v: %v", minecraftAuthURL, err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("POST %v: %v", minecraftAuthURL, resp.Status)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	data, err = ioutil.ReadAll(resp.Body)
	return string(data), err
}
