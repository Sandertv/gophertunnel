package franchise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/text/language"
)

// Token represents an authorization token used for franchise services.
//
// A Token encapsulates the fields required for authenticating and authorizing requests.
//
// Token can be obtained through [TokenConfig.Token] or an implementation of IdentityProvider.
// As each Token has expiration, it is recommended to use an IdentityProvider to refresh the token
// subsequently when it becomes invalid.
type Token struct {
	// AuthorizationHeader is the JWT string that is used to authorize and authenticate requests.
	// It should be included in the 'Authorization' header of requests for services that requires
	// authentication. It can be set to a [http.Request] using the [Token.SetAuthHeader] method.
	AuthorizationHeader string `json:"authorizationHeader"`

	// ValidUntil specifies the expiration time of the Token. Once the current time surpasses the expiration
	// time, the Token is no longer valid, and it needs to be refreshed to maintain access.
	ValidUntil time.Time `json:"validUntil"`

	// Treatments is a list of treatments that have been applied to the Token. Treatments may be specific
	// to certain services and can be overridden using [DeviceConfig.TreatmentOverrides]. These treatments
	// could affect how the Token is validated or used in various services.
	Treatments []string `json:"treatments"`

	// Configurations is a map of configurations related to different services. It is unknown what it is used for.
	Configurations map[string]Configuration `json:"configurations"`

	// TreatmentContext provides additional context for the treatments applied to the Token. It is unknown how it is used.
	TreatmentContext string `json:"treatmentContext"`
}

// SetAuthHeader sets an 'Authorization' header to the [http.Request] using the [Token.AuthorizationHeader].
func (t *Token) SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", t.AuthorizationHeader)
}

// Token obtains a Token by making a POST request to the service URI specified in the Environment of TokenConfig.
//
// It creates the request URL by appending '/api/v1.0/session/start' to the base service URI defined in the
// [TokenConfig.Environment]. It then encodes the TokenConfig as a JSON string and sends it in the body of the request.
func (conf TokenConfig) Token() (*Token, error) {
	if conf.Environment == nil {
		return nil, errors.New("minecraft/franchise: TokenConfig: Environment is nil")
	}
	u, err := url.Parse(conf.Environment.ServiceURI)
	if err != nil {
		return nil, fmt.Errorf("parse service URI: %w", err)
	}
	u = u.JoinPath("/api/v1.0/session/start")

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(conf); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", internal.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", req.Method, req.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s: %s", req.Method, req.URL, resp.Status)
	}

	var result internal.Result[*Token]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	if result.Data == nil {
		return nil, errors.New("minecraft/franchise: TokenConfig: result.Data is nil")
	}
	return result.Data, nil
}

// Configuration represents a configuration set for the ID by service.
type Configuration struct {
	ID         string            `json:"id"`
	Parameters map[string]string `json:"parameters"`
}

// AuthorizationEnvironment represents an environment configuration used for authorization purposes.
// It holds essential fields required for accessing with authorization services and can be retrieved
// from a Discovery using the [Discovery.Environment] method.
type AuthorizationEnvironment struct {
	// ServiceURI is the URI of the service where requests related to authorization should be directed.
	// It is the base URL used for making authorization requests.
	ServiceURI string `json:"serviceUri"`
	Issuer     string `json:"issuer"`
	// PlayFabTitleID is the title ID specific for PlayFab.
	PlayFabTitleID    string `json:"playFabTitleId"`
	EduPlayFabTitleID string `json:"eduPlayFabTitleId"`
}

// EnvironmentName ...
func (*AuthorizationEnvironment) EnvironmentName() string { return "auth" }

// IdentityProvider implements a TokenConfig method, which provides a TokenConfig used for authorization.
//
// IdentityProvider is implemented by various platforms that support identity-based authentication and
// authorization. Platforms implementing IdentityProvider can provide a TokenConfig, which contains the
// necessary configuration to obtain Tokens required for accessing with franchise services.
type IdentityProvider interface {
	// TokenConfig should return a TokenConfig that includes the necessary configuration for authorizing with
	// franchise services. An error may be returned during authenticating with external platforms.
	TokenConfig() (*TokenConfig, error)
}

// TokenConfig defines the configuration required to obtain a Token through the [TokenConfig.Token] method for a
// specified source. It is essential to set the Environment field to ensure proper functionality.
type TokenConfig struct {
	// Device holds configuration details related to the device for which the Token is being obtained.
	// It can include device-specific settings that might be required by the authorization service.
	Device DeviceConfig `json:"device,omitempty"`

	// User contains user identity details encapsulated in a UserConfig. It is used to specify an identity token
	// authenticated with external platform, such as PlayFab.
	User UserConfig `json:"user,omitempty"`

	// Environment specifies the environment in which the TokenConfig will be used. It contains the service URI and
	// other environment-specific details necessary for creating the request URL to obtain a Token. Environment is
	// crucial for the Token retrieval and is not included in the request body.
	Environment *AuthorizationEnvironment `json:"-"`
}

// defaultDeviceConfig returns a default DeviceConfig based on the AuthorizationEnvironment.
// It is called by the [TokenConfig.Token] method to set a default device configuration when
// the [TokenConfig.Device] field is not set.
func defaultDeviceConfig(env *AuthorizationEnvironment) *DeviceConfig {
	return &DeviceConfig{
		ApplicationType: ApplicationTypeMinecraftPE,
		Capabilities:    []string{},
		GameVersion:     protocol.CurrentVersion,
		ID:              uuid.New(),
		Memory:          strconv.FormatUint(16*(1<<30), 10),
		Platform:        PlatformWindows10,
		PlayFabTitleID:  env.PlayFabTitleID,
		StorePlatform:   StorePlatformUWPStore,
		Type:            DeviceTypeWindows10,
	}
}

// DeviceConfig holds the details for the device used in authorization. It contains several fields that defines the
// features and capabilities of the device, which are associated with the Token being obtained.
type DeviceConfig struct {
	// ApplicationType indicates the type of the application associated with the device. It could be
	// one of constants defined below and is typically 'MinecraftPE' for most cases.
	ApplicationType string `json:"applicationType,omitempty"`

	// Capabilities is a list of features and functionalities supported by the device. Example might
	// include graphics capabilities like 'RayTracing' or other hardware or software features.
	Capabilities []string `json:"capabilities,omitempty"`

	// GameVersion indicates the version of the game running on the device. It is typically [protocol.CurrentVersion]
	// to ensure compatibility with the current version of the protocol package.
	GameVersion string `json:"gameVersion,omitempty"`

	// ID is a unique ID for the device, represented as a UUID.
	ID uuid.UUID `json:"id,omitempty"`

	// HardwareMemoryTier ranks the tier of hardware memory measured in the device. For example, 5 for 16GB.
	// It is unknown that how the value is determined for this field, and is seemingly not required for issuing a token.
	HardwareMemoryTier int `json:"hardwareMemoryTier,omitempty"`

	// Memory is the total amount of memory available on the device, represented as a string.
	Memory string `json:"memory,omitempty"`

	// Platform denotes the platform on which the device operates. It could be one of constants defined below, such as 'Windows10'.
	Platform string `json:"platform,omitempty"`

	// PlayFabTitleID is a unique ID for the PlayFab title associated with the device. It is used to reference a specific
	// PlayFab title and is typically set from [AuthorizationEnvironment.PlayFabTitleID] to ensure proper association with
	// correct title.
	PlayFabTitleID string `json:"playFabTitleId,omitempty"`

	// StorePlatform represents the digital store platform where the application can be downloaded or purchased from.
	// It could be one of constants defined below, such as 'UWPStore', that indicates the source of the application.
	StorePlatform string `json:"storePlatform,omitempty"`
	// TreatmentOverrides specifies any custom treatments that should be applied to the Token. These treatments may affect
	// how the Token is handled or used across various services available through Discovery.
	TreatmentOverrides []string `json:"treatmentOverrides,omitempty"`

	// Type defines the general category or type of the device. It could be one of constants defined below, such as 'Windows10'.
	Type string `json:"type,omitempty"`
}

// ApplicationTypeMinecraftPE defines that the application is Minecraft Bedrock Edition. It is unknown if other values are supported.
const ApplicationTypeMinecraftPE = "MinecraftPE"

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

// defaultUserConfig returns the default value for UserConfig.
func defaultUserConfig() *UserConfig {
	base, _ := language.AmericanEnglish.Base()
	region, _ := language.AmericanEnglish.Region()

	return &UserConfig{
		Language:     language.AmericanEnglish,
		LanguageCode: base.String(),
		RegionCode:   region.String(),
	}
}

// UserConfig represents the configuration details for a user whose Token is being obtained for authorization.
// It includes various fields related to the identity and language preferences of the user.
type UserConfig struct {
	// Language is the [language.Tag] representing the language the user is currently using.
	Language language.Tag `json:"language,omitempty"`

	// Language represents the base language derived from [language.Tag.Base].
	// It provides the primary language without any regional variations.
	LanguageCode string `json:"languageCode,omitempty"`

	// RegionCode denotes the specific region associated with the language.
	// It is typically derived from [language.Tag.Region] and provides regional context.
	RegionCode string `json:"regionCode,omitempty"`

	// Token is the identity token used for authorization. It is a crucial field that holds the actual
	// token required to authenticate the user.
	Token string `json:"token,omitempty"`

	// TokenType specifies the type or provider of the identity represented by the Token. It indicates the
	// source or type of authentication. It could be one of constants defined below, such as 'PlayFab' for
	// common token type.
	TokenType string `json:"tokenType,omitempty"`
}

// TokenTypePlayFab indicates that PlayFab is used for the underlying identity provider for the token.
const TokenTypePlayFab = "PlayFab"
