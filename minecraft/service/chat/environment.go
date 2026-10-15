package chat

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/service/internal"
)

// Environment represents a service environment for communicating with the Chat service.
type Environment struct {
	internal.ServiceEnvironment

	// HTTPClient is the HTTP client used by Environment to make requests
	// to the endpoints provided by Chat service.
	// If nil, [http.DefaultClient] will be used instead.
	HTTPClient *http.Client `json:"-"`

	// keys is a map whose keys are the identifiers and values are the
	// public key bound to the ID. Each key must be validated using
	// verificationKey.validate before storing to this map, otherwise
	// verificationKey.hash would panic.
	keys map[string]*verificationKey
	// keysExpiry is the time when the cached key set expires.
	keysExpiry time.Time
	// keysMu guards keys from concurrent read/write access.
	keysMu sync.RWMutex
}

// DefaultEnvironment is the default [Environment] used when callers do not
// provide one explicitly. It may become outdated if Mojang changes the service endpoint.
var DefaultEnvironment = &Environment{
	ServiceEnvironment: internal.ServiceEnvironment{
		ServiceURI: &url.URL{
			Scheme: "https",
			Host:   "chat.franchise.minecraft-services.net",
		},
	},
}

// httpClient returns the HTTP client used for requests made by [Environment].
func (e *Environment) httpClient() *http.Client {
	if e.HTTPClient != nil {
		return e.HTTPClient
	}
	return http.DefaultClient
}
