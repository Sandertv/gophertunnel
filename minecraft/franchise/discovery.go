package franchise

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/franchise/internal"
	"net/http"
	"net/url"
)

// Discover obtains a Discovery for the specific version, which includes environments for various franchise services.
// It sends a GET request using [http.DefaultClient]. The version is typically [protocol.CurrentVersion] to ensure
// compatibility with the current version of the protocol package.
func Discover(build string) (*Discovery, error) {
	req, err := http.NewRequest(http.MethodGet, discoveryURL.JoinPath(build).String(), nil)
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
	var result internal.Result[*Discovery]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	if result.Data == nil {
		return nil, errors.New("minecraft/franchise: Discover: result.Data is nil")
	}
	return result.Data, nil
}

// Discovery provides access to environments for various franchise services based on the game
// version. It can be obtained from Discover using a specific game version.
//
// Example usage:
//
//	discovery, err := franchise.Discover(protocol.CurrentVersion)
//	if err != nil {
//		log.Fatalf("Error obtaining discovery: %s", err)
//	}
//
//	// Look up and decode an environment for authorization.
//	auth := new(franchise.AuthorizationEnvironment)
//	if err := discovery.Environment(auth, franchise.EnvironmentTypeProduction); err != nil {
//		log.Fatalf("Error reading environment for %q: %s", a.EnvironmentName(), err)
//	}
//
//	// Use discovery and auth for further use.
type Discovery struct {
	// ServiceEnvironments is a map where each key represents a service name. Each value is another map where keys are environment
	// types and values are environment-specific data represented as [json.RawMessage]. [Discovery.Environment] can be used to look
	// up and decode an Environment by its name and type.
	ServiceEnvironments map[string]map[string]json.RawMessage `json:"serviceEnvironments"`

	// SupportedEnvironments is a map where each key is the version of the game and
	// each value is a slice of supported environments types for that version.
	SupportedEnvironments map[string][]string `json:"supportedEnvironments"`
}

// Environment looks up for a value in [Discovery.ServiceEnvironments] with the name of the Environment and the environment type.
// If a value is found, which is a data represented in [json.RawMessage], it then decodes the data into the Environment. An error
// may be returned if the value cannot be found or during decoding the JSON data into the Environment.
func (d *Discovery) Environment(env Environment, typ string) error {
	e, ok := d.ServiceEnvironments[env.EnvironmentName()]
	if !ok {
		return errors.New("minecraft/franchise: environment not found")
	}
	data, ok := e[typ]
	if !ok {
		return errors.New("minecraft/franchise: environment with type not found")
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("decode environment: %w", err)
	}
	return nil
}

// Environment represents an environment for Discovery.
type Environment interface {
	// EnvironmentName returns the name of the environment.
	EnvironmentName() string
}

const (
	EnvironmentTypeProduction  = "prod"
	EnvironmentTypeDevelopment = "dev"
	EnvironmentTypeStaging     = "stage"
)

// discoveryURL is the base URL to make a GET request for obtaining a Discovery
// from Discover with a specific game version.
var discoveryURL = &url.URL{
	Scheme: "https",
	Host:   "client.discovery.minecraft-services.net",
	Path:   "/api/v1.0/discovery/MinecraftPE/builds/",
}
