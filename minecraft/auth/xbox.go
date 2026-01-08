package auth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// XBLToken holds info on the authorization token used for authenticating with XBOX Live.
type XBLToken struct {
	AuthorizationToken struct {
		DisplayClaims struct {
			// UserInfo contains the user information claimed from the authorization token.
			// GamerTag and XUID are only populated on the "xboxlive.com" relying party.
			// The rest only return UserHash.
			UserInfo []struct {
				GamerTag string `json:"gtg"`
				XUID     string `json:"xid"`
				UserHash string `json:"uhs"`
			} `json:"xui"`
		}
		IssueInstant time.Time
		NotAfter     time.Time
		Token        string
	}
}

// SetAuthHeader sets the 'Authorization' header used for Minecraft related endpoints that
// need an XBOX Live authenticated caller.
func (t XBLToken) SetAuthHeader(r *http.Request) {
	r.Header.Set("Authorization", fmt.Sprintf("XBL3.0 x=%v;%v", t.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash, t.AuthorizationToken.Token))
}

// expirationDelta is the amount of time before the token expires that it is considered valid.
const expirationDelta = time.Minute

// Valid returns whether the XBLToken is valid.
func (t XBLToken) Valid() bool {
	return time.Now().Before(t.AuthorizationToken.NotAfter.Add(-expirationDelta))
}

// XBLConfig specifies the configuration for authenticating with Xbox Live and Microsoft services.
type XBLConfig struct {
	// ClientID is the ID used for the SISU authorization flow.
	// It is also used for the OAuth2 device code flow in [RequestLiveToken].
	ClientID string
	// DeviceType indicates the device type used for requesting device tokens in Xbox Live.
	DeviceType string
	// Version indicates the version of the authentication library used in the client.
	Version string
	// UserAgent is the 'User-Agent' header sent by the authentication library used in the client.
	UserAgent string
}

// XBLTokenObtainer requests XBL tokens using a specific Config and HTTP client.
// If Client is nil, a default client is used.
type XBLTokenObtainer struct {
	Config XBLConfig
	Client *http.Client
}

// defaultXBLHTTPClient is the default HTTP client used for requests made by XBLTokenObtainer.
var defaultXBLHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true,
		},
	},
}

// httpClient returns the HTTP client used for requests made by XBLTokenObtainer.
func (o XBLTokenObtainer) httpClient() *http.Client {
	if o.Client != nil {
		return o.Client
	}
	return defaultXBLHTTPClient
}

// contextKey is a type used for context key used to [context.WithValue].
type contextKey struct{}

// tokenCacheContextKey is the context key used for holding an XBLTokenCache in [context.Context].
var tokenCacheContextKey contextKey

// XBLTokenCache caches device tokens for requesting Xbox Live tokens.
// It may be created from [XBLConfig.NewTokenCache] and included to a
// [context.Context] for re-using the device token in [RequestXBLToken].
type XBLTokenCache struct {
	// config is the Config used to request device tokens.
	// It contains platform-specific values for logging in with different device types.
	config XBLConfig
	// device caches the device token requested by XBLTokenCache.
	device *deviceToken
	// mu guards device from concurrent access.
	mu sync.Mutex
}

// NewTokenCache returns an XBLTokenCache that can be used to re-use XBL tokens
// in [RequestXBLToken].
func (conf XBLConfig) NewTokenCache() *XBLTokenCache {
	return &XBLTokenCache{
		config: conf,
	}
}

// WithXBLTokenCache returns a [context.Context] which contains the XBLTokenCache.
// The returned [context.Context] can be used in [RequestXBLToken] for
// re-using the device token as possible to avoid issuing too many device
// tokens and incurring rate limiting from XASD (Xbox Authentication Service for
// Devices).
func WithXBLTokenCache(parent context.Context, cache *XBLTokenCache) context.Context {
	return context.WithValue(parent, tokenCacheContextKey, cache)
}

// deviceToken returns the cached device token. If the device token is no longer
// valid or has not yet been requested, it will request a device token with a new
// proof key, using the HTTP client from the obtainer.
func (x *XBLTokenCache) deviceToken(ctx context.Context, o XBLTokenObtainer) (*deviceToken, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.device != nil && x.device.Valid() {
		return x.device, nil
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate proof key: %w", err)
	}
	if x.config != o.Config {
		return nil, errors.New("xbl token cache config mismatch")
	}
	d, err := o.obtainDeviceToken(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("obtain device token: %w", err)
	}
	x.device = d
	return d, nil
}

var (
	// AndroidConfig is the configuration used in Minecraft: Bedrock Edition for Android devices.
	AndroidConfig = XBLConfig{
		DeviceType: "Android",
		ClientID:   "0000000048183522",
		Version:    "8.0.0",
		UserAgent:  "XAL Android 2020.07.20200714.000",
	}
	// IOSConfig is the configuration used in Minecraft: Bedrock Edition for iOS devices.
	IOSConfig = XBLConfig{
		DeviceType: "iOS",
		ClientID:   "000000004c17c01a",
		Version:    "15.6.1",
		UserAgent:  "XAL iOS 2021.11.20211021.000",
	}
	// Win32Config is the configuration used in Minecraft: Bedrock Edition for Windows devices.
	// Please note that the actual GDK/UWP build of the game requests the device token in more different way.
	Win32Config = XBLConfig{
		DeviceType: "Win32",
		ClientID:   "0000000040159362",
		Version:    "10.0.25398.4909",
		UserAgent:  "XAL Win32 2021.11.20220411.002",
	}
	// NintendoConfig is the configuration used in Minecraft: Bedrock Edition for Nintendo Switch.
	NintendoConfig = XBLConfig{
		DeviceType: "Nintendo",
		ClientID:   "00000000441cc96b",
		Version:    "0.0.0",
		UserAgent:  "XAL",
	}
	// PlayStationConfig is the configuration used in Minecraft: Bedrock Edition for PlayStation devices.
	PlayStationConfig = XBLConfig{
		DeviceType: "Playstation",
		ClientID:   "000000004827c78e",
		Version:    "10.0.0",
		UserAgent:  "XAL",
	}
)

// RequestXBLToken calls [XBLConfig.RequestXBLToken] with the default device info.
func RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	return XBLTokenObtainer{Config: AndroidConfig}.RequestXBLToken(ctx, liveToken, relyingParty)
}

// RequestXBLToken requests an XBOX Live auth token using the passed Live token pair.
func (conf XBLConfig) RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	return XBLTokenObtainer{Config: conf}.RequestXBLToken(ctx, liveToken, relyingParty)
}

// RequestXBLToken requests an XBOX Live auth token using the passed Live token pair.
func (o XBLTokenObtainer) RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	if !liveToken.Valid() {
		return nil, fmt.Errorf("live token is no longer valid")
	}
	d, err := o.getDeviceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request device token: %w", err)
	}
	return o.obtainXBLToken(ctx, liveToken, d, relyingParty)
}

// getDeviceToken attempts to use the cache from [context.Context], otherwise it will request
// a new device token using a new proof key.
func (o XBLTokenObtainer) getDeviceToken(ctx context.Context) (*deviceToken, error) {
	if cache, ok := ctx.Value(tokenCacheContextKey).(*XBLTokenCache); ok && cache != nil {
		// If the context has a value with XBLTokenCache, we re-use them.
		return cache.deviceToken(ctx, o)
	}
	// We first generate an ECDSA private key which will be used to provide a 'ProofKey' to each of the
	// requests, and to sign these requests.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate proof key: %w", err)
	}
	d, err := o.obtainDeviceToken(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("obtain device token: %w", err)
	}
	return d, nil
}

func (o XBLTokenObtainer) obtainXBLToken(ctx context.Context, liveToken *oauth2.Token, device *deviceToken, relyingParty string) (*XBLToken, error) {
	data, err := json.Marshal(map[string]any{
		"AccessToken":       "t=" + liveToken.AccessToken,
		"AppId":             o.Config.ClientID,
		"DeviceToken":       device.Token,
		"Sandbox":           "RETAIL",
		"UseModernGamertag": true,
		"SiteName":          "user.auth.xboxlive.com",
		"RelyingParty":      relyingParty,
		"ProofKey": map[string]any{
			"crv": "P-256",
			"alg": "ES256",
			"use": "sig",
			"kty": "EC",
			"x":   base64.RawURLEncoding.EncodeToString(padTo32Bytes(device.proofKey.PublicKey.X)),
			"y":   base64.RawURLEncoding.EncodeToString(padTo32Bytes(device.proofKey.PublicKey.Y)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling XBL auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://sisu.xboxlive.com/authorize", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("POST %v: %w", "https://sisu.xboxlive.com/authorize", err)
	}
	req.Header.Set("x-xbl-contract-version", "1")
	sign(req, data, device.proofKey)

	resp, err := o.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %w", "https://sisu.xboxlive.com/authorize", err)
	}
	defer resp.Body.Close()

	updateServerTimeFromHeaders(resp.Header)

	if resp.StatusCode != 200 {
		// Xbox Live returns a custom error code in the x-err header.
		if errorCode := resp.Header.Get("x-err"); errorCode != "" {
			return nil, fmt.Errorf("POST %v: %v", "https://sisu.xboxlive.com/authorize", parseXboxErrorCode(errorCode))
		}
		return nil, fmt.Errorf("POST %v: %v", "https://sisu.xboxlive.com/authorize", resp.Status)
	}
	info := new(XBLToken)
	return info, json.NewDecoder(resp.Body).Decode(info)
}

// deviceToken is the token obtained by requesting a device token by posting to xblDeviceAuthURL. Its Token
// field may be used in a request to obtain the XSTS token.
type deviceToken struct {
	IssueInstant time.Time `json:"IssueInstant"`
	NotAfter     time.Time `json:"NotAfter"`
	Token        string

	// proofKey is the private key used to sign requests in Xbox Live.
	proofKey *ecdsa.PrivateKey
}

// Valid returns whether the device token is valid.
func (d *deviceToken) Valid() bool {
	return time.Now().Before(d.NotAfter)
}

// obtainDeviceToken sends a POST request to the device auth endpoint using the ECDSA private key passed to
// sign the request.
func (o XBLTokenObtainer) obtainDeviceToken(ctx context.Context, key *ecdsa.PrivateKey) (token *deviceToken, err error) {
	data, err := json.Marshal(map[string]any{
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
		"Properties": map[string]any{
			"AuthMethod": "ProofOfPossession",
			"Id":         "{" + uuid.New().String() + "}",
			"DeviceType": o.Config.DeviceType,
			"Version":    o.Config.Version,
			"ProofKey": map[string]any{
				"crv": "P-256",
				"alg": "ES256",
				"use": "sig",
				"kty": "EC",
				"x":   base64.RawURLEncoding.EncodeToString(padTo32Bytes(key.PublicKey.X)),
				"y":   base64.RawURLEncoding.EncodeToString(padTo32Bytes(key.PublicKey.Y)),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling device auth request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", "https://device.auth.xboxlive.com/device/authenticate", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("POST %v: %w", "https://device.auth.xboxlive.com/device/authenticate", err)
	}
	request.Header.Set("x-xbl-contract-version", "1")
	sign(request, data, key)

	resp, err := o.httpClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %w", "https://device.auth.xboxlive.com/device/authenticate", err)
	}

	updateServerTimeFromHeaders(resp.Header)

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("POST %v: %v", "https://device.auth.xboxlive.com/device/authenticate", resp.Status)
	}
	token = &deviceToken{proofKey: key}
	return token, json.NewDecoder(resp.Body).Decode(token)
}

// sign signs the request passed containing the body passed. It signs the request using the ECDSA private key
// passed. If the request has a 'ProofKey' field in the Properties field, that key must be passed here.
func sign(request *http.Request, body []byte, key *ecdsa.PrivateKey) {
	serverTimeMu.Lock()
	currentServerDate := serverTime
	serverTimeMu.Unlock()
	var currentTime int64
	if !currentServerDate.IsZero() {
		currentTime = windowsTimestamp(currentServerDate)
	} else { // Should never happen
		currentTime = windowsTimestamp(time.Now())
	}

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
	path := request.URL.Path
	if rq := request.URL.RawQuery; rq != "" {
		path += "?" + rq
	}
	hash.Write([]byte(path))
	hash.Write([]byte{0})

	// Authorization header if present, otherwise an empty string + 0 byte.
	hash.Write([]byte(request.Header.Get("Authorization")))
	hash.Write([]byte{0})

	// Body data (only up to a certain limit, but this limit is practically never reached) + 0 byte.
	hash.Write(body)
	hash.Write([]byte{0})

	// Sign the checksum produced, and combine the 'r' and 's' into a single signature.
	// Encode r and s as 32-byte, zero-padded big-endian values so the P-256 signature is always exactly 64 bytes long.
	r, s, _ := ecdsa.Sign(rand.Reader, key, hash.Sum(nil))
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])

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
func windowsTimestamp(t time.Time) int64 {
	return (t.Unix() + 11644473600) * 10000000
}

// padTo32Bytes converts a big.Int into a fixed 32-byte, zero-padded slice.
// This is used to ensure that the X and Y coordinates of the ECDSA public key are always 32 bytes long,
// because big.Int.Bytes() returns a minimal encoding which may sometimes be less than 32 bytes.
func padTo32Bytes(b *big.Int) []byte {
	out := make([]byte, 32)
	b.FillBytes(out)
	return out
}

// parseXboxError returns the message associated with an Xbox Live error code.
func parseXboxErrorCode(code string) string {
	switch code {
	case "2148916227":
		return "Your account was banned by Xbox for violating one or more Community Standards for Xbox and is unable to be used."
	case "2148916229":
		return "Your account is currently restricted and your guardian has not given you permission to play online. Login to https://account.microsoft.com/family/ and have your guardian change your permissions."
	case "2148916233":
		return "Your account currently does not have an Xbox profile. Please create one at https://signup.live.com/signup"
	case "2148916234":
		return "Your account has not accepted Xbox's Terms of Service. Please login and accept them."
	case "2148916235":
		return "Your account resides in a region that Xbox has not authorized use from. Xbox has blocked your attempt at logging in."
	case "2148916236":
		return "Your account requires proof of age. Please login to https://login.live.com/login.srf and provide proof of age."
	case "2148916237":
		return "Your account has reached its limit for playtime. Your account has been blocked from logging in."
	case "2148916238":
		return "The account date of birth is under 18 years and cannot proceed unless the account is added to a family by an adult."
	default:
		return fmt.Sprintf("unknown error code: %v", code)
	}
}
