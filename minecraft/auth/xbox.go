package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// xblUserAuthURL is the first URL that a POST request is made to, in order to obtain the XBOX Live user token.
const xblUserAuthURL = `https://user.auth.xboxlive.com/user/authenticate`

// xblDeviceAuthURL is the second URL that a POST request is made to, in order to authenticate a device.
const xblDeviceAuthURL = `https://device.auth.xboxlive.com/device/authenticate`

// xblTitleAuthURL is the third URL that a POST request is made to, in order to authenticate the title.
const xblTitleAuthURL = `https://title.auth.xboxlive.com/title/authenticate`

// xblAuthorizeURL is the last URL that a POST request is made to, in order to obtain the XSTS token, which
// is a combination of all previous tokens.
const xblAuthorizeURL = `https://xsts.auth.xboxlive.com/xsts/authorize`

// UserToken is the token obtained by requesting a user token by posting to xblUserAuthURL. Its Token field
// must be used in a request to the XSTS token.
type UserToken struct {
	IssueInstant  string
	NotAfter      string
	Token         string
	DisplayClaims struct {
		XUI []struct {
			UHS string `json:"uhs"`
		} `json:"xui"`
	}
}

// DeviceToken is the token obtained by requesting a device token by posting to xblDeviceAuthURL. Its Token
// field may be used in a request to obtain the XSTS token.
type DeviceToken struct {
	IssueInstant  string
	NotAfter      string
	Token         string
	DisplayClaims struct {
		XDI struct {
			DID string `json:"did"`
		} `json:"xdi"`
	}
}

// TitleToken is the token obtained by requesting a title token by posting to xblTitleAuthURL. Its Token field
// may be used in a request to obtain the XSTS token.
type TitleToken struct {
	IssueInstant  string
	NotAfter      string
	Token         string
	DisplayClaims struct {
		XTI struct {
			TID string `json:"tid"`
		} `json:"xti"`
	}
}

// XSTSToken is the token obtained by requesting an XSTS token from xblAuthorizeURL. It may be obtained using
// any of the tokens above, and is required for authenticating with Minecraft. Its Token and UserHash field
// in particular are used.
type XSTSToken struct {
	IssueInstant  string
	NotAfter      string
	Token         string
	DisplayClaims struct {
		XUI []struct {
			AgeGroup   string `json:"agg"`
			GamerTag   string `json:"gtg"`
			Privileges string `json:"prv"`
			XUID       string `json:"xid"`
			UserHash   string `json:"uhs"`
		} `json:"xui"`
	}
}

// RequestXSTSToken requests an XSTS token using the passed Live token pair. The token pair must be valid
// when passed in. RequestXSTSToken will not attempt to refresh the token pair if it not valid.
// RequestXSTSToken obtains the XSTS token by using the UserToken, DeviceToken and TitleToken. It appears only
// one of these tokens is actually required to produce an XSTS token valid to authenticate with Minecraft.
func RequestXSTSToken(liveToken *TokenPair) (*XSTSToken, error) {
	if !liveToken.Valid() {
		return nil, fmt.Errorf("live token is no longer valid")
	}
	c := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true,
		},
	}}
	defer c.CloseIdleConnections()
	// We first generate an ECDSA private key which will be used to provide a 'ProofKey' to each of the
	// requests, and to sign these requests.
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// All following requests here use the same ECDSA private key. This is required, and failing to do so
	// means that the signature of the second request will be refused.
	userToken, err := userToken(c, liveToken.access, key)
	if err != nil {
		return nil, err
	}
	deviceToken, err := deviceToken(c, key)
	if err != nil {
		return nil, err
	}
	titleToken, err := titleToken(c, liveToken.access, deviceToken.Token, key)
	if err != nil {
		return nil, err
	}
	return xstsToken(c, userToken.Token, deviceToken.Token, titleToken.Token, key)
}

// userToken sends a POST request to xblUserAuthURL using the Live access token passed, and the ECDSA private
// key to sign the request. Signing the request is not actually mandatory, but we do so anyway just to be
// sure.
func userToken(c *http.Client, accessToken string, key *ecdsa.PrivateKey) (token *UserToken, err error) {
	data, _ := json.Marshal(map[string]interface{}{
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
		"Properties": map[string]interface{}{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "t=" + accessToken,
			// Note that the ProofKey field here does not need to be present. Omitting this field will still
			// return a valid user token.
			"ProofKey": map[string]interface{}{
				"crv": "P-256",
				"alg": "ES256",
				"use": "sig",
				"kty": "EC",
				"x":   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.Bytes()),
			},
		},
	})
	request, _ := http.NewRequest("POST", xblUserAuthURL, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-xbl-contract-version", "1")

	// Signing the user token request is actually not mandatory. It may be omitted altogether, including the
	// ProofKey field in the Properties of the request.
	sign(request, data, key)

	resp, err := c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", xblUserAuthURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST %v: %v", xblUserAuthURL, resp.Status)
	}
	token = &UserToken{}
	return token, json.NewDecoder(resp.Body).Decode(token)
}

// deviceToken sends a POST request to xblDeviceAuthURL using the ECDSA private key passed to sign the
// request. Note that the device token is not mandatory to obtain a valid XSTS token.
func deviceToken(c *http.Client, key *ecdsa.PrivateKey) (token *DeviceToken, err error) {
	data, _ := json.Marshal(map[string]interface{}{
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
		"Properties": map[string]interface{}{
			"DeviceType": "Nintendo",
			// These may simply be random UUIDs.
			"Id":           uuid.Must(uuid.NewRandom()).String(),
			"SerialNumber": uuid.Must(uuid.NewRandom()).String(),
			"Version":      "0.0.0.0",
			// Note the different AuthMethod here. Other requests typically have the RPS AuthMethod, but this
			// uses ProofOfPossession.
			"AuthMethod": "ProofOfPossession",
			"ProofKey": map[string]interface{}{
				"crv": "P-256",
				"alg": "ES256",
				"use": "sig",
				"kty": "EC",
				"x":   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.Bytes()),
			},
		},
	})
	request, _ := http.NewRequest("POST", xblDeviceAuthURL, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-xbl-contract-version", "1")
	sign(request, data, key)

	resp, err := c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", xblDeviceAuthURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST %v: %v", xblDeviceAuthURL, resp.Status)
	}
	token = &DeviceToken{}
	return token, json.NewDecoder(resp.Body).Decode(token)
}

// titleToken sends a POST request to xblTitleAuthURL using the device and Live access token passed. The
// request is signed using the ECDSA private key passed.
func titleToken(c *http.Client, accessToken, deviceToken string, key *ecdsa.PrivateKey) (token *TitleToken, err error) {
	data, _ := json.Marshal(map[string]interface{}{
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
		"Properties": map[string]interface{}{
			"AuthMethod":  "RPS",
			"DeviceToken": deviceToken,
			"SiteName":    "user.auth.xboxlive.com",
			"RpsTicket":   "t=" + accessToken,
			"ProofKey": map[string]interface{}{
				"crv": "P-256",
				"alg": "ES256",
				"use": "sig",
				"kty": "EC",
				"x":   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.Bytes()),
			},
		},
	})
	request, _ := http.NewRequest("POST", xblTitleAuthURL, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-xbl-contract-version", "1")
	sign(request, data, key)

	resp, err := c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", xblTitleAuthURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST %v: %v", xblTitleAuthURL, resp.Status)
	}
	token = &TitleToken{}
	return token, json.NewDecoder(resp.Body).Decode(token)
}

// xstsToken sends a POST request to xblAuthorizeURL using the user, device and title token passed, and the
// ECDSA private key to sign the request. The device token, title token and signature are not mandatory to
// produce a valid XSTS token, but we require them here just in case.
func xstsToken(c *http.Client, userToken, deviceToken, titleToken string, key *ecdsa.PrivateKey) (token *XSTSToken, err error) {
	data, _ := json.Marshal(map[string]interface{}{
		// RelyingParty MUST be this URL to produce an XSTS token which may be used for Minecraft
		// authentication.
		"RelyingParty": "https://multiplayer.minecraft.net/",
		"TokenType":    "JWT",
		"Properties": map[string]interface{}{
			// DeviceToken is not required for Minecraft auth. The key may simply not be present.
			"DeviceToken": deviceToken,
			// TitleToken is also not required for Minecraft auth. The key may simply not be present.
			"TitleToken": titleToken,
			"UserTokens": []string{userToken},
			"SandboxId":  "RETAIL",
		},
	})
	request, _ := http.NewRequest("POST", xblAuthorizeURL, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("x-xbl-contract-version", "1")

	// Signing the XSTS token request is not necessary. The header may simply not be present.
	sign(request, data, key)

	resp, err := c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", xblAuthorizeURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	token = &XSTSToken{}
	return token, json.NewDecoder(resp.Body).Decode(token)
}

// sign signs the request passed containing the body passed. It signs the request using the ECDSA private key
// passed. If the request has a 'ProofKey' field in the Properties field, that key must be passed here.
func sign(request *http.Request, body []byte, key *ecdsa.PrivateKey) {
	currentTime := windowsTimestamp()
	hash := sha256.New()

	// Signature policy version (0, 0, 0, 1) + 0 byte.
	buf := bytes.NewBuffer([]byte{0, 0, 0, 1, 0})
	// Timestamp + 0 byte.
	_ = binary.Write(buf, binary.BigEndian, currentTime)
	buf.Write([]byte{0})
	hash.Write(buf.Bytes())

	// HTTP method, generally POST + 0 byte.
	hash.Write([]byte("POST"))
	hash.Write([]byte{0})
	// Request uri path + raw query + 0 byte.
	hash.Write([]byte(request.URL.Path + request.URL.RawQuery))
	hash.Write([]byte{0})

	// Authorization header if present, otherwise an empty string + 0 byte.
	hash.Write([]byte(request.Header.Get("Authorization")))
	hash.Write([]byte{0})

	// Body data (only up to a certain limit, but this limit is practically never reached) + 0 byte.
	hash.Write(body)
	hash.Write([]byte{0})

	// Sign the checksum produced, and combine the 'r' and 's' into a single signature.
	r, s, _ := ecdsa.Sign(rand.Reader, key, hash.Sum(nil))
	signature := append(r.Bytes(), s.Bytes()...)

	// The signature begins with 12 bytes, the first being the signature policy version (0, 0, 0, 1) again,
	// and the other 8 the timestamp again.
	buf = bytes.NewBuffer([]byte{0, 0, 0, 1})
	_ = binary.Write(buf, binary.BigEndian, currentTime)

	// Append the signature to the other 12 bytes, and encode the signature with standard base64 encoding.
	sig := append(buf.Bytes(), signature...)
	request.Header.Set("Signature", base64.StdEncoding.EncodeToString(sig))
}

// windowsTimestamp returns a Windows specific timestamp. It has a certain offset from Unix time which must be
// accounted for.
func windowsTimestamp() int64 {
	return (time.Now().Unix() + 11644473600) * 10000000
}
