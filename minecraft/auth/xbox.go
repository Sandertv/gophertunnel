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
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/df-mc/go-xsapi"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth/authclient"
	"golang.org/x/oauth2"
)

// XBLToken holds info on the authorization token used for authenticating with XBOX Live.
type XBLToken struct {
	AuthorizationToken AuthorizationToken
	// key is the private key used as 'ProofKey' for authentication.
	// It is used for signing requests in [XBLToken.SetAuthHeader].
	key *ecdsa.PrivateKey
}

type AuthorizationToken struct {
	DisplayClaims DisplayClaims
	IssueInstant  time.Time
	NotAfter      time.Time
	Token         string
}

type DisplayClaims struct {
	// UserInfo is the user information from the authorization token.
	// GamerTag and XUID are only populated on the "xboxlive.com" relying party.
	// The rest only return UserHash.
	UserInfo []UserInfo `json:"xui"`
}

type DeviceDisplayClaims struct {
	DeviceInfo DeviceInfo `json:"xdi"`
}

type DeviceInfo struct {
	DeviceID string `json:"did"`
	DCS      string `json:"dcs"`
}

// UserInfo is the user claims structure used by XSAPI and also used in XSTS token display claims.
type UserInfo = xsapi.DisplayClaims

// SetAuthHeader sets the 'Authorization' header used for Minecraft related endpoints that
// need an XBOX Live authenticated caller.
func (t XBLToken) SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", t.String())

	if t.key == nil {
		return
	}

	var body []byte
	ok := false
	if req.GetBody != nil {
		rc, err := req.GetBody()
		if err == nil && rc != nil {
			body, _ = io.ReadAll(rc)
			_ = rc.Close()
			ok = true
		}
	} else if b, ok2 := req.Body.(interface{ Bytes() []byte }); ok2 {
		body = b.Bytes()
		ok = true
	}

	if ok {
		if err := sign(req, body, t.key); err != nil {
			slog.Error("signing XBL token", "error", err)
			return
		}
	}
}

// String returns a string that may be used for the 'Authorization' header used for Minecraft
// related endpoints that need an XBOX Live authenticated caller.
func (t XBLToken) String() string {
	if len(t.AuthorizationToken.DisplayClaims.UserInfo) == 0 {
		panic("XBLToken.String: received empty display claims user info")
	}
	return fmt.Sprintf("XBL3.0 x=%s;%s", t.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash, t.AuthorizationToken.Token)
}

// DisplayClaims returns a [xsapi.DisplayClaims] from the token. It can be used by the XSAPI
// package to include display claims in requests that require them.
func (t XBLToken) DisplayClaims() xsapi.DisplayClaims {
	if len(t.AuthorizationToken.DisplayClaims.UserInfo) == 0 {
		panic("XBLToken.DisplayClaims: received empty display claims user info")
	}
	return t.AuthorizationToken.DisplayClaims.UserInfo[0]
}

// expirationDelta is the amount of time before the token expires that it is considered valid.
const expirationDelta = time.Minute

// Valid returns whether the XBLToken is valid.
func (t XBLToken) Valid() bool {
	return time.Now().Before(t.AuthorizationToken.NotAfter.Add(-expirationDelta))
}

// Config specifies the configuration for authenticating with Xbox Live and Microsoft services.
// This struct should remain immutable.
type Config struct {
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

func newDefaultXBLHTTPClient() *http.Client {
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok || baseTransport == nil {
		baseTransport = &http.Transport{}
	}
	transport := baseTransport.Clone()
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	transport.TLSClientConfig.Renegotiation = tls.RenegotiateOnceAsClient

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// defaultXBLHTTPClient is the default HTTP client used for requests made by Xbox Live auth helpers.
var defaultXBLHTTPClient = newDefaultXBLHTTPClient()

// xblHTTPClient returns the HTTP client used for requests made by Xbox Live auth helpers.
// The HTTP client is obtained from the context via ctx.Value(oauth2.HTTPClient).
// If not present, a default client is used.
func xblHTTPClient(ctx context.Context) *http.Client {
	if ctx != nil {
		if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok && c != nil {
			return c
		}
	}
	return defaultXBLHTTPClient
}

// contextKey is a type used for context key used to [context.WithValue].
type contextKey struct{}

// tokenCacheContextKey is the context key used for holding an XBLTokenCache in [context.Context].
var tokenCacheContextKey contextKey

// XBLTokenCache caches device tokens for requesting Xbox Live tokens.
// It may be created from [Config.NewTokenCache] and included to a
// [context.Context] for re-using the device token in [RequestXBLToken].
type XBLTokenCache struct {
	// config is the Config used to request device tokens.
	// It contains platform-specific values for logging in with different device types.
	config Config
	// device caches the device token requested by XBLTokenCache.
	device *deviceToken
	// xsts caches the most recent XSTS tokens (XBLToken) issued for relying parties.
	// The key is the relying party string.
	xsts map[string]*XBLToken
	// inflight tracks in-progress XSTS token requests per relying party to avoid duplicate network calls
	// when multiple goroutines request the same relying party at the same time.
	inflight map[string]*xstsInFlight
	// deviceInflight tracks an in-progress device token request to avoid concurrent
	// duplicate device-auth calls under contention.
	deviceInflight *deviceInFlight
	// mu guards device from concurrent access.
	mu sync.Mutex
}

type xstsInFlight struct {
	done chan struct{}
	tok  *XBLToken
	err  error
}

type deviceInFlight struct {
	done chan struct{}
	tok  *deviceToken
	err  error
}

// NewTokenCache returns an XBLTokenCache that can be used to re-use XBL tokens
// in [RequestXBLToken].
func (conf Config) NewTokenCache() *XBLTokenCache {
	return &XBLTokenCache{
		config:   conf,
		xsts:     make(map[string]*XBLToken),
		inflight: make(map[string]*xstsInFlight),
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
// proof key.
func (x *XBLTokenCache) deviceToken(ctx context.Context, conf Config) (*deviceToken, error) {
	x.mu.Lock()
	if x.device != nil && x.device.Valid() {
		d := x.device
		x.mu.Unlock()
		return d, nil
	}

	if in := x.deviceInflight; in != nil {
		done := in.done
		x.mu.Unlock()
		select {
		case <-done:
			return in.tok, in.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	in := &deviceInFlight{done: make(chan struct{})}
	x.deviceInflight = in
	x.mu.Unlock()

	finish := func(tok *deviceToken, err error) (*deviceToken, error) {
		x.mu.Lock()
		x.deviceInflight = nil
		if err == nil && tok != nil {
			x.device = tok
		}
		in.tok = tok
		in.err = err
		close(in.done)
		x.mu.Unlock()
		return tok, err
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return finish(nil, fmt.Errorf("generate proof key: %w", err))
	}
	d, err := conf.obtainDeviceToken(ctx, key)
	if err != nil {
		return finish(nil, fmt.Errorf("obtain device token: %w", err))
	}
	return finish(d, nil)
}

var (
	// AndroidConfig is the configuration used in Minecraft: Bedrock Edition for Android devices.
	AndroidConfig = Config{
		DeviceType: "Android",
		ClientID:   "0000000048183522",
		Version:    "8.0.0",
		UserAgent:  "XAL Android 2020.07.20200714.000",
	}
	// IOSConfig is the configuration used in Minecraft: Bedrock Edition for iOS devices.
	IOSConfig = Config{
		DeviceType: "iOS",
		ClientID:   "000000004c17c01a",
		Version:    "15.6.1",
		UserAgent:  "XAL iOS 2021.11.20211021.000",
	}
	// Win32Config is the configuration used in Minecraft: Bedrock Edition for Windows devices.
	// Please note that the actual GDK/UWP build of the game requests the device token in more different way.
	Win32Config = Config{
		DeviceType: "Win32",
		ClientID:   "0000000040159362",
		Version:    "10.0.25398.4909",
		UserAgent:  "XAL Win32 2021.11.20220411.002",
	}
	// NintendoConfig is the configuration used in Minecraft: Bedrock Edition for Nintendo Switch.
	NintendoConfig = Config{
		DeviceType: "Nintendo",
		ClientID:   "00000000441cc96b",
		Version:    "0.0.0",
		UserAgent:  "XAL",
	}
	// PlayStationConfig is the configuration used in Minecraft: Bedrock Edition for PlayStation devices.
	PlayStationConfig = Config{
		DeviceType: "Playstation",
		ClientID:   "000000004827c78e",
		Version:    "10.0.0",
		UserAgent:  "XAL",
	}
)

// RequestXBLToken requests an Xbox Live token using a default device config.
// If an [XBLTokenCache] is present in ctx (via [WithXBLTokenCache]), its Config is used instead.
func RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	if ctx != nil {
		if cache, _ := ctx.Value(tokenCacheContextKey).(*XBLTokenCache); cache != nil {
			return cache.config.RequestXBLToken(ctx, liveToken, relyingParty)
		}
	}
	return AndroidConfig.RequestXBLToken(ctx, liveToken, relyingParty)
}

// normalizeRelyingPartyKey normalizes a relying party string for use as a cache key.
func normalizeRelyingPartyKey(relyingParty string) string {
	return strings.TrimRight(relyingParty, "/")
}

func (x *XBLTokenCache) requestXBLToken(ctx context.Context, conf Config, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	rpKey := normalizeRelyingPartyKey(relyingParty)

	x.mu.Lock()
	if x.config != conf { // Warning: This should be changed if pointer fields are added to the Config.
		x.mu.Unlock()
		return nil, errors.New("xbl token cache config mismatch")
	}
	if tok := x.xsts[rpKey]; tok != nil && tok.Valid() {
		x.mu.Unlock()
		return tok, nil
	}
	// inflight tracks in-progress XSTS token requests per relying party to avoid duplicate network calls
	if in := x.inflight[rpKey]; in != nil {
		done := in.done
		x.mu.Unlock()
		select {
		case <-done:
			return in.tok, in.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	in := &xstsInFlight{done: make(chan struct{})}
	if x.inflight == nil {
		x.inflight = make(map[string]*xstsInFlight)
	}
	x.inflight[rpKey] = in
	x.mu.Unlock()

	finish := func(tok *XBLToken, err error) (*XBLToken, error) {
		x.mu.Lock()
		delete(x.inflight, rpKey)
		if err == nil && tok != nil {
			if x.xsts == nil {
				x.xsts = make(map[string]*XBLToken)
			}
			x.xsts[rpKey] = tok
		}
		in.tok = tok
		in.err = err
		close(in.done)
		x.mu.Unlock()
		return tok, err
	}

	d, err := conf.getDeviceToken(ctx)
	if err != nil {
		return finish(nil, fmt.Errorf("request device token: %w", err))
	}
	xbl, err := conf.obtainXBLToken(ctx, liveToken, d, relyingParty)
	if err != nil {
		return finish(nil, err)
	}
	return finish(xbl, nil)
}

// RequestXBLToken requests an XBOX Live auth token using the passed Live token pair.
func (conf Config) RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if !liveToken.Valid() {
		return nil, fmt.Errorf("live token is no longer valid")
	}

	cache, _ := ctx.Value(tokenCacheContextKey).(*XBLTokenCache)
	if cache != nil {
		return cache.requestXBLToken(ctx, conf, liveToken, relyingParty)
	}

	d, err := conf.getDeviceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request device token: %w", err)
	}
	xbl, err := conf.obtainXBLToken(ctx, liveToken, d, relyingParty)
	if err != nil {
		return nil, err
	}

	return xbl, nil
}

// getDeviceToken attempts to use the cache from [context.Context], otherwise it will request
// a new device token using a new proof key.
func (conf Config) getDeviceToken(ctx context.Context) (*deviceToken, error) {
	if cache, ok := ctx.Value(tokenCacheContextKey).(*XBLTokenCache); ok && cache != nil {
		// If the context has a value with XBLTokenCache, we re-use them.
		return cache.deviceToken(ctx, conf)
	}
	// We first generate an ECDSA private key which will be used to provide a 'ProofKey' to each of the
	// requests, and to sign these requests.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate proof key: %w", err)
	}
	d, err := conf.obtainDeviceToken(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("obtain device token: %w", err)
	}
	return d, nil
}

func (conf Config) obtainXBLToken(ctx context.Context, liveToken *oauth2.Token, device *deviceToken, relyingParty string) (*XBLToken, error) {
	const sisuAuthUrl = "https://sisu.xboxlive.com/authorize"
	data, err := json.Marshal(map[string]any{
		"AccessToken":       "t=" + liveToken.AccessToken,
		"AppId":             conf.ClientID,
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

	req, err := http.NewRequestWithContext(ctx, "POST", sisuAuthUrl, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("POST %v: %w", sisuAuthUrl, err)
	}
	req.Header.Set("x-xbl-contract-version", "1")
	if err := sign(req, data, device.proofKey); err != nil {
		return nil, fmt.Errorf("signing XBL auth request: %w", err)
	}

	resp, err := authclient.SendRequestWithRetries(ctx, xblHTTPClient(ctx), req, authclient.RetryOptions{Attempts: 5})
	if err != nil {
		var body []byte
		if resp != nil && resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
		}
		return nil, newXboxNetworkError("POST", sisuAuthUrl, err, body)
	}
	defer resp.Body.Close()

	updateServerTimeFromHeaders(resp.Header)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		xboxErr := newXboxHTTPError("POST", sisuAuthUrl, resp, body)
		if accountErr, err := newAccountCreationRequiredError(xboxErr, resp.Header, body, device); err == nil {
			return nil, accountErr
		}
		return nil, xboxErr
	}
	info := new(XBLToken)
	if err := json.NewDecoder(resp.Body).Decode(info); err != nil {
		return nil, fmt.Errorf("decode XBL token: %w", err)
	}
	info.key = device.proofKey
	return info, nil
}

// deviceToken is the token obtained by requesting a device token by posting to xblDeviceAuthURL. Its Token
// field may be used in a request to obtain the XSTS token.
type deviceToken struct {
	DisplayClaims DeviceDisplayClaims `json:"DisplayClaims"`
	IssueInstant  time.Time           `json:"IssueInstant"`
	NotAfter      time.Time           `json:"NotAfter"`
	Token         string

	// proofKey is the private key used to sign requests in Xbox Live.
	proofKey *ecdsa.PrivateKey
}

// Valid returns whether the device token is valid.
func (d *deviceToken) Valid() bool {
	return time.Now().Before(d.NotAfter.Add(-expirationDelta))
}

// obtainDeviceToken sends a POST request to the device auth endpoint using the ECDSA private key passed to
// sign the request.
func (conf Config) obtainDeviceToken(ctx context.Context, key *ecdsa.PrivateKey) (token *deviceToken, err error) {
	properties := map[string]any{
		"AuthMethod": "ProofOfPossession",
		"Id":         "{" + uuid.New().String() + "}",
		"DeviceType": conf.DeviceType,
		"Version":    conf.Version,
		"ProofKey": map[string]any{
			"crv": "P-256",
			"alg": "ES256",
			"use": "sig",
			"kty": "EC",
			"x":   base64.RawURLEncoding.EncodeToString(padTo32Bytes(key.PublicKey.X)),
			"y":   base64.RawURLEncoding.EncodeToString(padTo32Bytes(key.PublicKey.Y)),
		},
	}

	switch conf.DeviceType {
	case AndroidConfig.DeviceType, NintendoConfig.DeviceType:
		properties["Id"] = "{" + uuid.NewString() + "}"
	case IOSConfig.DeviceType:
		properties["Id"] = strings.ToUpper(uuid.NewString())
	case PlayStationConfig.DeviceType:
		properties["Id"] = uuid.NewString()
	case Win32Config.DeviceType, "Xbox":
		properties["Id"] = "{" + strings.ToUpper(uuid.NewString()) + "}"
		properties["SerialNumber"] = properties["Id"]
	default:
		return nil, fmt.Errorf("unknown device type: %s", conf.DeviceType)
	}

	data, err := json.Marshal(map[string]any{
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
		"Properties":   properties,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling device auth request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", "https://device.auth.xboxlive.com/device/authenticate", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("POST %v: %w", "https://device.auth.xboxlive.com/device/authenticate", err)
	}

	request.Header.Set("Cache-Control", "no-store, must-revalidate, no-cache")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("x-xbl-contract-version", "1")
	request.Header.Set("Accept-Encoding", "gzip, deflate, compress")
	request.Header.Set("Accept-Language", "en-US, en;q=0.9")
	if err := sign(request, data, key); err != nil {
		return nil, fmt.Errorf("signing device auth request: %w", err)
	}

	resp, err := authclient.SendRequestWithRetries(ctx, xblHTTPClient(ctx), request, authclient.RetryOptions{Attempts: 5})
	if err != nil {
		var body []byte
		if resp != nil && resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
		}
		return nil, newXboxNetworkError("POST", "https://device.auth.xboxlive.com/device/authenticate", err, body)
	}

	updateServerTimeFromHeaders(resp.Header)

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, newXboxHTTPError("POST", "https://device.auth.xboxlive.com/device/authenticate", resp, body)
	}
	token = &deviceToken{proofKey: key}
	return token, json.NewDecoder(resp.Body).Decode(token)
}

// sign signs the request passed containing the body passed. It signs the request using the ECDSA private key
// passed. If the request has a 'ProofKey' field in the Properties field, that key must be passed here.
func sign(request *http.Request, body []byte, key *ecdsa.PrivateKey) error {
	serverTimeMu.Lock()
	delta := serverTimeDelta
	serverTimeMu.Unlock()
	var currentTime int64
	if delta != 0 {
		currentTime = windowsTimestamp(time.Now().Add(delta))
	} else {
		currentTime = windowsTimestamp(time.Now())
	}

	hash := sha256.New()

	// Signature policy version (0, 0, 0, 1) + 0 byte.
	buf := bytes.NewBuffer([]byte{0, 0, 0, 1, 0})
	// Timestamp + 0 byte.
	if err := binary.Write(buf, binary.BigEndian, currentTime); err != nil {
		return fmt.Errorf("writing current time: %w", err)
	}
	buf.Write([]byte{0})
	hash.Write(buf.Bytes())

	// HTTP method, generally POST + 0 byte.
	hash.Write([]byte(request.Method))
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
	r, s, err := ecdsa.Sign(rand.Reader, key, hash.Sum(nil))
	if err != nil {
		return fmt.Errorf("signing hash: %w", err)
	}
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])

	// The signature begins with 12 bytes, the first being the signature policy version (0, 0, 0, 1) again,
	// and the other 8 the timestamp again.
	buf = bytes.NewBuffer([]byte{0, 0, 0, 1})
	if err := binary.Write(buf, binary.BigEndian, currentTime); err != nil {
		return fmt.Errorf("writing current time: %w", err)
	}

	// Append the signature to the other 12 bytes, and encode the signature with standard base64 encoding.
	sig := append(buf.Bytes(), signature...)
	request.Header.Set("Signature", base64.StdEncoding.EncodeToString(sig))
	return nil
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
