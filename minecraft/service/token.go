package service

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/service/internal"
	"golang.org/x/text/language"
)

// TokenSource supplies a Token for authenticating with various network
// services in Minecraft: Bedrock Edition from underlying identity providers
// including PlayFab and Xbox Live.
//
// It is recommended to use [AuthorizationEnvironment.TokenSource].
type TokenSource interface {
	// Token returns a Token that is not expired and is determined
	// to be valid from [Token.Valid]. An error may be returned by
	// identity providers such as PlayFab or Xbox Live, rather than
	// by the authorization service.
	Token() (*Token, error)
}

// AuthorizationEnvironment represents an authorization environment for Minecraft-related services.
type AuthorizationEnvironment struct {
	// ServiceURI is the base URI of the authorization service where requests should be
	// directed. Methods implemented in [Environment] parses this URI as a [url.URL] then
	// appends a relative path for making an API call to the authorization service.
	ServiceURI *url.URL `json:"serviceUri"`
	// Issuer is the issuer used for OpenID Token Authentication.
	Issuer *url.URL `json:"issuer"`
	// PlayFabTitleID is the title ID specific for PlayFab for retail versions of the
	// game. It is typically '20CA2', and this is not something that could be easily
	// changed. By using PlayFabTitleID, the API host for PlayFab will be '<titleID.playfabapi.com>'.
	PlayFabTitleID string `json:"playFabTitleId"`
	// EduPlayFabTitleID is the title ID specific for PlayFab for Education Edition
	// of the game. It is used in educational versions to authenticate with a Student
	// account and to log in with some of the services for Education Edition.
	EduPlayFabTitleID string `json:"eduPlayFabTitleId"`

	// HTTPClient is the HTTP client used for requests made by AuthorizationEnvironment.
	// If nil, [http.DefaultClient] is used.
	HTTPClient *http.Client `json:"-"`

	// verifier verifies OpenID Multiplayer Token issued by the authorization service.
	// It is cached and kept by [Environment.Verifier] to reduce network time.
	verifier *oidc.IDTokenVerifier
	// verifierMu is a mutex that should be held when verifier is in access.
	verifierMu sync.Mutex
}

// httpClient returns the HTTP client used for requests made by AuthorizationEnvironment.
func (e *AuthorizationEnvironment) httpClient() *http.Client {
	if e.HTTPClient != nil {
		return e.HTTPClient
	}
	return http.DefaultClient
}

// ServiceName implements [service.Environment.ServiceName] and returns "auth".
func (e *AuthorizationEnvironment) ServiceName() string {
	return "auth"
}

// UnmarshalJSON implements [json.Unmarshaler.UnmarshalJSON].
// Since UnmarshalText() is not implemented in [url.URL] and [json.Unmarshal]
// attempts to decode the string URL as a struct, it will first decode the URL
// fields as strings then parse them manually.
// See: https://github.com/golang/go/issues/52638
func (e *AuthorizationEnvironment) UnmarshalJSON(b []byte) error {
	type Alias AuthorizationEnvironment
	data := struct {
		*Alias
		ServiceURI string `json:"serviceUri"`
		Issuer     string `json:"issuer"`
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	var err error
	e.ServiceURI, err = url.Parse(data.ServiceURI)
	if err != nil {
		return fmt.Errorf("parse ServiceURI: %w", err)
	}
	e.Issuer, err = url.Parse(data.Issuer)
	if err != nil {
		return fmt.Errorf("parse Issuer: %w", err)
	}
	return nil
}

// Token issues a [Token] through the authorization service using the TokenConfig.
// If TokenConfig is missing some values for the required fields, it will override
// them with default values.
func (e *AuthorizationEnvironment) Token(ctx context.Context, config TokenConfig) (*Token, error) {
	defaultUserConfig(&config.User)
	defaultDeviceConfig(e, &config.Device)
	if config.User.Token == "" {
		return nil, errors.New("minecraft/service: UserConfig.Token is empty")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(config); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	requestURL := e.ServiceURI.JoinPath("/api/v1.0/session/start").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, buf)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", internal.UserAgent)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, internal.Err(resp)
	}
	var result internal.Result[*Token]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	if result.Data == nil || !result.Data.Valid() {
		return nil, errors.New("minecraft/service: AuthorizationEnvironment: invalid token result")
	}
	return result.Data, nil
}

// Renew requests a refresh of a token that may soon expire. The user config must contain
// a valid PlayFab token that belong to the same user identity that was previously used
// for the Token. It is recommended to use TokenSource instead which subsequently renews the token.
func (e *AuthorizationEnvironment) Renew(ctx context.Context, token *Token, user UserConfig) (*Token, error) {
	defaultUserConfig(&user)
	if user.Token == "" {
		return nil, errors.New("minecraft/service: UserConfig.Token is empty")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(user); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	requestURL := e.ServiceURI.JoinPath("/api/v1.0/session/renew").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, buf)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	token.SetAuthHeader(req)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, internal.Err(resp)
	}
	var result internal.Result[*Token]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	if result.Data == nil || !result.Data.Valid() {
		return nil, errors.New("minecraft/service: invalid renew token result")
	}
	return result.Data, nil
}

// Verifier returns an [oidc.IDTokenVerifier] that can be used to verify the multiplayer
// token sent from clients in the Login packet to authenticate themselves with a remote
// OpenID configuration.
func (e *AuthorizationEnvironment) Verifier() (*oidc.IDTokenVerifier, error) {
	return e.VerifierContext(context.Background())
}

// VerifierContext returns an [oidc.IDTokenVerifier] that can be used to verify the multiplayer
// token sent from clients in the Login packet to authenticate themselves with a remote
// OpenID configuration.
func (e *AuthorizationEnvironment) VerifierContext(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	e.verifierMu.Lock()
	defer e.verifierMu.Unlock()
	if e.verifier != nil {
		return e.verifier, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.Issuer.JoinPath("/.well-known/openid-configuration").String(), nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", internal.UserAgent)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, internal.Err(resp)
	}
	var config oidc.ProviderConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	keys, err := e.publicKeys(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("obtain public keys: %w", err)
	}

	// We need to append '/' on the issuer if not present.
	issuer := e.Issuer.JoinPath().String()
	e.verifier = oidc.NewVerifier(issuer, &oidc.StaticKeySet{
		PublicKeys: keys,
	}, &oidc.Config{
		ClientID:             "api://auth-minecraft-services/multiplayer",
		SupportedSigningAlgs: config.Algorithms,
	})
	return e.verifier, nil
}

// publicKeys resolves the public keys from the JWKs URL of the [oidc.ProviderConfig].
// Those keys are used for verifying multiplayer tokens issued by the authorization service.
func (e *AuthorizationEnvironment) publicKeys(ctx context.Context, config oidc.ProviderConfig) ([]crypto.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.JWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", internal.UserAgent)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, internal.Err(resp)
	}
	keyset, err := decodeKeySet(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode key set: %w", err)
	}
	keys := make([]crypto.PublicKey, len(keyset.Keys))
	for i, key := range keyset.Keys {
		keys[i] = key.Key
	}
	return keys, nil
}

// decodeKeySet decodes the key set obtained from the authorization service
// with minor patches to support 'x5t' field being hex-encoded instead of
// [base64.RawURLEncoding].
func decodeKeySet(r io.Reader) (*jose.JSONWebKeySet, error) {
	var data struct {
		Keys []map[string]any `json:"keys"`
	}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}
	set := &jose.JSONWebKeySet{
		Keys: make([]jose.JSONWebKey, len(data.Keys)),
	}
	for i, key := range data.Keys {
		x5t, ok := key["x5t"].(string)
		if !ok {
			return nil, errors.New("no x5t found in jwk")
		}

		// Microsoft uses hex instead of base64 for the 'x5t' field, which violates the JOSE spec.
		// For a SHA-1 thumbprint (20 bytes), we expect either:
		// - Hex: 40 chars
		// - Base64URL (no padding): 27 chars
		var fingerprint []byte
		switch len(x5t) {
		case 40:
			var err error
			fingerprint, err = hex.DecodeString(x5t)
			if err != nil {
				return nil, fmt.Errorf("decode x5t hex: %w", err)
			}
			key["x5t"] = base64.RawURLEncoding.EncodeToString(fingerprint)
		case 27:
			var err error
			fingerprint, err = base64.RawURLEncoding.DecodeString(x5t)
			if err != nil {
				return nil, fmt.Errorf("decode x5t base64: %w", err)
			}
		default:
			return nil, fmt.Errorf("invalid x5t length: %d", len(x5t))
		}
		if n := len(fingerprint); n != 20 {
			return nil, fmt.Errorf("fingerprint is not 20 bytes long: %d", n)
		}

		// jose.JSONWebKey validates during JSON unmarshalling, so after patching x5t we re-encode the
		// JWK and let the custom UnmarshalJSON validate and decode it.
		b, err := json.Marshal(key)
		if err != nil {
			return nil, fmt.Errorf("encode reformatted jwk: %w", err)
		}
		var jwk jose.JSONWebKey
		if err := jwk.UnmarshalJSON(b); err != nil {
			return nil, fmt.Errorf("decode reformatted jwk: %w", err)
		}
		set.Keys[i] = jwk
	}
	return set, nil
}

// MultiplayerToken issues a token signed by the authorization service that can be used
// to authenticate with a multiplayer server. The public key can will be used as the 'cpk'
// claim of the token.
// Servers can verify this JWT using the remote OpenID configuration published by the
// authorization service and validate the claims to authenticate the player.
func (e *AuthorizationEnvironment) MultiplayerToken(ctx context.Context, src TokenSource, key *ecdsa.PublicKey) (string, error) {
	token, err := src.Token()
	if err != nil {
		return "", fmt.Errorf("request service token: %w", err)
	}
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", fmt.Errorf("encode public key: %w", err)
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(map[string]any{
		// This will be directly used as the 'cpk' claim in the token.
		"publicKey": base64.StdEncoding.EncodeToString(b),
	}); err != nil {
		return "", fmt.Errorf("encode request body: %w", err)
	}

	requestURL := e.ServiceURI.JoinPath("/api/v1.0/multiplayer/session/start").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, buf)
	if err != nil {
		return "", fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	token.SetAuthHeader(req)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", internal.Err(resp)
	}
	var result internal.Result[*multiplayerToken]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response body: %w", err)
	}
	if result.Data == nil || !result.Data.Valid() {
		return "", errors.New("minecraft/service: invalid multiplayer token result")
	}
	return result.Data.SignedToken, nil
}

// multiplayerToken encapsulates a JWT token issued by the authorization
// service with additional details, which can be used for authenticating
// with multiplayer servers using the OpenID configuration.
type multiplayerToken struct {
	// IssuedAt indicates the time the multiplayer token was issued.
	IssuedAt time.Time `json:"issuedAt"`
	// SignedToken is a JWT string that could be used by clients in the
	// connection request encapsulated in a Login packet. Servers are
	// expected to validate this JWT token with the OpenID configuration
	// of the authorization service.
	SignedToken string `json:"signedToken"`
	// ValidUntil is the expiration time for the multiplayer token.
	ValidUntil time.Time `json:"validUntil"`
}

// Valid returns whether multiplayerToken is valid.
func (t *multiplayerToken) Valid() bool {
	return t.SignedToken != "" && time.Now().Before(t.ValidUntil)
}

// defaultUserConfig sets default values for some of the fields that is
// not present on the UserConfig.
func defaultUserConfig(user *UserConfig) {
	if user.LanguageCode == language.Und {
		user.LanguageCode = language.AmericanEnglish
	}
	if user.Language == "" {
		base, _ := user.LanguageCode.Base()
		user.Language = base.String()
	}
	if user.RegionCode == "" {
		region, _ := user.LanguageCode.Region()
		user.RegionCode = region.String()
	}
}

// defaultDeviceConfig sets default values for some of the fields that
// is not presented on the DeviceConfig. An [AuthorizationEnvironment]
// is used to set the PlayFabID field.
func defaultDeviceConfig(auth *AuthorizationEnvironment, device *DeviceConfig) {
	if device.ApplicationType == "" {
		device.ApplicationType = ApplicationTypeMinecraftPE
	}
	if device.GameVersion == "" {
		device.GameVersion = protocol.CurrentVersion
	}
	if device.ID == "" {
		device.ID = uuid.NewString()
	}
	if device.Memory == "" {
		device.Memory = strconv.FormatUint(16*(1<<30), 10)
	}
	if device.Platform == "" {
		device.Platform = PlatformWindows10
	}
	if device.StorePlatform == "" {
		device.StorePlatform = StorePlatformUWPStore
	}
	if device.Type == "" {
		device.Type = DeviceTypeWindows10
	}
	if device.PlayFabTitleID == "" {
		device.PlayFabTitleID = auth.PlayFabTitleID
	}
}

// Token represents an authentication token used for various network
// services in Minecraft: Bedrock Edition.
//
// A Token may be issued using [AuthorizationEnvironment.Token] with a [TokenConfig].
// As each Token has expiration, it is recommended to use a [TokenSource]
// so it can be renewed subsequently when it becomes invalid.
type Token struct {
	// AuthorizationHeader is the JWT string that is used as the 'Authorization' header
	// to the requests ongoing to various network services for Minecraft: Bedrock Edition.
	//
	// [Token.SetAuthHeader] can be used to easily set an 'Authorization' header to this
	// value in a [http.Request].
	AuthorizationHeader string `json:"authorizationHeader"`

	// ValidUntil is the expiration time of the Token.
	// Once the current time surpasses the expiration time, the Token
	// is no longer valid, and needs to be either requested again or renewed
	// before the Token expires in a specific delta.
	ValidUntil time.Time `json:"validUntil"`

	// Treatments is a list of treatments that have been assigned to the Token.
	// Treatments may affect how the Token is validated or handled in various network
	// services in Minecraft: Bedrock Edition, and can be overridden using
	// [DeviceConfig.TreatmentOverrides].
	Treatments []string `json:"treatments"`

	// Configuration is a map of configuration specific to the application
	// the Token was authenticated in.
	Configuration map[string]Configuration `json:"configurations"`

	// TreatmentContext is a combination of Treatments separated by ';'.
	// It is unknown how it is used.
	TreatmentContext string `json:"treatmentContext"`
}

const expirationDelta = time.Minute

// Valid returns a bool indicating if the Token is valid.
func (t *Token) Valid() bool {
	return t.AuthorizationHeader != "" && time.Now().Before(t.ValidUntil.Add(-expirationDelta))
}

// SetAuthHeader sets an 'Authorization' header of the request to [Token.AuthorizationHeader].
func (t *Token) SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", t.AuthorizationHeader)
}

// Configuration represents a configuration set for the application.
type Configuration struct {
	// ID is the name of the Configuration.
	// It is typically 'Minecraft'.
	ID string `json:"id"`
	// Parameters contains data-driven parameters for the
	// components of the game client.
	//
	// It encapsulates every value in string even if the underlying type
	// is not string. For example the value true will be present as the
	// string 'true'.
	Parameters map[string]string `json:"parameters"`
}

// TokenConfig defines the configuration required to obtain a Token
// through the authorization service.
//
// It contains configurations for both devices and users associated with a Token.
type TokenConfig struct {
	// Device holds a configuration of the device for which the Token is being associated.
	// Device contains other non-user-specific fields including treatments or the game version.
	Device DeviceConfig `json:"device,omitempty"`

	// User contains user identity encapsulated in a UserConfig.
	// User contains an identity token authenticated with PlayFab via an external platform.
	User UserConfig `json:"user,omitempty"`
}

// UserConfig represents the configuration of the user whose Token
// is authenticated for. It includes various fields related to the
// identity and language preferences of the user.
type UserConfig struct {
	// Language is the base language of the user without region.
	// For example, if the game language was configured to
	// 'en-US', then Language will be 'en'.
	Language string `json:"language,omitempty"`

	// LanguageCode is the user language with the region.
	// It is typically same as what user has configured in the game.
	LanguageCode language.Tag `json:"languageCode,omitempty"`

	// RegionCode denotes the specific region associated with the language.
	// It is typically derived from [language.Tag.Region].
	// For example, if the game language was configured to
	// 'en-US', then RegionCode will be 'US'.
	RegionCode string `json:"regionCode,omitempty"`

	// Token is the identity token used for authentication.
	// It is a crucial field that holds the actual token required to
	// authenticate the user.
	//
	// If TokenType is set to 'PlayFab', then the Token will be the
	// session ticket taken from the initial login response from PlayFab,
	// which is obtained using either an Xbox Live token or custom account.
	Token string `json:"token,omitempty"`

	// TokenType specifies the provider of the Token.
	// The only known value for TokenType is 'PlayFab'.
	TokenType string `json:"tokenType,omitempty"`
}

// DeviceConfig describes the configuration of the device that is running the application.
// It contains several fields that defines the features and capabilities of the device,
// which are associated with the Token being obtained.
type DeviceConfig struct {
	// ApplicationType indicates the type of the application associated with the device. It could be
	// one of constants defined below and is typically 'MinecraftPE' for most cases.
	ApplicationType string `json:"applicationType,omitempty"`

	// Capabilities is a list of features and functionalities supported by the device. Example might
	// include graphics capabilities like 'RayTracing' or other hardware or software features.
	// It is represented as an empty JSON array if none of capabilities are specified.
	Capabilities []string `json:"capabilities"`

	// GameVersion indicates the version of the game running on the device. It is typically [protocol.CurrentVersion]
	// to ensure compatibility with the current version of the `protocol` package.
	GameVersion string `json:"gameVersion,omitempty"`

	// ID is a unique ID associated with the device. It is typically represented in UUID.
	// The simplest way to generate ID is via [github.com/google/uuid.NewString].
	ID string `json:"id,omitempty"`

	// HardwareMemoryTier ranks the tier of the hardware memory measured in the device.
	// It is unclear how this value is determined and how it is used.
	// For example, HardwareMemoryTier is 5 for 16GB devices.
	HardwareMemoryTier int `json:"hardwareMemoryTier,omitempty"`

	// Memory is the total amount of the memory available on the device,
	// represented as a numerical string. It defaults to 16GB if not present.
	Memory string `json:"memory,omitempty"`

	// Platform denotes the platform on which the device operates.
	// Platform could be one of the constants below, including 'Windows10'
	// for Windows 10/11 devices.
	Platform string `json:"platform,omitempty"`

	// PlayFabTitleID is the title ID specific to PlayFab.
	// It is typically '20CA2' for retail versions of the game.
	PlayFabTitleID string `json:"playFabTitleId,omitempty"`

	// StorePlatform represents the digital store platform where an in-app purchase is made.
	// StorePlatform could be one of the constants defined below, including 'UWPStore'
	// for Windows Store.
	StorePlatform string `json:"storePlatform,omitempty"`

	// TreatmentOverrides specifies any custom treatments that should be assigned to the Token.
	// These treatments may affect how the Token is handled or used across various services.
	//
	// For example, the signaling service may reject delivering signals to remote NetherNet network
	// if 'mc-signaling-usewebsockets' is not assigned for the token via TreatmentOverrides.
	TreatmentOverrides []string `json:"treatmentOverrides,omitempty"`

	// Type defines the general type of the device.
	// Type could be one of the constants below, including 'Windows10'
	// for Windows 10/11 devices.
	Type string `json:"type,omitempty"`
}

const (
	// CapabilityRayTracing indicates that the device is capable for ray tracing.
	CapabilityRayTracing = "RayTracing"
	// CapabilityVibrantVisuals indicates that the device is capable for rendering with Vibrant Visuals mode.
	CapabilityVibrantVisuals = "VibrantVisuals"
)

// PlatformWindows10 is reported on both Windows 10/11 devices running either UWP or GDK installation of the game.
const PlatformWindows10 = "Windows10"

// StorePlatformUWPStore defines that Microsoft Store is the primary store platform for the device, and is reported
// on both Windows 10/11 devices running either UWP or GDK installation of the game.
const StorePlatformUWPStore = "uwp.store"

// DeviceTypeWindows10 is reported on both Windows 10/11 devices running either UWP or GDK installation of the game.
const DeviceTypeWindows10 = "Windows10"

// TokenTypePlayFab indicates that PlayFab is used for the underlying identity provider for the TokenConfig.
const TokenTypePlayFab = "PlayFab"
