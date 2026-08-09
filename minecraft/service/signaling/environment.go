package signaling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/internal"
)

// DefaultPingFrequency is used when the signaling service does not provide a
// positive ping interval.
const DefaultPingFrequency = time.Second * 15

// Environment represents an environment for the signaling service.
type Environment struct {
	// ServiceURI is the base endpoint URL for the signaling service.
	// The scheme is typically 'wss://'.
	ServiceURI *url.URL `json:"serviceUri"`
	// TurnURI is the 'turn://' URI for Microsoft's TURN server. NetherNet
	// connections use the TURN server embedded in
	// [github.com/df-mc/go-nethernet.Credentials]
	// for actual WebRTC negotiation, so the purpose of this field is unknown.
	TurnURI string `json:"turnUri"`
	// StunURI is the 'stun://' URI for Microsoft's STUN server. NetherNet
	// connections use the STUN server embedded in
	// [github.com/df-mc/go-nethernet.Credentials]
	// for actual WebRTC negotiation, so the purpose of this field is unknown.
	StunURI string `json:"stunUri"`
}

// ServiceName always returns 'signaling' as the name of the service environment.
// It implements [service.Environment] so it can be derived using [service.Discovery.Environment].
func (e *Environment) ServiceName() string {
	return "signaling"
}

// UnmarshalJSON decodes the ServiceURI field to string then parses as URL.
// Other URI fields such as TurnURI is not validated as it is not used in
// the actual WebRTC negotiation for NetherNet connections.
func (e *Environment) UnmarshalJSON(b []byte) (err error) {
	type Alias Environment
	data := struct {
		*Alias
		ServiceURI string `json:"serviceUri"`
	}{Alias: (*Alias)(e)}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	// [url.Parse] accepts empty strings and returns a valid url.URL with no error,
	// so we must explicitly validate that ServiceURI is not empty.
	if data.ServiceURI == "" {
		return errors.New("service/signaling: Environment.ServiceURI cannot be empty string")
	}
	e.ServiceURI, err = url.Parse(data.ServiceURI)
	if err != nil {
		return fmt.Errorf("parse service URI: %w", err)
	}
	return nil
}

// Configuration returns a Configuration using the environment's service URI.
// The discovery service has already resolved and determined the optimal server
// responsible for serving the signaling service. Because of that, this method
// does not involve any additional HTTP requests and returns a static configuration.
func (e *Environment) Configuration(context.Context, *http.Client, service.TokenSource) (*Configuration, error) {
	return &Configuration{
		ServiceURI:    e.ServiceURI,
		PingFrequency: DefaultPingFrequency,
	}, nil
}

// AFDEnvironment represents an environment for the signaling AFD (Azure Front Door) service.
// The signaling AFD service is seemingly hosted on Azure Front Door and used for resolving the
// URI of the signaling service nearest to the user. Currently, signaling AFD service seems to
// be only used for JSON-RPC connections.
type AFDEnvironment struct {
	internal.ServiceEnvironment
}

// ServiceName implements [service.Environment]. It always returns 'signaling-afd'.
func (e *AFDEnvironment) ServiceName() string {
	return "signaling-afd"
}

// Configuration dynamically resolves the configuration for the nearest signaling service.
// The returned Configuration is typically used by a Dialer to establish a WebSocket
// connection with the signaling service.
func (e *AFDEnvironment) Configuration(ctx context.Context, client *http.Client, src service.TokenSource) (*Configuration, error) {
	token, err := src.ServiceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request service token: %w", err)
	}

	requestURL := e.ServiceURI.JoinPath("/api/v1.0/configuration").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request environment: %w", err)
	}
	req.Header.Set("User-Agent", "libhttpclient/1.0.0.0")
	token.SetAuthHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var result internal.Result[*Configuration]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}
		if result.Data == nil {
			return nil, errors.New("service/signaling: AFDEnvironment: invalid configuration result")
		}
		return result.Data, nil
	default:
		return nil, internal.Err(resp)
	}
}

// ConfigurationProvider supplies configurations for establishing WebSocket
// connections to the signaling service. Both [Environment] and [AFDEnvironment]
// implement this interface.
type ConfigurationProvider interface {
	// Configuration resolves a configuration for establishing a long-lived WebSocket
	// connection with the signaling service. The given HTTP client must be used for invoking
	// HTTP requests. The token source may be used for endpoints that require authorization.
	Configuration(ctx context.Context, client *http.Client, src service.TokenSource) (*Configuration, error)
}

// Configuration represents the parameters for establishing a long-lived WebSocket connection
// with the signaling service. It is typically populated by a ConfigurationProvider (Environment
// or AFDEnvironment) and consumed by a Dialer to establish and maintain the connection.
type Configuration struct {
	// ServiceURI is the 'wss://' endpoint URI for the signaling service.
	// It may be locating to the nearest regional service to the caller.
	ServiceURI *url.URL `json:"signalingUri"`
	// PingFrequency is the interval for sending ping messages to keep
	// the WebSocket connection alive. The Dialer uses this value in a
	// goroutine to periodically send pings to the server.
	PingFrequency time.Duration `json:"pingFrequency"`
}

// UnmarshalJSON decodes the given JSON data into [Configuration].
// Some of the fields cannot be decoded directly into Go types so
// we first decode them as strings then manually parse them into appropriate types.
func (cfg *Configuration) UnmarshalJSON(b []byte) (err error) {
	type Alias Configuration
	data := struct {
		*Alias
		ServiceURI    string `json:"signalingUri"`
		PingFrequency string `json:"pingFrequency"`
	}{Alias: (*Alias)(cfg)}
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	}
	// [url.Parse] accepts empty strings and returns a valid url.URL with no error,
	// so we must explicitly validate that ServiceURI is not empty.
	if data.ServiceURI == "" {
		return errors.New("service/signaling: Configuration.ServiceURI cannot be empty string")
	}
	cfg.ServiceURI, err = url.Parse(data.ServiceURI)
	if err != nil {
		return fmt.Errorf("service/signaling: parse Configuration.ServiceURI: %w", err)
	}
	if data.PingFrequency == "" {
		cfg.PingFrequency = DefaultPingFrequency
		return nil
	}
	parts := strings.Split(data.PingFrequency, ":")
	if len(parts) != 3 {
		return fmt.Errorf("service/signaling: parse Configuration.PingFrequency: invalid value %q", data.PingFrequency)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("service/signaling: parse Configuration.PingFrequency hours: %w", err)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("service/signaling: parse Configuration.PingFrequency minutes: %w", err)
	}
	s, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("service/signaling: parse Configuration.PingFrequency seconds: %w", err)
	}
	if h < 0 || m < 0 || s < 0 {
		return fmt.Errorf("service/signaling: Configuration.PingFrequency cannot be negative: %q", data.PingFrequency)
	}
	cfg.PingFrequency = time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(s)*time.Second
	if cfg.PingFrequency <= 0 {
		return fmt.Errorf("service/signaling: Configuration.PingFrequency must be positive: %q", data.PingFrequency)
	}
	return nil
}
