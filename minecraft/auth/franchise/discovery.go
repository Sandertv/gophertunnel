package franchise

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sandertv/gophertunnel/minecraft/auth/franchise/internal"
)

const userAgent = "libhttpclient/1.0.0.0"

var discovered = map[string]*Discovery{}

func Discover(build string) (*Discovery, error) {
	if discovery, ok := discovered[build]; ok {
		return discovery, nil
	}
	req, err := http.NewRequest(http.MethodGet, discoveryURL.JoinPath(build).String(), nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

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
		return nil, errors.New("franchise: Discover: result.Data is nil")
	}
	discovered[build] = result.Data
	return result.Data, nil
}

type Discovery struct {
	ServiceEnvironments   map[string]map[string]json.RawMessage `json:"serviceEnvironments"`
	SupportedEnvironments map[string][]string                   `json:"supportedEnvironments"`
}

func (d *Discovery) Environment(env Environment, typ string) error {
	e, ok := d.ServiceEnvironments[env.EnvironmentName()]
	if !ok {
		return errors.New("franchise: environment not found")
	}
	data, ok := e[typ]
	if !ok {
		return errors.New("franchise: environment with type not found")
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("decode environment: %w", err)
	}
	return nil
}

type Environment interface {
	EnvironmentName() string
}

const (
	EnvironmentTypeProduction  = "prod"
	EnvironmentTypeDevelopment = "dev"
	EnvironmentTypeStaging     = "stage"
)

var discoveryURL = &url.URL{
	Scheme: "https",
	Host:   "client.discovery.minecraft-services.net",
	Path:   "/api/v1.0/discovery/MinecraftPE/builds/",
}
