package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/df-mc/go-xsapi/v2/xal"
	"github.com/df-mc/go-xsapi/v2/xal/sisu"
	"github.com/df-mc/go-xsapi/v2/xal/xasd"
	"github.com/df-mc/go-xsapi/v2/xal/xsts"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// XBLToken holds info on the authorization token used for authenticating with XBOX Live.
type XBLToken struct {
	// AuthorizationToken is the XSTS token that relies on the specific relying party.
	// Some fields are only populated on the relying party "http://xboxlive.com".
	AuthorizationToken *xsts.Token
}

// SetAuthHeader sets the 'Authorization' header used for Minecraft related endpoints that
// need an XBOX Live authenticated caller.
func (t XBLToken) SetAuthHeader(r *http.Request) {
	if t.AuthorizationToken == nil || len(t.AuthorizationToken.DisplayClaims.UserInfo) == 0 {
		return
	}
	t.AuthorizationToken.SetAuthHeader(r)
}

// Valid returns whether the XBLToken is valid.
func (t XBLToken) Valid() bool {
	return t.AuthorizationToken != nil && t.AuthorizationToken.Valid()
}

// contextKey is a type used for context key used to [context.WithValue].
type contextKey struct{}

// tokenCacheContextKey is the context key used for holding an XBLTokenCache in [context.Context].
var tokenCacheContextKey contextKey

// XBLTokenCache caches device tokens for requesting Xbox Live tokens.
// It may be created from [Config.NewTokenCache] and included to a
// [context.Context] for re-using the device token in [RequestXBLToken].
type XBLTokenCache struct {
	// conf is the Config used to create this XBLTokenCache.
	conf Config
	// device is only present if the [XBLTokenCache] was created from [Config.NewTokenCache].
	device xasd.TokenSource
	// session is the SISU session cached by the [XBLTokenCache].
	session *sisu.Session
	// sessionMu guards session from concurrent read/write access.
	sessionMu sync.RWMutex
}

// Session returns a [sisu.Sesison] cached in [XBLTokenCache].
// Callers can save its snapshot via [sisu.Session.Snapshot]
// and restore it when creating a new session. The session can
// then be passed to [Config.ReuseTokenCache] for usage in [RequestXBLToken] again.
func (cache *XBLTokenCache) Session() *sisu.Session {
	cache.sessionMu.RLock()
	defer cache.sessionMu.RUnlock()
	return cache.session
}

// Device returns a [xasd.TokenSource] which supplies device tokens.
func (cache *XBLTokenCache) Device() xasd.TokenSource {
	if cache.device == nil {
		return cache.Session()
	}
	return cache.device
}

// ContextSession attempts to obtain [sisu.Session] from the given [context.Context].
// The [oauth2.TokenSource] is used to create a SISU session on the cache when needed.
// Callers can set their own session to the context by using [ReuseTokenCache].
func ContextSession(ctx context.Context, src oauth2.TokenSource) *sisu.Session {
	if cache, ok := ctx.Value(tokenCacheContextKey).(*XBLTokenCache); ok {
		cache.sessionMu.Lock()
		defer cache.sessionMu.Unlock()
		if cache.session == nil {
			cache.session = cache.conf.New(src, &sisu.SessionConfig{
				DeviceTokenSource: cache.device,
			})
		}
		return cache.session
	}
	return AndroidConfig.New(src, nil)
}

// NewTokenCache returns an XBLTokenCache that can be used to re-use XBL tokens
// in [RequestXBLToken].
func (conf Config) NewTokenCache() *XBLTokenCache {
	return &XBLTokenCache{
		conf:   conf,
		device: xasd.ReuseTokenSource(conf.Config.Config, nil, nil),
	}
}

// ReuseTokenCache returns an [XBLTokenCache] that uses the provided [sisu.Session]
// to request XBL tokens. Callers can embed the returned [XBLTokenCache] via [WithXBLTokenCache]
// for usage in [RequestXBLToken].
func (conf Config) ReuseTokenCache(session *sisu.Session) *XBLTokenCache {
	if session == nil {
		panic("auth: ReuseTokenCache: session cannot be nil")
	}
	return &XBLTokenCache{
		conf:    conf,
		session: session,
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

// Config encapsulates configuration for authenticating with Xbox Live services in a specific title.
type Config struct {
	// An embedded [sisu.Config] describes the SISU configuration used in the title.
	sisu.Config
}

var (
	// AndroidConfig is the configuration used in Minecraft: Bedrock Edition for Android devices.
	AndroidConfig = Config{
		sisu.Config{
			Config: xal.Config{
				// This indicates the device is running Android 13.
				Device: xal.Device{
					Type:    xal.DeviceTypeAndroid,
					Version: "13",
				},
				UserAgent: "XAL Android 2025.04.20250326.000",
				TitleID:   1739947436,
				Sandbox:   "RETAIL",
			},
			ClientID:    "0000000048183522",
			RedirectURI: "ms-xal-0000000048183522://auth",
		},
	}

	// IOSConfig is the configuration used in Minecraft: Bedrock Edition for iOS devices.
	IOSConfig = Config{
		sisu.Config{
			Config: xal.Config{
				Device: xal.Device{
					Type:    xal.DeviceTypeIOS,
					Version: "15.6.1",
				},
				UserAgent: "XAL iOS 2021.11.20211021.000",
				TitleID:   1810924247,
				Sandbox:   "RETAIL",
			},
			ClientID:    "000000004c17c01a",
			RedirectURI: "ms-xal-000000004c17c01a://auth",
		},
	}

	// Win32Config is the configuration for Minecraft: Bedrock Edition on Windows
	// devices. It is provided for reference only and does not support authentication,
	// as retrieving the RPS ticket required for device token requests is not yet known.
	Win32Config = Config{
		sisu.Config{
			Config: xal.Config{
				// Real devices obtain a device token using an RPS ticket retrieved from
				// Windows Live (login.live.com/RST2.srf). Retrieving the RPS ticket is
				// not yet known, so this configuration is not functional at this time.
				Device: xal.Device{
					Type:    xal.DeviceTypeWin32,
					Version: "10.0.28000", // NT version
				},
				UserAgent: "XAL GRTS 2025.11.20251105.000",
				TitleID:   896928775,
				Sandbox:   "RETAIL",
			},
			ClientID:    "0000000040159362",
			RedirectURI: "ms-xal-0000000040159362://auth",
		},
	}
	// NintendoConfig is the configuration used in Minecraft: Bedrock Edition for Nintendo Switch.
	NintendoConfig = Config{
		sisu.Config{
			Config: xal.Config{
				Device: xal.Device{
					Type:    xal.DeviceTypeNintendo,
					Version: "0.0.0",
				},
				UserAgent: "XAL",
				TitleID:   2047319603,
				Sandbox:   "RETAIL",
			},
			ClientID: "00000000441cc96b",
		},
	}
	// PlayStationConfig is the configuration used in Minecraft: Bedrock Edition for PlayStation devices.
	PlayStationConfig = Config{
		sisu.Config{
			Config: xal.Config{
				Device: xal.Device{
					Type:    xal.DeviceTypePlayStation,
					Version: "10.0.0",
				},
				UserAgent: "XAL",
				Sandbox:   "RETAIL",
				// TODO: Obtain TitleID from titlehub
			},
			ClientID: "000000004827c78e",
		},
	}

	// ServiceConfigID is the Service Configuration ID (SCID) used in release versions of Minecraft.
	// It is used for searching/hosting multiplayer sessions and querying achievements.
	ServiceConfigID = uuid.MustParse("4fc10100-5f7a-4470-899b-280835760c07")
	// PreviewServiceConfigID is the Service Configuration ID (SCID) used in preview versions of Minecraft.
	// It is used for searching/hosting multiplayer sessions and querying achievements in preview version.
	PreviewServiceConfigID = uuid.MustParse("00000000-0000-0000-0000-0000717d695f")
)

// RequestXBLToken requests an Xbox Live token using a default device config.
// If an [XBLTokenCache] is present in ctx (via [WithXBLTokenCache]), its Config is used instead.
func RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	var conf Config
	if cache, ok := ctx.Value(tokenCacheContextKey).(*XBLTokenCache); ok {
		conf = cache.conf
	} else {
		conf = AndroidConfig
	}
	return conf.RequestXBLToken(ctx, liveToken, relyingParty)
}

// RequestXBLToken requests an Xbox Live token using the OAuth2 token identifying the user's Microsoft Account.
// If an [XBLTokenCache] is present in ctx (via [WithXBLTokenCache]), it reuses or newly creates a SISU session inside the cache.
func (conf Config) RequestXBLToken(ctx context.Context, liveToken *oauth2.Token, relyingParty string) (*XBLToken, error) {
	var s *sisu.Session
	if cache, ok := ctx.Value(tokenCacheContextKey).(*XBLTokenCache); ok {
		if cache.conf != conf {
			return nil, errors.New("auth: Config.RequestXBLToken: config mismatch")
		}
		cache.sessionMu.Lock()
		if cache.session == nil {
			cache.session = conf.New(conf.TokenSource(context.WithoutCancel(ctx), liveToken), &sisu.SessionConfig{
				DeviceTokenSource: cache.device,
			})
		}
		s = cache.session
		cache.sessionMu.Unlock()
	} else {
		// If the cache storage does not exist, we request a new session every time
		// which may cause rate-limiting issues.
		s = conf.New(conf.TokenSource(context.WithoutCancel(ctx), liveToken), nil)
	}
	token, err := s.XSTSToken(ctx, relyingParty)
	if err != nil {
		return nil, err
	}
	return newXBLToken(token)
}

// newXBLToken wraps an XSTS token after validating the fields required for
// Minecraft/Xbox Authorization headers.
func newXBLToken(token *xsts.Token) (*XBLToken, error) {
	if token == nil {
		return nil, errors.New("auth: xsts token is nil")
	}
	if len(token.DisplayClaims.UserInfo) == 0 {
		return nil, errors.New("auth: xsts token has no user info")
	}
	if !token.Valid() {
		return nil, errors.New("auth: xsts token is invalid")
	}
	// Wrap the resulting token in XBLToken to maintain compatibility with the old code.
	return &XBLToken{AuthorizationToken: token}, nil
}
